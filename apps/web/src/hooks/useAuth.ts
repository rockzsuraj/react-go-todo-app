import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { authApi } from '../api';
import { logger } from '../services/logger';
import { APIErrorHandler } from '../utils/errorHandler';
import { User } from '../types/user';

// Global state to completely prevent auth calls
let hasAuthFailed = false;
let authFailureTime: number | null = null;
const AUTH_BLOCK_DURATION = 30 * 1000; // 30 seconds (reduced from 10 minutes)
let isRateLimited = false;
let rateLimitTime: number | null = null;
const RATE_LIMIT_BLOCK_DURATION = 2 * 60 * 1000; // 2 minutes (reduced from 15 minutes)

// Check if auth should be blocked
const shouldBlockAuth = () => {
  // Block if rate limited
  if (isRateLimited && rateLimitTime) {
    console.log('🚫 useAuth: Rate limited, blocking');
    return Date.now() - rateLimitTime! < RATE_LIMIT_BLOCK_DURATION;
  }
  
  // Block if auth failed
  if (hasAuthFailed && authFailureTime) {
    console.log('🚫 useAuth: Auth failed recently, blocking');
    return Date.now() - authFailureTime! < AUTH_BLOCK_DURATION;
  }
  
  console.log('🔓 useAuth: Auth allowed');
  return false;
};

// Reset auth block
const resetAuthBlock = () => {
  hasAuthFailed = false;
  authFailureTime = null;
  isRateLimited = false;
  rateLimitTime = null;
};

// Set rate limit block
const setRateLimitBlock = () => {
  isRateLimited = true;
  rateLimitTime = Date.now();
  logger.warn('Auth rate limited - blocking for 15 minutes');
};

export function useAuth() {
  return useQuery<User | null>({
    queryKey: ['auth'],
    queryFn: async () => {
      console.log('🔐 useAuth: Starting auth check');
      
      // Block auth calls if previously failed within block duration
      if (shouldBlockAuth()) {
        console.log('🚫 useAuth: Auth blocked due to previous failure or rate limit');
        return null;
      }

      try {
        console.log('📡 useAuth: Calling authApi.getMe()');
        const userResponse = await authApi.getMe();
        console.log('✅ useAuth: Auth successful, user:', userResponse.user);
        // Reset failure state on successful auth
        resetAuthBlock();
        return userResponse.user;
      } catch (error: unknown) {
        console.error('❌ useAuth: Auth check failed:', error);
        
        const apiError = APIErrorHandler.getError(error);
        if (!apiError) {
          // Unknown error, set generic auth failure
          hasAuthFailed = true;
          authFailureTime = Date.now();
          if (window.location.pathname !== '/login') {
            window.location.href = '/login';
          }
          return null;
        }

        // Check if this is a rate limiting error
        if (APIErrorHandler.isRateLimitError(apiError)) {
          setRateLimitBlock();
          return null;
        }
        
        // For auth errors, block for longer period and redirect
        if (APIErrorHandler.isAuthError(apiError)) {
          hasAuthFailed = true;
          authFailureTime = Date.now();
          
          if (window.location.pathname !== '/login') {
            window.location.href = '/login';
          }
          return null;
        }
        
        // For other errors, log but don't block
        logger.warn('Non-auth error in auth check:', apiError);
        return null;
      }
    },
    retry: 1,
    retryDelay: attemptIndex => Math.min(1000 * 2 ** attemptIndex, 5000),
    staleTime: 30 * 1000, // 30 seconds (reduced from 2 minutes)
    gcTime: 5 * 60 * 1000, // 5 minutes (was cacheTime)
    refetchOnWindowFocus: true, // Enable to check auth when user returns to tab
    refetchOnReconnect: true,
    refetchInterval: false,
    refetchIntervalInBackground: false,
    enabled: !shouldBlockAuth(),
    // Prevent query from running if blocked
    structuralSharing: false,
  });
}

export function useLogout() {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      await authApi.logout();
    },
    onSuccess: () => {
      // Reset auth block on successful logout
      resetAuthBlock();
      // Clear the cache so no sensitive data remains
      qc.clear();
      // Redirect to login
      window.location.href = '/login';
    },
    onError: (error) => {
      logger.error('Logout failed:', error);
      // Still clear cache and redirect on logout error
      qc.clear();
      window.location.href = '/login';
    }
  });
}

// Export function to manually reset auth block (useful for testing)
export const resetAuthFailure = () => {
  resetAuthBlock();
};