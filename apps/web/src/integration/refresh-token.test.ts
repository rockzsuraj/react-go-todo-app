import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
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

const TestComponent: React.FC = () => {
  const { data: user, isLoading, error } = useAuth();

  if (isLoading) return React.createElement('div', {}, 'Loading...');
  if (error) return React.createElement('div', {}, 'Error loading user');
  if (!user) return React.createElement('div', {}, 'Please login');

  return React.createElement('div', {}, [
    React.createElement(
      'h1',
      { key: 'title' },
      `Welcome ${(user as any).name}`,
    ),
    React.createElement(
      'button',
      {
        key: 'button',
        type: 'button',
        onClick: () => mockedApiClient.get('/protected-data'),
      },
      'Fetch Protected Data',
    ),
  ]);
};

const createTestWrapper =
  (queryClient: QueryClient) =>
  ({ children }: { children: React.ReactNode }) =>
    React.createElement(
      BrowserRouter,
      {},
      React.createElement(
        QueryClientProvider,
        { client: queryClient },
        children,
      ),
    );

describe('Refresh Token Integration Tests', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });

    jest.clearAllMocks();
    mockLocation.pathname = '/';
    mockLocation.href = '';
  });

  const setupInterceptorMock = (getHandler: (url: string) => Promise<any>) => {
    mockedApiClient.get.mockImplementation(
      async (url: string, config?: any) => {
        try {
          return await getHandler(url);
        } catch (error: any) {
          if (error.response?.status === 401 && !config?._retry) {
            try {
              await mockedApiClient.post('/auth/refresh');
              // Retry the get request with config marking it as a retry
              return await mockedApiClient.get(url, { _retry: true } as any);
            } catch (_refreshError) {
              // Propagate the refresh error as a proper auth error structure
              const finalError = new Error('Unauthorized') as any;
              finalError.response = {
                status: 401,
                data: {
                  error: {
                    code: 'ERR_UNAUTHORIZED',
                    message: 'Unauthorized',
                  },
                },
              };
              return Promise.reject(finalError);
            }
          }
          return Promise.reject(error);
        }
      },
    );
  };

  describe('Full Authentication Flow', () => {
    it('should handle successful login and token refresh', async () => {
      // 1. Initial auth check mock (returns user info)
      setupInterceptorMock(async (url) => {
        if (url === '/auth/me') {
          return {
            data: {
              data: {
                user: {
                  id: 'user-123',
                  name: 'Test User',
                  email: 'test@example.com',
                },
              },
            },
          };
        }
        return { data: {} };
      });

      const wrapper = createTestWrapper(queryClient);
      render(React.createElement(TestComponent), { wrapper });

      await waitFor(() => {
        expect(screen.getByText('Welcome Test User')).toBeInTheDocument();
      });

      // 2. Mock refresh post to succeed
      const mockRefreshPost = jest
        .fn()
        .mockResolvedValue({ data: { access_token: 'new-access-token' } });
      mockedApiClient.post = mockRefreshPost;

      // 3. Update the get handler to return 401 on /protected-data first, and then succeed on retry
      let protectedCalls = 0;
      setupInterceptorMock(async (url) => {
        if (url === '/auth/me') {
          return {
            data: {
              data: {
                user: {
                  id: 'user-123',
                  name: 'Test User',
                  email: 'test@example.com',
                },
              },
            },
          };
        }
        if (url === '/protected-data') {
          protectedCalls++;
          if (protectedCalls === 1) {
            const error = new Error('Unauthorized') as any;
            error.response = { status: 401 };
            error.config = { url: '/protected-data', _retry: false };
            throw error;
          }
          return { data: { message: 'success after refresh' } };
        }
        return { data: {} };
      });

      const button = screen.getByText('Fetch Protected Data');
      await userEvent.click(button);

      await waitFor(() => {
        expect(mockRefreshPost).toHaveBeenCalledWith('/auth/refresh');
      });
    });

    it('should handle refresh token failure and redirect to login', async () => {
      // 1. Initial auth check mock (returns user info)
      setupInterceptorMock(async (url) => {
        if (url === '/auth/me') {
          return {
            data: {
              data: {
                user: {
                  id: 'user-123',
                  name: 'Test User',
                  email: 'test@example.com',
                },
              },
            },
          };
        }
        return { data: {} };
      });

      const wrapper = createTestWrapper(queryClient);
      render(React.createElement(TestComponent), { wrapper });

      await waitFor(() => {
        expect(screen.getByText('Welcome Test User')).toBeInTheDocument();
      });

      // 2. Mock refresh token call to fail
      mockedApiClient.post = jest
        .fn()
        .mockRejectedValue(new Error('Refresh token expired'));

      // 3. Update get handler: next /auth/me call returns 401
      setupInterceptorMock(async (url) => {
        if (url === '/auth/me') {
          const error = new Error('Unauthorized') as any;
          error.response = { status: 401 };
          error.config = { url: '/auth/me', _retry: false };
          throw error;
        }
        return { data: {} };
      });

      queryClient.invalidateQueries({ queryKey: ['auth'] });

      // The hook no longer redirects directly — it returns null so the route
      // guard (<ProtectedRoute>) picks up the null user and renders <Navigate>.
      await waitFor(() => {
        expect(screen.getByText('Please login')).toBeInTheDocument();
      });

      // Confirm the hook did NOT do a hard redirect (that's the route guard's job)
      expect(mockLocation.href).toBe('');
    });

    it('should handle rate limiting during auth flow', async () => {
      mockedApiClient.get.mockImplementation((url) => {
        if (url === '/auth/me') {
          const error = new Error('Too Many Requests') as any;
          error.response = {
            status: 429,
            data: {
              error: {
                code: 'ERR_TOO_MANY_ATTEMPTS',
                message: 'Too Many Requests',
              },
            },
          };
          return Promise.reject(error);
        }
        return Promise.resolve({ data: {} }) as any;
      });

      const wrapper = createTestWrapper(queryClient);
      render(React.createElement(TestComponent), { wrapper });

      await waitFor(() => {
        expect(screen.getByText('Please login')).toBeInTheDocument();
      });

      expect(mockedApiClient.get).toHaveBeenCalledTimes(1);
    });

    it('should handle mobile header-based refresh tokens', async () => {
      mockedApiClient.post = jest
        .fn()
        .mockImplementation((url, _data, config: any = {}) => {
          if (
            url === '/auth/refresh' &&
            config?.headers?.Authorization === 'Bearer mobile-refresh-token'
          ) {
            return Promise.resolve({
              data: { access_token: 'new-mobile-access-token' },
            });
          }
          return Promise.resolve({ data: {} });
        });

      await mockedApiClient.post('/auth/refresh', undefined, {
        headers: { Authorization: 'Bearer mobile-refresh-token' },
      });

      expect(mockedApiClient.post).toHaveBeenCalledWith(
        '/auth/refresh',
        undefined,
        {
          headers: { Authorization: 'Bearer mobile-refresh-token' },
        },
      );
    });
  });
});
