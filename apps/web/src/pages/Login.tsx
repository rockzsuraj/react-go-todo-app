import { Navigate, useSearchParams } from 'react-router-dom';
import LoginCard from '../components/LoginCard';
import { useAuth } from '../hooks/useAuth';
import { usePageTitle } from '../hooks/usePageTitle';

export default function Login() {
  usePageTitle('Login');
  const { data: user, isLoading } = useAuth();
  const [searchParams] = useSearchParams();

  const errorCode = searchParams.get('error') ?? undefined;
  const errorMessage = searchParams.get('message') ?? undefined;

  if (isLoading) {
    return (
      <div className="container mt-5 text-center">
        Checking authentication...
      </div>
    );
  }

  if (user) {
    return <Navigate to="/" replace />;
  }

  return <LoginCard errorCode={errorCode} errorMessage={errorMessage} />;
}
