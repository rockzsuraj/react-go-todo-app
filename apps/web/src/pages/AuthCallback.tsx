import { useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { authApi } from '../api';
import { usePageTitle } from '../hooks/usePageTitle';
import { APIErrorHandler } from '../utils/errorHandler';

export default function AuthCallback() {
  usePageTitle('Authenticating');
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  useEffect(() => {
    const checkAuth = async () => {
      try {
        // Cookies are already set by the backend on the redirect response.
        // Just fetch the current user to populate the cache.
        const authResponse = await authApi.getMe();
        queryClient.setQueryData(['auth'], authResponse.user);
        navigate('/', { replace: true });
      } catch (error) {
        console.error('AuthCallback failed:', error);
        const apiError = APIErrorHandler.getError(error);
        const errorCode = apiError?.code ?? 'ERR_AUTH';
        const errorMsg = apiError
          ? APIErrorHandler.getUserFriendlyMessage(apiError)
          : 'Authentication failed. Please try logging in again.';
        navigate(
          `/login?error=${encodeURIComponent(errorCode)}&message=${encodeURIComponent(errorMsg)}`,
          { replace: true },
        );
      }
    };

    checkAuth();
  }, [navigate, queryClient]);

  return (
    <div className="container mt-5 text-center">
      <div className="card p-4">
        <h4 className="mb-3">Signing you in…</h4>
        <p className="text-muted">You will be redirected shortly.</p>
      </div>
    </div>
  );
}
