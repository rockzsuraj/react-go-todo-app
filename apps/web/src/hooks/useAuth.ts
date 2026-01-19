import { useQuery, useQueryClient } from '@tanstack/react-query';
import { authApi } from '../api';
import { useNavigate } from 'react-router-dom';
import { logger } from '../services/logger';

export interface User {
  id: string;
  name: string;
  email: string;
  picture: string; // Google profile image
}

export function useAuth() {
  return useQuery<{ user: User } | null>({
    queryKey: ['auth'],
    queryFn: async () => {
      const token = localStorage.getItem('auth_token');
      if (!token) return null;
      return authApi.getMe();
    },
    retry: false,
    staleTime: Infinity, // User info doesn't change often
  });
}

export function useLogout() {
  const qc = useQueryClient();
  // const navigate = useNavigate(); // Removed unused variable

  const logout = async () => {
    try {
      // 1. Call your modular API to invalidate the session on the backend
      // This will use your Axios interceptor and handle cookies/headers automatically
      await authApi.logout();
    } catch (error) {
      logger.error('Logout failed:', error);
    } finally {
      // 2. Clean up local state
      localStorage.removeItem('auth_token');

      // 3. Clear the React Query cache so the next user doesn't see old data
      qc.clear();

      // 4. Redirect to login
      // We use window.location.href for a "Hard Reset" which is safer for logout,
      // or navigate('/login') for a "Soft Reset".
      window.location.href = '/login';
    }
  };

  return logout;
}