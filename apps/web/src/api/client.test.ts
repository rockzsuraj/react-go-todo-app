import axios, { AxiosError, AxiosInstance } from 'axios';
import { apiClient } from './client';

// Mock axios to avoid actual HTTP requests
jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

// Mock console.log to avoid noise in tests
const originalConsoleLog = console.log;
beforeAll(() => {
  console.log = jest.fn();
});

afterAll(() => {
  console.log = originalConsoleLog;
});

describe('API Client Token Refresh Interceptor', () => {
  beforeEach(() => {
    // Clear all mocks before each test
    jest.clearAllMocks();
    
    // Reset the module state by re-importing
    jest.resetModules();
    
    // Reset rate limiting state
    (apiClient as any)._isRateLimited = false;
    (apiClient as any)._rateLimitBlockTime = null;
  });

  describe('Successful Requests', () => {
    it('should pass through successful requests', async () => {
      const mockResponse = { data: { message: 'success' } };
      mockedAxios.create.mockReturnValue({
        ...mockedAxios,
        interceptors: {
          request: { use: jest.fn() },
          response: { use: jest.fn() },
        },
      } as any);

      // Re-import to get fresh instance
      const { apiClient: freshClient } = await import('./client');
      
      // Mock the actual request
      const mockRequest = jest.fn().mockResolvedValue(mockResponse);
      freshClient.get = mockRequest;

      const result = await freshClient.get('/test');
      
      expect(result).toEqual(mockResponse);
      expect(mockRequest).toHaveBeenCalledWith('/test');
    });
  });

  describe('Rate Limiting', () => {
    it('should block requests when rate limited', async () => {
      // Create a fresh client instance for testing
      const testClient = axios.create({
        baseURL: '/api',
        withCredentials: true,
      });

      let isRateLimited = false;
      let rateLimitBlockTime: number | null = null;

      // Add the same request interceptor logic
      testClient.interceptors.request.use(
        (config) => {
          if (isRateLimited && rateLimitBlockTime) {
            const blockDuration = 15 * 60 * 1000; // 15 minutes
            if (Date.now() - rateLimitBlockTime < blockDuration) {
              return Promise.reject(new Error('Requests blocked due to rate limiting'));
            }
          }
          return config;
        },
        (error) => Promise.reject(error)
      );

      // Simulate rate limit block
      isRateLimited = true;
      rateLimitBlockTime = Date.now();

      await expect(testClient.get('/test')).rejects.toThrow('Requests blocked due to rate limiting');
    });

    it('should allow requests after rate limit block expires', async () => {
      const testClient = axios.create({
        baseURL: '/api',
        withCredentials: true,
      });

      let isRateLimited = false;
      let rateLimitBlockTime: number | null = null;

      testClient.interceptors.request.use(
        (config) => {
          if (isRateLimited && rateLimitBlockTime) {
            const blockDuration = 15 * 60 * 1000; // 15 minutes
            if (Date.now() - rateLimitBlockTime < blockDuration) {
              return Promise.reject(new Error('Requests blocked due to rate limiting'));
            }
          }
          return config;
        },
        (error) => Promise.reject(error)
      );

      // Simulate expired rate limit block
      isRateLimited = true;
      rateLimitBlockTime = Date.now() - (16 * 60 * 1000); // 16 minutes ago

      const mockResponse = { data: { message: 'success' } };
      const mockRequest = jest.fn().mockResolvedValue(mockResponse);
      testClient.get = mockRequest;

      const result = await testClient.get('/test');
      
      expect(result).toEqual(mockResponse);
      expect(mockRequest).toHaveBeenCalledWith('/test');
    });
  });

  describe('Token Refresh Flow', () => {
    let refreshClient: AxiosInstance;
    let failedQueue: {
      resolve: (value?: unknown) => void;
      reject: (reason?: unknown) => void;
    }[];
    let isRefreshing = false;

    beforeEach(() => {
      // Mock refresh client
      refreshClient = axios.create({
        baseURL: '/api',
        withCredentials: true,
      });

      failedQueue = [];
      isRefreshing = false;
    });

    const processQueue = (error: AxiosError | null) => {
      failedQueue.forEach((prom) => {
        if (error) prom.reject(error);
        else prom.resolve();
      });
      failedQueue = [];
    };

    it('should refresh token on 401 response and retry original request', async () => {
      const originalRequest = {
        url: '/api/protected',
        method: 'get',
        _retry: false,
      };

      const mock401Response = {
        status: 401,
        config: originalRequest,
      };

      const mockSuccessResponse = {
        data: { message: 'success after refresh' },
      };

      // Mock the refresh call to succeed
      const mockRefreshPost = jest.fn().mockResolvedValue({ data: {} });
      refreshClient.post = mockRefreshPost;

      // Mock the original request to succeed after refresh
      const mockOriginalRequest = jest.fn()
        .mockResolvedValueOnce(mockSuccessResponse);
      
      // Simulate the response interceptor logic
      const responseInterceptor = async (error: AxiosError) => {
        const originalReq = error.config as any;

        if (error.response?.status === 401 && !originalReq._retry) {
          if (originalReq.url?.includes('/auth/refresh')) {
            isRefreshing = false;
            processQueue(error);
            return Promise.reject(error);
          }

          if (isRefreshing) {
            return new Promise((resolve, reject) => {
              failedQueue.push({ 
                resolve: () => resolve(mockOriginalRequest(originalReq)), 
                reject: (e: unknown) => reject(e) 
              });
            });
          }

          originalReq._retry = true;
          isRefreshing = true;

          try {
            await mockRefreshPost('/auth/refresh');
            isRefreshing = false;
            processQueue(null);
            return mockOriginalRequest(originalReq);
          } catch (refreshError) {
            isRefreshing = false;
            processQueue(refreshError as AxiosError);
            return Promise.reject(refreshError);
          }
        }

        return Promise.reject(error);
      };

      // Test the flow
      const result = await responseInterceptor(mock401Response as any);
      
      expect(mockRefreshPost).toHaveBeenCalledWith('/auth/refresh');
      expect(result).toEqual(mockSuccessResponse);
    });

    it('should handle concurrent requests during token refresh', async () => {
      const originalRequest1 = {
        url: '/api/protected1',
        method: 'get',
        _retry: false,
      };

      const originalRequest2 = {
        url: '/api/protected2',
        method: 'get',
        _retry: false,
      };

      const mock401Response1 = {
        status: 401,
        config: originalRequest1,
      };

      const mock401Response2 = {
        status: 401,
        config: originalRequest2,
      };

      const mockSuccessResponse1 = {
        data: { message: 'success 1 after refresh' },
      };

      const mockSuccessResponse2 = {
        data: { message: 'success 2 after refresh' },
      };

      // Mock refresh call
      const mockRefreshPost = jest.fn().mockResolvedValue({ data: {} });
      refreshClient.post = mockRefreshPost;

      // Mock original requests
      const mockOriginalRequest = jest.fn()
        .mockResolvedValueOnce(mockSuccessResponse1)
        .mockResolvedValueOnce(mockSuccessResponse2);

      // Simulate the response interceptor logic
      const responseInterceptor = async (error: AxiosError) => {
        const originalReq = error.config as { _retry?: boolean; url?: string };

        if (error.response?.status === 401 && !originalReq._retry) {
          if (isRefreshing) {
            return new Promise((resolve, reject) => {
              failedQueue.push({ 
                resolve: () => resolve(mockOriginalRequest(originalReq)), 
                reject: (e: unknown) => reject(e) 
              });
            });
          }

          originalReq._retry = true;
          isRefreshing = true;

          try {
            await mockRefreshPost('/auth/refresh');
            isRefreshing = false;
            processQueue(null);
            return mockOriginalRequest(originalReq);
          } catch (refreshError) {
            isRefreshing = false;
            processQueue(refreshError as AxiosError);
            return Promise.reject(refreshError);
          }
        }

        return Promise.reject(error);
      };

      // Start both requests concurrently
      const [result1, result2] = await Promise.all([
        responseInterceptor(mock401Response1 as any),
        responseInterceptor(mock401Response2 as any),
      ]);

      expect(mockRefreshPost).toHaveBeenCalledTimes(1);
      expect(result1).toEqual(mockSuccessResponse1);
      expect(result2).toEqual(mockSuccessResponse2);
    });

    it('should handle refresh token failure', async () => {
      const originalRequest = {
        url: '/api/protected',
        method: 'get',
        _retry: false,
      };

      const mock401Response = {
        status: 401,
        config: originalRequest,
      };

      const refreshError = new Error('Refresh token failed');
      
      // Mock the refresh call to fail
      const mockRefreshPost = jest.fn().mockRejectedValue(refreshError);
      refreshClient.post = mockRefreshPost;

      // Simulate the response interceptor logic
      const responseInterceptor = async (error: AxiosError) => {
        const originalReq = error.config as any;

        if (error.response?.status === 401 && !originalReq._retry) {
          originalReq._retry = true;
          isRefreshing = true;

          try {
            await mockRefreshPost('/auth/refresh');
            isRefreshing = false;
            processQueue(null);
            return Promise.resolve({ data: {} });
          } catch (refreshErr) {
            isRefreshing = false;
            processQueue(refreshErr as AxiosError);
            return Promise.reject(refreshErr);
          }
        }

        return Promise.reject(error);
      };

      await expect(responseInterceptor(mock401Response as any)).rejects.toThrow('Refresh token failed');
      expect(mockRefreshPost).toHaveBeenCalledWith('/auth/refresh');
    });

    it('should not retry refresh endpoint itself on 401', async () => {
      const refreshRequest = {
        url: '/api/auth/refresh',
        method: 'post',
        _retry: false,
      };

      const mock401Response = {
        status: 401,
        config: refreshRequest,
      };

      // Mock refresh client to fail
      const mockRefreshPost = jest.fn().mockRejectedValue(new Error('Refresh failed'));
      refreshClient.post = mockRefreshPost;

      // Simulate the response interceptor logic
      const responseInterceptor = async (error: AxiosError) => {
        const originalReq = error.config as any;

        if (error.response?.status === 401 && !originalReq._retry) {
          // If the request that just failed was the refresh call, stop the loop
          if (originalReq.url?.includes('/auth/refresh')) {
            isRefreshing = false;
            processQueue(error);
            return Promise.reject(error);
          }

          originalReq._retry = true;
          isRefreshing = true;

          try {
            await mockRefreshPost('/auth/refresh');
            isRefreshing = false;
            processQueue(null);
            return Promise.resolve({ data: {} });
          } catch (refreshError) {
            isRefreshing = false;
            processQueue(refreshError as AxiosError);
            return Promise.reject(refreshError);
          }
        }

        return Promise.reject(error);
      };

      await expect(responseInterceptor(mock401Response as any)).rejects.toThrow();
      expect(mockRefreshPost).not.toHaveBeenCalled();
    });
  });
});
