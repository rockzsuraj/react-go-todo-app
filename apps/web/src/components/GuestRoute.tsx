import { useQueryClient } from '@tanstack/react-query';
import { Navigate, Outlet } from 'react-router-dom';
import type { User } from '../types/user';

/**
 * Renders child routes only when the user is NOT authenticated.
 * Used for the /login page so logged-in users are redirected away.
 *
 * Reads from the React Query cache only (no network request) so that
 * visiting /login never fires an unnecessary /auth/me call.
 */
export default function GuestRoute() {
  const queryClient = useQueryClient();
  const user = queryClient.getQueryData<User | null>(['auth']);

  if (user) {
    return <Navigate to="/" replace />;
  }

  return <Outlet />;
}
