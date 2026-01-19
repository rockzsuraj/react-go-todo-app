import React, { Component, ErrorInfo, ReactNode } from 'react';
import { logger } from '../services/logger';

interface Props {
    children?: ReactNode;
}

interface State {
    hasError: boolean;
}

class ErrorBoundary extends Component<Props, State> {
    public state: State = {
        hasError: false,
    };

    public static getDerivedStateFromError(_: Error): State {
        // Update state so the next render will show the fallback UI.
        return { hasError: true };
    }

    public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        logger.error("Uncaught error in component tree:", error, errorInfo);
    }

    public render() {
        if (this.state.hasError) {
            return (
                <div className="container mt-5 text-center">
                    <div className="alert alert-danger" role="alert">
                        <h4 className="alert-heading">Something went wrong!</h4>
                        <p>We're sorry, an unexpected error occurred. Please try refreshing the page.</p>
                        <hr />
                        <button
                            type="button"
                            className="btn btn-primary"
                            onClick={() => window.location.reload()}
                        >
                            Refresh Page
                        </button>
                    </div>
                </div>
            );
        }

        return this.props.children;
    }
}

export default ErrorBoundary;
