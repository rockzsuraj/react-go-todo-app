import { useQueryClient } from '@tanstack/react-query';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { usePageTitle } from '../hooks/usePageTitle';
import { authApi } from '../api';

export default function AuthCallback() {
  usePageTitle('Authenticating');
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [isChecking, setIsChecking] = useState(true);

  useEffect(() => {
    const checkAuth = async () => {
      console.log('🔍 AuthCallback: Starting auth check');
      try {
        // Directly fetch auth state to ensure we have the latest data
        console.log('📡 AuthCallback: Fetching auth data...');
        const authResponse = await authApi.getMe();
        console.log('✅ AuthCallback: Auth successful:', authResponse);
        
        // Update the query cache with fresh user data
        queryClient.setQueryData(['auth'], authResponse.user);
        console.log('💾 AuthCallback: Cache updated with user:', authResponse.user);
        
        // Also invalidate to ensure any other instances refetch
        await queryClient.invalidateQueries({ queryKey: ['auth'] });
        console.log('🔄 AuthCallback: Cache invalidated');
        
        // Wait a moment to ensure state is propagated
        await new Promise(resolve => setTimeout(resolve, 1000));
        console.log('⏰ AuthCallback: Wait completed, navigating to home...');
        
        // Navigate to home
        navigate('/', { replace: true });
      } catch (error) {
        console.error('❌ AuthCallback: Auth failed:', error);
        // If direct fetch fails, invalidate and try normal flow
        console.log('🔄 AuthCallback: Trying fallback flow...');
        await queryClient.invalidateQueries({ queryKey: ['auth'] });
        await new Promise(resolve => setTimeout(resolve, 1000));
        navigate('/', { replace: true });
      } finally {
        setIsChecking(false);
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