import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import LoadingSkeleton from './LoadingSkeleton';

/**
 * Renders child routes only when the user is authenticated.
 * Shows a loading skeleton while the auth state is resolving,
 * then redirects to /login if unauthenticated.
 */
export default function ProtectedRoute() {
  const { data: user, isLoading } = useAuth();

  if (isLoading) {
    return <LoadingSkeleton />;
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}
