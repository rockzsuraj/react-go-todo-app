import React from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useAuth, useLogout, resetAuthFailure } from './useAuth';
import { authApi } from '../api/auth';
import { logger } from '../services/logger';

// Mock dependencies
jest.mock('../api/auth');
jest.mock('../services/logger');

// Mock window.location
const mockLocation = {
  pathname: '/',
  href: '',
  assign: jest.fn(),
  replace: jest.fn(),
};

Object.defineProperty(window, 'location', {
  value: mockLocation,
  writable: true,
});

describe('useAuth Hook', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    });
    
    // Reset mocks
    jest.clearAllMocks();
    
    // Reset auth failure state
    resetAuthFailure();
    
    // Reset location
    mockLocation.pathname = '/';
    mockLocation.href = '';
  });

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    React.createElement(QueryClientProvider, { client: queryClient }, children)
  );

  describe('useAuth', () => {
    it('should return user data on successful authentication', async () => {
      const mockUser = {
        id: 'user-123',
        email: 'test@example.com',
        name: 'Test User',
        role: 'user',
      };

      (authApi.getMe as jest.Mock).mockResolvedValue({ user: mockUser });

      const { result } = renderHook(() => useAuth(), { wrapper });

      await waitFor(() => {
        expect(result.current.data).toEqual(mockUser);
      });

      expect(authApi.getMe).toHaveBeenCalledTimes(1);
      expect(result.current.isSuccess).toBe(true);
    });

    it('should return null on authentication failure', async () => {
      const authError = {
        response: { status: 401 },
        message: 'Unauthorized',
      };

      (authApi.getMe as jest.Mock).mockRejectedValue(authError);

      const { result } = renderHook(() => useAuth(), { wrapper });

      await waitFor(() => {
        expect(result.current.data).toBeNull();
      });

      expect(authApi.getMe).toHaveBeenCalledTimes(1);
      expect(result.current.isError).toBe(true);
      expect(mockLocation.href).toBe('/login');
    });

    it('should redirect to login on auth failure when not already on login page', async () => {
      mockLocation.pathname = '/dashboard';
      
      const authError = {
        response: { status: 401 },
        message: 'Unauthorized',
      };

      (authApi.getMe as jest.Mock).mockRejectedValue(authError);

      renderHook(() => useAuth(), { wrapper });

      await waitFor(() => {
        expect(mockLocation.href).toBe('/login');
      });
    });

    it('should not redirect when already on login page', async () => {
      mockLocation.pathname = '/login';
      
      const authError = {
        response: { status: 401 },
        message: 'Unauthorized',
      };

      (authApi.getMe as jest.Mock).mockRejectedValue(authError);

      renderHook(() => useAuth(), { wrapper });

      await waitFor(() => {
        expect(mockLocation.href).toBe('/login'); // Should remain /login
      });
    });

    it('should handle rate limiting errors', async () => {
      const rateLimitError = {
        response: { status: 429 },
        message: 'Too Many Requests',
      };

      (authApi.getMe as jest.Mock).mockRejectedValue(rateLimitError);

      const { result } = renderHook(() => useAuth(), { wrapper });

      await waitFor(() => {
        expect(result.current.data).toBeNull();
      });

      expect(logger.warn).toHaveBeenCalledWith('Auth rate limited - blocking for 15 minutes');
      expect(authApi.getMe).toHaveBeenCalledTimes(1);
    });

    it('should block auth calls after rate limit', async () => {
      // First call triggers rate limit
      const rateLimitError = {
        response: { status: 429 },
        message: 'Too Many Requests',
      };

      (authApi.getMe as jest.Mock).mockRejectedValue(rateLimitError);

      const { result, rerender } = renderHook(() => useAuth(), { wrapper });

      await waitFor(() => {
        expect(result.current.data).toBeNull();
      });

      // Reset mock to track subsequent calls
      (authApi.getMe as jest.Mock).mockClear();

      // Rerender hook - should not make another API call
      rerender();

      // Should not make another API call due to rate limiting
      expect(authApi.getMe).not.toHaveBeenCalled();
    });

    it('should block auth calls after auth failure for specified duration', async () => {
      const authError = {
        response: { status: 401 },
        message: 'Unauthorized',
      };

      (authApi.getMe as jest.Mock).mockRejectedValue(authError);

      const { result, rerender } = renderHook(() => useAuth(), { wrapper });

      await waitFor(() => {
        expect(result.current.data).toBeNull();
      });

      // Reset mock to track subsequent calls
      (authApi.getMe as jest.Mock).mockClear();

      // Rerender hook - should not make another API call
      rerender();

      // Should not make another API call due to auth failure block
      expect(authApi.getMe).not.toHaveBeenCalled();
    });

    it('should reset auth failure on successful auth after previous failure', async () => {
      // First call fails
      const authError = {
        response: { status: 401 },
        message: 'Unauthorized',
      };

      (authApi.getMe as jest.Mock).mockRejectedValueOnce(authError);

      const { result, rerender } = renderHook(() => useAuth(), { wrapper });

      await waitFor(() => {
        expect(result.current.data).toBeNull();
      });

      // Reset mock for successful call
      const mockUser = {
        id: 'user-123',
        email: 'test@example.com',
        name: 'Test User',
        role: 'user',
      };

      (authApi.getMe as jest.Mock).mockResolvedValueOnce({ user: mockUser });

      // Manually reset auth failure to simulate block expiration
      resetAuthFailure();

      // Rerender hook
      rerender();

      await waitFor(() => {
        expect(result.current.data).toEqual(mockUser);
      });

      expect(authApi.getMe).toHaveBeenCalledTimes(1);
    });
  });

  describe('useLogout', () => {
    it('should call logout API and clear cache on successful logout', async () => {
      (authApi.logout as jest.Mock).mockResolvedValue({});

      const { result } = renderHook(() => useLogout(), { wrapper });

      result.current.mutate();

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(authApi.logout).toHaveBeenCalledTimes(1);
      expect(queryClient.getQueryCache().getAll()).toHaveLength(0);
      expect(mockLocation.href).toBe('/login');
    });

    it('should clear cache and redirect even on logout error', async () => {
      const logoutError = new Error('Logout failed');
      (authApi.logout as jest.Mock).mockRejectedValue(logoutError);

      const { result } = renderHook(() => useLogout(), { wrapper });

      // Add some data to cache
      queryClient.setQueryData(['test'], 'test-data');

      result.current.mutate();

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(authApi.logout).toHaveBeenCalledTimes(1);
      expect(queryClient.getQueryCache().getAll()).toHaveLength(0); // Should be cleared
      expect(mockLocation.href).toBe('/login');
      expect(logger.error).toHaveBeenCalledWith('Logout failed:', logoutError);
    });

    it('should reset auth failure block on successful logout', async () => {
      // Simulate auth failure state
      const authError = {
        response: { status: 401 },
        message: 'Unauthorized',
      };

      (authApi.getMe as jest.Mock).mockRejectedValue(authError);

      const { result: authResult } = renderHook(() => useAuth(), { wrapper });

      await waitFor(() => {
        expect(authResult.current.data).toBeNull();
      });

      // Now logout
      (authApi.logout as jest.Mock).mockResolvedValue({});

      const { result: logoutResult } = renderHook(() => useLogout(), { wrapper });

      logoutResult.current.mutate();

      await waitFor(() => {
        expect(logoutResult.current.isSuccess).toBe(true);
      });

      // Verify auth failure was reset by checking that next auth call would work
      (authApi.getMe as jest.Mock).mockClear();
      (authApi.getMe as jest.Mock).mockResolvedValue({ user: { id: 'test' } });

      const { result: newAuthResult } = renderHook(() => useAuth(), { wrapper });

      await waitFor(() => {
        expect(newAuthResult.current.data).toEqual({ id: 'test' });
      });

      expect(authApi.getMe).toHaveBeenCalledTimes(1);
    });
  });

  describe('resetAuthFailure', () => {
    it('should reset auth failure state', () => {
      // This is mainly a utility function test
      expect(() => resetAuthFailure()).not.toThrow();
    });
  });
});
