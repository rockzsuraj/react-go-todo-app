import { Navigate, useSearchParams } from 'react-router-dom';
import LoginCard from '../components/LoginCard';
import { Spinner } from '../components/ui';
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
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          minHeight: '100vh',
        }}
      >
        <Spinner size="lg" variant="primary" />
      </div>
    );
  }

  if (user) {
    return <Navigate to="/" replace />;
  }

  return <LoginCard errorCode={errorCode} errorMessage={errorMessage} />;
}
