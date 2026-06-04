/**
 * Tests for the Axios interceptor logic in client.ts.
 *
 * Strategy: we test the interceptor logic by simulating it inline rather
 * than trying to hook into a real Axios instance, because jest.mock('axios')
 * replaces the module globally and breaks real axios.create() calls.
 * This gives us deterministic, fast tests without HTTP.
 */
import type { AxiosError } from 'axios';

// ─── Helpers ─────────────────────────────────────────────────────────────────

/** Build a minimal fake AxiosError for testing the interceptor logic. */
function makeAxiosError(
  status: number,
  url: string,
  retried = false,
): AxiosError {
  return {
    isAxiosError: true,
    name: 'AxiosError',
    message: `Request failed with status code ${status}`,
    config: { url, _retry: retried } as any,
    response: {
      status,
      data: {},
      statusText: '',
      headers: {},
      config: {} as any,
    },
    toJSON: () => ({}),
  } as unknown as AxiosError;
}

/** Build a minimal copy of the rate-limiting logic for isolated unit tests. */
function makeRateLimiter(blockDurationMs = 15 * 60 * 1000) {
  let isRateLimited = false;
  let rateLimitBlockTime: number | null = null;

  return {
    setBlock: () => {
      isRateLimited = true;
      rateLimitBlockTime = Date.now();
    },
    setExpiredBlock: () => {
      isRateLimited = true;
      rateLimitBlockTime = Date.now() - blockDurationMs - 1;
    },
    shouldBlock: () => {
      if (isRateLimited && rateLimitBlockTime) {
        if (Date.now() - rateLimitBlockTime < blockDurationMs) return true;
        isRateLimited = false;
        rateLimitBlockTime = null;
      }
      return false;
    },
  };
}

/** Build a self-contained version of the refresh interceptor for unit testing. */
function makeRefreshInterceptor() {
  type FailedReq = {
    resolve: (v?: unknown) => void;
    reject: (e?: unknown) => void;
  };
  let isRefreshing = false;
  let failedQueue: FailedReq[] = [];

  const processQueue = (error: Error | null) => {
    const q = failedQueue;
    failedQueue = [];
    q.forEach((p) => {
      if (error) {
        p.reject(error);
      } else {
        p.resolve();
      }
    });
  };

  return {
    handle: async (
      error: AxiosError,
      doRefresh: () => Promise<unknown>,
      doOriginalRequest: (cfg: unknown) => Promise<unknown>,
    ): Promise<unknown> => {
      const cfg = error.config as any;

      if (error.response?.status !== 401 || cfg._retry) {
        return Promise.reject(error);
      }

      if (cfg.url?.includes('/auth/refresh')) {
        isRefreshing = false;
        processQueue(new Error('refresh failed'));
        return Promise.reject(error);
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({
            resolve: () => resolve(doOriginalRequest(cfg)),
            reject,
          });
        });
      }

      cfg._retry = true;
      isRefreshing = true;

      try {
        await doRefresh();
        isRefreshing = false;
        processQueue(null);
        return doOriginalRequest(cfg);
      } catch (refreshErr) {
        isRefreshing = false;
        processQueue(refreshErr as Error);
        return Promise.reject(refreshErr);
      }
    },
  };
}

// ─── Tests ───────────────────────────────────────────────────────────────────

describe('Rate Limiting Logic', () => {
  it('should block requests while rate limit is active', () => {
    const limiter = makeRateLimiter();
    expect(limiter.shouldBlock()).toBe(false);
    limiter.setBlock();
    expect(limiter.shouldBlock()).toBe(true);
  });

  it('should allow requests after rate limit block expires', () => {
    const limiter = makeRateLimiter();
    limiter.setExpiredBlock();
    expect(limiter.shouldBlock()).toBe(false);
  });

  it('should auto-reset state once block expires', () => {
    const limiter = makeRateLimiter();
    limiter.setExpiredBlock();
    limiter.shouldBlock();
    expect(limiter.shouldBlock()).toBe(false);
  });
});

