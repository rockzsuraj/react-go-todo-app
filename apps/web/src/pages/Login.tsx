import { Navigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { usePageTitle } from '../hooks/usePageTitle';
import LoginCard from '../components/LoginCard';

export default function Login() {
  usePageTitle('Login');
  const { data: user, isLoading } = useAuth();

  if (isLoading) {
    return <div className="container mt-5 text-center">Checking authentication...</div>;
  }

  if (user) {
    return <Navigate to="/" replace />;
  }

  return <LoginCard />;
}
