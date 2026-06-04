import { useEffect, useState } from 'react';
import './LoginCard.css';

interface LoginCardProps {
  errorCode?: string;
  errorMessage?: string;
}

// Map error codes to descriptive titles & SVG icon names
function errorMeta(code?: string): {
  title: string;
  icon: 'lock' | 'clock' | 'shield' | 'alert';
} {
  if (!code) return { title: 'Something went wrong', icon: 'alert' };
  if (code.includes('RATE') || code.includes('MANY'))
    return { title: 'Too many attempts', icon: 'shield' };
  if (code.includes('EXPIRED') || code.includes('TOKEN'))
    return { title: 'Session expired', icon: 'clock' };
  if (code.includes('UNAUTHORIZED') || code.includes('AUTH'))
    return { title: 'Access denied', icon: 'lock' };
  return { title: 'Something went wrong', icon: 'alert' };
}

function ErrorIcon({ type }: { type: 'lock' | 'clock' | 'shield' | 'alert' }) {
  const icons: Record<string, JSX.Element> = {
    lock: (
      <svg
        width="22"
        height="22"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <title>Lock Icon</title>
        <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
        <path d="M7 11V7a5 5 0 0 1 10 0v4" />
      </svg>
    ),
    clock: (
      <svg
        width="22"
        height="22"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <title>Clock Icon</title>
        <circle cx="12" cy="12" r="10" />
        <polyline points="12 6 12 12 16 14" />
      </svg>
    ),
    shield: (
      <svg
        width="22"
        height="22"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <title>Shield Icon</title>
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
        <line x1="12" y1="8" x2="12" y2="12" />
        <line x1="12" y1="16" x2="12.01" y2="16" />
      </svg>
    ),
    alert: (
      <svg
        width="22"
        height="22"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <title>Alert Icon</title>
        <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
        <line x1="12" y1="9" x2="12" y2="13" />
        <line x1="12" y1="17" x2="12.01" y2="17" />
      </svg>
    ),
  };
  return <span className="login-error-icon">{icons[type]}</span>;
}

export default function LoginCard({ errorCode, errorMessage }: LoginCardProps) {
  const [showError, setShowError] = useState(false);
  const loginUrl = `/api/auth/google/login?redirect=${encodeURIComponent(
    `${window.location.origin}/oauth/callback`,
  )}`;

  const meta = errorMeta(errorCode);

  // Animate error in after mount
  useEffect(() => {
    if (errorMessage) {
      const t = setTimeout(() => setShowError(true), 80);
      return () => clearTimeout(t);
    }
    setShowError(false);
  }, [errorMessage]);

  return (
    <div className="login-page">
      {/* Decorative blobs */}
      <div className="login-blob login-blob-1" />
      <div className="login-blob login-blob-2" />

      <div className="login-card-wrapper">
        {/* Error toast above card */}
        {errorMessage && (
          <div
            className={`login-error-toast ${showError ? 'login-error-toast--visible' : ''}`}
            role="alert"
          >
            <ErrorIcon type={meta.icon} />
            <div className="login-error-body">
              <span className="login-error-title">{meta.title}</span>
              <span className="login-error-message">{errorMessage}</span>
              {errorCode && (
                <code className="login-error-code">{errorCode}</code>
              )}
            </div>
            <button
              className="login-error-dismiss"
              onClick={() => setShowError(false)}
              aria-label="Dismiss"
              type="button"
            >
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2.5"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <title>Dismiss Icon</title>
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          </div>
        )}

        {/* Main card */}
        <div className="login-card">
          <div className="login-card-header">
            <div className="login-logo">
              <svg
                width="32"
                height="32"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <title>App Logo</title>
                <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
                <polyline points="22 4 12 14.01 9 11.01" />
              </svg>
            </div>
            <h1 className="login-title">Welcome back</h1>
            <p className="login-subtitle">Sign in to manage your todos</p>
          </div>

          <div className="login-card-body">
            <a className="login-google-btn" href={loginUrl}>
              <svg
                className="login-google-svg"
                width="20"
                height="20"
                viewBox="0 0 48 48"
              >
                <title>Google Logo</title>
                <path
                  fill="#EA4335"
                  d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"
                />
                <path
                  fill="#4285F4"
                  d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"
                />
                <path
                  fill="#FBBC05"
                  d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"
                />
                <path
                  fill="#34A853"
                  d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"
                />
              </svg>
              Continue with Google
            </a>

            <div className="login-divider">
              <span>Secured with OAuth 2.0</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
