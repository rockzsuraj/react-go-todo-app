import { useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { logger } from '../services/logger';
import { usePageTitle } from '../hooks/usePageTitle';

export default function AuthCallback() {
  usePageTitle('Authenticating');
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  useEffect(() => {
    try {
      const params = new URLSearchParams(window.location.search);
      const token = params.get('token');
      if (token) {
        localStorage.setItem('auth_token', token);

        // Invalidate auth query so other components know we are logged in
        queryClient.invalidateQueries({ queryKey: ['auth'] });

        params.delete('token');
        const newSearch = params.toString();
        const newUrl = '/' + (newSearch ? '?' + newSearch : '');
        // Replace history so token is not visible in back button
        navigate(newUrl, { replace: true });
        return;
      }

      // ...
    } catch (err) {
      logger.error('Error processing auth callback:', err);

    }

    // If no token present, just go home
    navigate('/', { replace: true });
  }, [navigate]);

  return (
    <div className="container mt-5 text-center">
      <div className="card p-4">
        <h4 className="mb-3">Signing you in…</h4>
        <p className="text-muted">You will be redirected shortly.</p>
      </div>
    </div>
  );
}
