import axios, { AxiosError, InternalAxiosRequestConfig, AxiosInstance } from 'axios';

interface FailedRequest {
  resolve: (value?: unknown) => void;
  reject: (reason?: unknown) => void;
}

// API Error interface matching backend response
interface APIError {
  code: string;
  message: string;
  details?: string;
}

// Enhanced error type for better type safety
type EnhancedAxiosError = AxiosError & {
  response?: {
    data?: {
      error?: APIError;
    };
    status?: number;
    statusText?: string;
    headers?: Record<string, string>;
  };
};

export const apiClient: AxiosInstance = axios.create({
  baseURL: '/api',
  withCredentials: true,
});

// A separate client for refreshing ensures that refresh call 
// itself doesn't trigger this interceptor recursively.
const refreshClient = axios.create({
  baseURL: '/api',
  withCredentials: true,
});

let isRefreshing = false;
let failedQueue: FailedRequest[] = [];

// Global rate limit block
let isRateLimited = false;
let rateLimitBlockTime: number | null = null;
const RATE_LIMIT_BLOCK_DURATION = 15 * 60 * 1000; // 15 minutes

// Check if requests should be blocked due to rate limiting
const shouldBlockRequests = (): boolean => {
  if (isRateLimited && rateLimitBlockTime) {
    return Date.now() - rateLimitBlockTime < RATE_LIMIT_BLOCK_DURATION;
  }
  return false;
};

// Set rate limit block
const setRateLimitBlock = (): void => {
  isRateLimited = true;
  rateLimitBlockTime = Date.now();
  console.log('Rate limit detected - blocking all requests for 15 minutes');
};

const processQueue = (error: EnhancedAxiosError | null): void => {
  failedQueue.forEach((prom) => {
    if (error) prom.reject(error);
    else prom.resolve();
  });
  failedQueue = [];
};

// Type guard to check if error is our enhanced AxiosError
const isEnhancedAxiosError = (error: unknown): error is EnhancedAxiosError => {
  return error !== null && 
         typeof error === 'object' && 
         'response' in error &&
         error instanceof Error &&
         'config' in error;
};

// Request interceptor to block requests when rate limited
apiClient.interceptors.request.use(
  (config) => {
    // Block all requests if rate limited
    if (shouldBlockRequests()) {
      return Promise.reject(new Error('Requests blocked due to rate limiting'));
    }
    return config;
  },
  (error) => Promise.reject(error)
);

apiClient.interceptors.response.use(
  (response) => response,
  async (err: unknown) => {
    if (!isEnhancedAxiosError(err)) {
      return Promise.reject(err);
    }

    const originalRequest = err.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // Handle rate limiting (429) - block all future requests
    if (err.response?.status === 429) {
      setRateLimitBlock();
      return Promise.reject(err);
    }

    // Handle 401 Unauthorized
    if (err.response?.status === 401 && !originalRequest._retry) {
      
      // If request that just failed was refresh call, stop the loop
      if (originalRequest.url?.includes('/auth/refresh')) {
        isRefreshing = false;
        processQueue(err);
        return Promise.reject(err);
      }

      // If another request is already refreshing, queue this one
      if (isRefreshing) {
        return new Promise<unknown>((resolve, reject) => {
          failedQueue.push({ 
            resolve: () => resolve(apiClient(originalRequest)), 
            reject: (e) => reject(e) 
          });
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        // Attempt to rotate the token
        await refreshClient.post('/auth/refresh');
        
        isRefreshing = false;
        processQueue(null); 

        // Retry the original failing request
        return apiClient(originalRequest); 
      } catch (refreshError) {
        isRefreshing = false;
        processQueue(refreshError as EnhancedAxiosError);
        
        // If refresh call fails, reject the original request
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(err);
  }
);