import { Component, type ErrorInfo, type ReactNode } from 'react';
import { logger } from '../services/logger';
import { APIErrorHandler } from '../utils/errorHandler';

interface Props {
  children?: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    logger.error('Uncaught error in component tree:', error, errorInfo);
  }

  private getErrorMessage(error: Error): string {
    const apiError = APIErrorHandler.getError(error);
    if (apiError) {
      return APIErrorHandler.getUserFriendlyMessage(apiError);
    }
    return error.message || 'An unexpected error occurred';
  }

  private getErrorTitle(error: Error): string {
    const apiError = APIErrorHandler.getError(error);
    if (apiError) {
      if (APIErrorHandler.isAuthError(apiError)) {
        return 'Authentication Error';
      }
      if (APIErrorHandler.isValidationError(apiError)) {
        return 'Validation Error';
      }
      if (APIErrorHandler.isRateLimitError(apiError)) {
        return 'Rate Limit Exceeded';
      }
    }
    return 'Something went wrong!';
  }

  public render() {
    const { hasError, error } = this.state;
    if (hasError && error) {
      const apiError = APIErrorHandler.getError(error);
      return (
        <div className="container mt-5 text-center">
          <div className="alert alert-danger" role="alert">
            <h4 className="alert-heading">{this.getErrorTitle(error)}</h4>
            <p>{this.getErrorMessage(error)}</p>
            <hr />
            <div className="d-flex gap-2 justify-content-center">
              <button
                type="button"
                className="btn btn-primary"
                onClick={() => window.location.reload()}
              >
                Refresh Page
              </button>
              {apiError && APIErrorHandler.isAuthError(apiError) && (
                <button
                  type="button"
                  className="btn btn-outline-primary"
                  onClick={() => {
                    window.location.href = '/login';
                  }}
                >
                  Go to Login
                </button>
              )}
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
