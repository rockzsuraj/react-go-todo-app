import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { authApi } from '../api';
import { logger } from '../services/logger';
import type { User } from '../types/user';

/**
 * Fetches the current user from the server.
 * Returns null if the user is not authenticated (any error is treated as
 * "not logged in" — the route guard will handle the redirect).
 */
export function useAuth() {
  const queryClient = useQueryClient();

  // Listen for the session-expired event fired by the Axios interceptor
  // when a token refresh fails. Update cache so route guards redirect immediately.
  useEffect(() => {
    const handleExpired = () => {
      queryClient.setQueryData(['auth'], null);
    };
    window.addEventListener('auth:session-expired', handleExpired);
    return () => {
      window.removeEventListener('auth:session-expired', handleExpired);
    };
  }, [queryClient]);

  return useQuery<User | null>({
    queryKey: ['auth'],
    queryFn: async () => {
      try {
        const response = await authApi.getMe();
        return response.user;
      } catch {
        // Any error (401, network, etc.) means not authenticated.
        // Route guards handle the redirect; the hook just returns null.
        return null;
      }
    },
    retry: false,
    staleTime: 30_000,           // re-use cached user for 30 s
    gcTime: 5 * 60_000,          // keep in cache for 5 min after unmount
    refetchOnWindowFocus: false, // disable aggressive refetching on window focus to prevent rate-limiting
    refetchOnReconnect: true,
  });
}

export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => authApi.logout(),
    onSettled: () => {
      // Clear all cached data and redirect regardless of logout API success/failure.
      queryClient.clear();
      window.location.href = '/login';
    },
    onError: (error) => {
      logger.error('[useAuth] Logout API call failed (redirecting anyway):', error);
    },
  });
}
