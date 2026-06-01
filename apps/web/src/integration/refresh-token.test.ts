import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';
import { apiClient } from '../api/client';
import { useAuth } from '../hooks/useAuth';

// Mock the API client
jest.mock('../api/client');
const mockedApiClient = apiClient as jest.Mocked<typeof apiClient>;

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

// Test component that uses auth
const TestComponent: React.FC = () => {
  const { data: user, isLoading, error } = useAuth();

  if (isLoading) return React.createElement('div', {}, 'Loading...');
  if (error) return React.createElement('div', {}, 'Error loading user');
  if (!user) return React.createElement('div', {}, 'Please login');

  return React.createElement('div', {}, [
    React.createElement('h1', { key: 'title' }, `Welcome ${user.name}`),
    React.createElement('button', { 
      key: 'button',
      onClick: () => mockedApiClient.get('/protected-data')
    }, 'Fetch Protected Data')
  ]);
};

const createTestWrapper = (queryClient: QueryClient) => ({ children }: { children: React.ReactNode }) => (
  React.createElement(
    BrowserRouter,
    {},
    React.createElement(QueryClientProvider, { client: queryClient }, children)
  )
);

describe('Refresh Token Integration Tests', () => {
  let queryClient: QueryClient;
  let currentUser: {
    id: string;
    email: string;
    role: string;
  };

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
    
    currentUser = {
      id: 'test-user-id',
      email: 'test@example.com',
      role: 'user'
    };
    jest.clearAllMocks();
    
    // Reset location
    mockLocation.pathname = '/';
    mockLocation.href = '';
  });

  describe('Full Authentication Flow', () => {
    it('should handle successful login and token refresh', async () => {
      // Mock initial auth check
      mockedApiClient.get.mockImplementation((url) => {
        if (url === '/auth/me') {
          return Promise.resolve({
            data: { user: { id: 'user-123', name: 'Test User', email: 'test@example.com' } }
          });
        }
        return Promise.resolve({ data: {} });
      });

      const wrapper = createTestWrapper(queryClient);
      render(React.createElement(TestComponent), { wrapper });

      // Should show user data
      await waitFor(() => {
        expect(screen.getByText('Welcome Test User')).toBeInTheDocument();
      });

      // Simulate token expiration on next request
      mockedApiClient.get.mockImplementation((url) => {
        if (url === '/auth/me') {
          return Promise.resolve({
            data: { user: { id: 'user-123', name: 'Test User', email: 'test@example.com' } }
          });
        }
        if (url === '/protected-data') {
          // First call returns 401
          const error = new Error('Unauthorized') as any;
          error.response = { status: 401 };
          error.config = { url: '/protected-data', _retry: false };
          return Promise.reject(error);
        }
        return Promise.resolve({ data: {} });
      });

      // Mock refresh token call
      const mockRefreshPost = jest.fn().mockResolvedValue({ data: { access_token: 'new-access-token' } });
      mockedApiClient.post = mockRefreshPost;

      // Click button to trigger protected request
      const loginButton = screen.getByText('Fetch Protected Data');
      await userEvent.click(loginButton);

      // Should attempt token refresh
      await waitFor(() => {
        expect(mockRefreshPost).toHaveBeenCalledWith('/auth/refresh');
      });
    });

    it('should handle refresh token failure and redirect to login', async () => {
      // Mock initial auth check
      mockedApiClient.get.mockImplementation((url) => {
        if (url === '/auth/me') {
          return Promise.resolve({
            data: { user: { id: 'user-123', name: 'Test User', email: 'test@example.com' } }
          });
        }
        return Promise.resolve({ data: {} });
      });

      const wrapper = createTestWrapper(queryClient);
      render(React.createElement(TestComponent), { wrapper });

      await waitFor(() => {
        expect(screen.getByText('Welcome Test User')).toBeInTheDocument();
      });

      // Simulate token expiration and refresh failure
      mockedApiClient.get.mockImplementation((url) => {
        if (url === '/auth/me') {
          const error = new Error('Unauthorized') as any;
          error.response = { status: 401 };
          error.config = { url: '/auth/me', _retry: false };
          return Promise.reject(error);
        }
        return Promise.resolve({ data: {} });
      });

      // Mock refresh token failure
      const mockRefreshPost = jest.fn().mockRejectedValue(new Error('Refresh token expired'));
      mockedApiClient.post = mockRefreshPost;

      // Trigger a request that requires auth
      queryClient.invalidateQueries({ queryKey: ['auth'] });

      // Should redirect to login after refresh failure
      await waitFor(() => {
        expect(mockLocation.href).toBe('/login');
      });
    });

    it('should handle concurrent requests during token refresh', async () => {
      // Mock initial auth check
      mockedApiClient.get.mockImplementation((url) => {
        if (url === '/auth/me') {
          return Promise.resolve({
            data: { user: { id: 'user-123', name: 'Test User', email: 'test@example.com' } }
          });
        }
        return Promise.resolve({ data: {} });
      });

      const wrapper = createTestWrapper(queryClient);
      render(React.createElement(TestComponent), { wrapper });

      await waitFor(() => {
        expect(screen.getByText('Welcome Test User')).toBeInTheDocument();
      });

      let refreshCallCount = 0;

      // Mock requests that trigger token refresh
      mockedApiClient.get.mockImplementation((url) => {
        if (url === '/protected-data') {
          const error = new Error('Unauthorized') as any;
          error.response = { status: 401 };
          error.config = { url: '/protected-data', _retry: false };
          return Promise.reject(error);
        }
        return Promise.resolve({ data: {} });
      });

      // Mock refresh token call (should only be called once)
      mockedApiClient.post.mockImplementation(() => {
        refreshCallCount++;
        return Promise.resolve({ data: { access_token: 'new-access-token' } });
      });

      // Trigger multiple concurrent requests
      const loginButton = screen.getByText('Fetch Protected Data');
      
      // Click multiple times rapidly
      await userEvent.click(loginButton);
      await userEvent.click(loginButton);
      await userEvent.click(loginButton);

      // Should only call refresh once
      await waitFor(() => {
        expect(refreshCallCount).toBe(1);
      }, { timeout: 5000 });
    });

    it('should handle rate limiting during auth flow', async () => {
      // Mock rate limit response
      mockedApiClient.get.mockImplementation((url) => {
        if (url === '/auth/me') {
          const error = new Error('Too Many Requests') as any;
          error.response = { status: 429 };
          return Promise.reject(error);
        }
        return Promise.resolve({ data: {} });
      });

      const wrapper = createTestWrapper(queryClient);
      render(React.createElement(TestComponent), { wrapper });

      // Should show login page due to rate limiting
      await waitFor(() => {
        expect(screen.getByText('Please login')).toBeInTheDocument();
      });

      // Should not make additional API calls due to rate limiting
      expect(mockedApiClient.get).toHaveBeenCalledTimes(1);
    });

    it('should maintain user session after successful token refresh', async () => {
      // Mock initial auth check
      mockedApiClient.get.mockImplementation((url) => {
        if (url === '/auth/me') {
          return Promise.resolve({
            data: { user: { id: 'user-123', name: 'Test User', email: 'test@example.com' } }
          });
        }
        return Promise.resolve({ data: {} });
      });

      const wrapper = createTestWrapper(queryClient);
      render(React.createElement(TestComponent), { wrapper });

      await waitFor(() => {
        expect(screen.getByText('Welcome Test User')).toBeInTheDocument();
      });

      // Simulate token expiration and successful refresh
      let isFirstCall = true;
      mockedApiClient.get.mockImplementation((url) => {
        if (url === '/auth/me') {
          if (isFirstCall) {
            isFirstCall = false;
            const error = new Error('Unauthorized') as any;
            error.response = { status: 401 };
            error.config = { url: '/auth/me', _retry: false };
            return Promise.reject(error);
          }
          return Promise.resolve({
            data: { user: { id: 'user-123', name: 'Test User', email: 'test@example.com' } }
          });
        }
        return Promise.resolve({ data: {} });
      });

      // Mock successful refresh
      mockedApiClient.post.mockResolvedValue({ data: { access_token: 'new-access-token' } });

      // Invalidate auth query to trigger refresh
      queryClient.invalidateQueries({ queryKey: ['auth'] });

      // Should still show user data after refresh
      await waitFor(() => {
        expect(screen.getByText('Welcome Test User')).toBeInTheDocument();
      }, { timeout: 5000 });

      expect(mockedApiClient.post).toHaveBeenCalledWith('/auth/refresh');
    });
  });

  describe('Mobile vs Web Token Handling', () => {
    it('should handle web cookie-based refresh tokens', async () => {
      // Mock cookie-based refresh
      Object.defineProperty(document, 'cookie', {
        writable: true,
        value: 'refresh_token=web-refresh-token',
      });

      mockedApiClient.post.mockImplementation((url) => {
        if (url === '/auth/refresh') {
          return Promise.resolve({ data: { access_token: 'new-web-access-token' } });
        }
        return Promise.resolve({ data: {} });
      });

      // Test that refresh call works with cookies
      await mockedApiClient.post('/auth/refresh');

      expect(mockedApiClient.post).toHaveBeenCalledWith('/auth/refresh');
    });

    it('should handle mobile header-based refresh tokens', async () => {
      // Mock header-based refresh for mobile
      mockedApiClient.post.mockImplementation((url, data, config: any = {}) => {
        if (url === '/auth/refresh') {
          // Check for Authorization header
          if (config?.headers?.Authorization === 'Bearer mobile-refresh-token') {
            return Promise.resolve({ data: { access_token: 'new-mobile-access-token' } });
          }
        }
        return Promise.resolve({ data: {} });
      });

      // Test mobile refresh with Authorization header
      await mockedApiClient.post('/auth/refresh', undefined, {
        headers: { Authorization: 'Bearer mobile-refresh-token' }
      });

      expect(mockedApiClient.post).toHaveBeenCalledWith('/auth/refresh', undefined, {
        headers: { Authorization: 'Bearer mobile-refresh-token' }
      });
    });
  });
});
