import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { authApi } from '../api/auth';
import { logger } from '../services/logger';
import { useAuth, useLogout } from './useAuth';

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
        queries: { retry: false },
      },
    });

    jest.clearAllMocks();
    mockLocation.pathname = '/';
    mockLocation.href = '';
  });

  const wrapper = ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);

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

    it('should return null on authentication failure (401)', async () => {
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
      // Hook should not surface an error — it returns null instead
      expect(result.current.isError).toBe(false);
    });

    it('should return null on network error without redirecting', async () => {
      (authApi.getMe as jest.Mock).mockRejectedValue(new Error('Network Error'));

      const { result } = renderHook(() => useAuth(), { wrapper });

      await waitFor(() => {
        expect(result.current.data).toBeNull();
      });

      // The hook must NOT redirect — that is the route guard's responsibility
      expect(mockLocation.href).toBe('');
    });

    it('should return null on rate limiting without blocking subsequent calls', async () => {
      const rateLimitError = {
        response: { status: 429 },
        message: 'Too Many Requests',
      };

      (authApi.getMe as jest.Mock).mockRejectedValue(rateLimitError);

      const { result } = renderHook(() => useAuth(), { wrapper });

      await waitFor(() => {
        expect(result.current.data).toBeNull();
      });

      expect(authApi.getMe).toHaveBeenCalledTimes(1);
      expect(result.current.isError).toBe(false);
    });
  });

  describe('useLogout', () => {
    it('should call logout API and redirect to /login on success', async () => {
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

    it('should clear cache and redirect even when logout API fails', async () => {
      const logoutError = new Error('Logout failed');
      (authApi.logout as jest.Mock).mockRejectedValue(logoutError);

      const { result } = renderHook(() => useLogout(), { wrapper });

      queryClient.setQueryData(['test'], 'test-data');
      result.current.mutate();

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(authApi.logout).toHaveBeenCalledTimes(1);
      expect(queryClient.getQueryCache().getAll()).toHaveLength(0);
      expect(mockLocation.href).toBe('/login');
      expect(logger.error).toHaveBeenCalledWith(
        '[useAuth] Logout API call failed (redirecting anyway):',
        logoutError,
      );
    });
  });
});
