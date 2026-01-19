import axios from 'axios';
import config from '../config/config';

export const apiClient = axios.create({
  baseURL: `${config.baseUrl}/api`,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,
});

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

apiClient.interceptors.response.use(
  (res) => res,
  async (err) => {
    const originalRequest = err.config;

    // Check if 401, not a retry, and NOT the refresh endpoint itself
    if (
      err.response?.status === 401 &&
      !originalRequest._retry &&
      !originalRequest.url?.includes('/auth/refresh')
    ) {
      originalRequest._retry = true;

      try {
        // Attempt refresh (using default axios to avoid interceptor loop, or just be careful)
        // We can reuse apiClient because we blocked /auth/refresh above
        const { data } = await apiClient.post<{ token: string }>('/auth/refresh');

        // Update local storage
        localStorage.setItem('auth_token', data.token);

        // Update header for retry
        originalRequest.headers.Authorization = `Bearer ${data.token}`;

        // Retry original request
        return apiClient(originalRequest);
      } catch (refreshErr) {
        // Refresh failed (e.g. cookie expired)
        localStorage.removeItem('auth_token');
        window.location.href = '/login';
        return Promise.reject(refreshErr);
      }
    }

    return Promise.reject(err);
  }
);