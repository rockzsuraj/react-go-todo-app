import axios, {
  type AxiosInstance,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from 'axios';

interface FailedRequest {
  resolve: (value: AxiosResponse) => void;
  reject: (reason?: unknown) => void;
}

interface RetryableRequestConfig extends InternalAxiosRequestConfig {
  _retry?: boolean;
  _renderWakeRetries?: number;
}

export const apiClient: AxiosInstance = axios.create({
  baseURL: '/api',
  withCredentials: true,
});

// Separate client for the refresh call so it never triggers the interceptor recursively.
const refreshClient = axios.create({
  baseURL: '/api',
  withCredentials: true,
});

const renderWakeRetryDelays = [2_000, 5_000, 10_000, 20_000];

export const isRenderHibernateRateLimit = (
  status: number | undefined,
  routing: unknown,
): boolean => status === 429 && routing === 'hibernate-rate-limited';

let isRefreshing = false;
let failedQueue: FailedRequest[] = [];

const processQueue = (error: unknown): void => {
  const queue = failedQueue;
  failedQueue = [];
  for (const prom of queue) {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(null as unknown as AxiosResponse);
    }
  }
};

apiClient.interceptors.response.use(
  (response) => response,
  async (err: unknown) => {
    if (!axios.isAxiosError(err)) {
      return Promise.reject(err);
    }

    const originalRequest = err.config as RetryableRequestConfig;

    // Render free web services can return this response while waking from
    // hibernation. Retry only this platform-generated 429.
    const renderRouting = err.response?.headers?.['x-render-routing'];
    if (isRenderHibernateRateLimit(err.response?.status, renderRouting)) {
      const retryCount = originalRequest._renderWakeRetries ?? 0;
      if (retryCount < renderWakeRetryDelays.length) {
        originalRequest._renderWakeRetries = retryCount + 1;
        await new Promise((resolve) =>
          window.setTimeout(resolve, renderWakeRetryDelays[retryCount]),
        );
        return apiClient(originalRequest);
      }
    }

    // Pass through non-401s immediately.
    if (err.response?.status !== 401 || originalRequest._retry) {
      return Promise.reject(err);
    }

    // Never attempt refresh for auth-related endpoints.
    // This naturally prevents the logout → 401 → refresh → 401 loop.
    if (originalRequest.url?.includes('/auth/')) {
      return Promise.reject(err);
    }

    // Queue subsequent requests while a refresh is in-flight.
    if (isRefreshing) {
      return new Promise<AxiosResponse>((resolve, reject) => {
        failedQueue.push({
          resolve: () => resolve(apiClient(originalRequest)),
          reject: (e) => reject(e),
        });
      });
    }

    originalRequest._retry = true;
    isRefreshing = true;

    try {
      await refreshClient.post('/auth/refresh');
      processQueue(null);
      return apiClient(originalRequest);
    } catch (refreshError) {
      processQueue(refreshError);
      // Notify the app that the session has fully expired so route guards can react.
      window.dispatchEvent(new CustomEvent('auth:session-expired'));
      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  },
);