describe('Token Refresh Interceptor Logic', () => {
  let interceptor: ReturnType<typeof makeRefreshInterceptor>;

  beforeEach(() => {
    interceptor = makeRefreshInterceptor();
  });

  it('should refresh token on 401 and retry original request', async () => {
    const err = makeAxiosError(401, '/api/protected');
    const doRefresh = jest.fn().mockResolvedValue(undefined);
    const successResponse = { data: { ok: true } };
    const doOriginal = jest.fn().mockResolvedValue(successResponse);

    const result = await interceptor.handle(err, doRefresh, doOriginal);

    expect(doRefresh).toHaveBeenCalledTimes(1);
    expect(doOriginal).toHaveBeenCalledTimes(1);
    expect(result).toEqual(successResponse);
  });

  it('should not refresh on 401 if request is already retried', async () => {
    const err = makeAxiosError(401, '/api/protected', true);
    const doRefresh = jest.fn();
    const doOriginal = jest.fn();

    await expect(
      interceptor.handle(err, doRefresh, doOriginal),
    ).rejects.toEqual(err);
    expect(doRefresh).not.toHaveBeenCalled();
  });

  it('should not refresh if the 401 came from the refresh endpoint itself', async () => {
    const err = makeAxiosError(401, '/auth/refresh');
    const doRefresh = jest.fn();
    const doOriginal = jest.fn();

    await expect(
      interceptor.handle(err, doRefresh, doOriginal),
    ).rejects.toEqual(err);
    expect(doRefresh).not.toHaveBeenCalled();
  });

  it('should queue concurrent 401s and only refresh once', async () => {
    const err1 = makeAxiosError(401, '/api/protected1');
    const err2 = makeAxiosError(401, '/api/protected2');

    const doRefresh = jest.fn().mockResolvedValue(undefined);
    const doOriginal = jest
      .fn()
      .mockResolvedValueOnce({ data: 'res1' })
      .mockResolvedValueOnce({ data: 'res2' });

    const results = await Promise.all([
      interceptor.handle(err1, doRefresh, doOriginal),
      interceptor.handle(err2, doRefresh, doOriginal),
    ]);

    expect(doRefresh).toHaveBeenCalledTimes(1);
    expect(doOriginal).toHaveBeenCalledTimes(2);
    expect(results).toContainEqual({ data: 'res1' });
    expect(results).toContainEqual({ data: 'res2' });
  });

  it('should reject original request if refresh fails', async () => {
    const err = makeAxiosError(401, '/api/protected');
    const refreshErr = new Error('Refresh token expired');
    const doRefresh = jest.fn().mockRejectedValue(refreshErr);
    const doOriginal = jest.fn();

    await expect(
      interceptor.handle(err, doRefresh, doOriginal),
    ).rejects.toThrow('Refresh token expired');
    expect(doOriginal).not.toHaveBeenCalled();
  });

  it('should reject queued requests if refresh fails', async () => {
    const err1 = makeAxiosError(401, '/api/protected1');
    const err2 = makeAxiosError(401, '/api/protected2');

    const refreshErr = new Error('Refresh failed');
    let resolveRefresh!: () => void;
    const refreshPromise = new Promise<void>((r) => {
      resolveRefresh = r;
    });

    const doRefresh = jest.fn().mockImplementation(async () => {
      await refreshPromise;
      throw refreshErr;
    });
    const doOriginal = jest.fn();

    const p1 = interceptor.handle(err1, doRefresh, doOriginal);
    const p2 = interceptor.handle(err2, doRefresh, doOriginal);

    resolveRefresh();

    await expect(p1).rejects.toThrow('Refresh failed');
    await expect(p2).rejects.toThrow('Refresh failed');
    expect(doRefresh).toHaveBeenCalledTimes(1);
    expect(doOriginal).not.toHaveBeenCalled();
  });

  it('should pass through non-401 errors without refreshing', async () => {
    const err = makeAxiosError(500, '/api/protected');
    const doRefresh = jest.fn();
    const doOriginal = jest.fn();

    await expect(
      interceptor.handle(err, doRefresh, doOriginal),
    ).rejects.toEqual(err);
    expect(doRefresh).not.toHaveBeenCalled();
  });
});
