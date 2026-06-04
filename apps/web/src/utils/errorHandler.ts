// API Error interface matching backend response
export interface APIError {
  code: string;
  message: string;
  details?: string;
}

// Error codes that should trigger logout
const AUTH_ERROR_CODES = [
  'ERR_UNAUTHORIZED',
  'ERR_AUTH',
  'ERR_MISSING_TOKEN',
  'ERR_INVALID_TOKEN',
  'ERR_TOKEN_EXPIRED',
];

// Error codes that should show user-friendly messages
const VALIDATION_ERROR_CODES = [
  'ERR_INVALID_REQUEST',
  'ERR_MISSING_USER_ID',
  'ERR_INVALID_FILTER',
];

// Rate limit error codes
const RATE_LIMIT_ERROR_CODES = ['ERR_TOO_MANY_ATTEMPTS'];

export const APIErrorHandler = {
  getError(error: unknown): APIError | null {
    if (APIErrorHandler.isAxiosError(error)) {
      const responseData = error.response?.data;
      const apiError = responseData?.error;
      if (apiError && APIErrorHandler.isValidAPIError(apiError)) {
        return apiError;
      }
    }

    // Handle network errors or other issues
    if (error instanceof Error) {
      return {
        code: 'ERR_NETWORK',
        message: error.message || 'Network error occurred',
      };
    }

    return null;
  },

  isAxiosError(
    error: unknown,
  ): error is { response?: { data?: { error?: APIError } } } {
    return (
      error !== null &&
      typeof error === 'object' &&
      'response' in error &&
      typeof (error as { response?: unknown }).response === 'object'
    );
  },

  isValidAPIError(error: unknown): error is APIError {
    return (
      error !== null &&
      typeof error === 'object' &&
      'code' in error &&
      'message' in error &&
      typeof (error as { code: unknown }).code === 'string' &&
      typeof (error as { message: unknown }).message === 'string'
    );
  },

  isAuthError(error: APIError): boolean {
    return AUTH_ERROR_CODES.includes(error.code);
  },

  isValidationError(error: APIError): boolean {
    return VALIDATION_ERROR_CODES.includes(error.code);
  },

  isRateLimitError(error: APIError): boolean {
    return RATE_LIMIT_ERROR_CODES.includes(error.code);
  },

  getUserFriendlyMessage(error: APIError): string {
    switch (error.code) {
      case 'ERR_UNAUTHORIZED':
      case 'ERR_AUTH':
        return 'Please log in to continue';

      case 'ERR_MISSING_TOKEN':
        return 'Authentication required. Please log in again.';

      case 'ERR_INVALID_TOKEN':
      case 'ERR_TOKEN_EXPIRED':
        return 'Your session has expired. Please log in again.';

      case 'ERR_TOO_MANY_ATTEMPTS':
        return 'Too many attempts. Please try again later.';

      case 'ERR_INVALID_REQUEST':
        return 'Invalid request. Please check your input.';

      case 'ERR_MISSING_USER_ID':
        return 'User ID is required.';

      case 'ERR_INVALID_FILTER':
        return 'Invalid filter value.';

      case 'ERR_NETWORK':
        return 'Network error. Please check your connection.';

      default:
        return error.message || 'An error occurred';
    }
  },

  shouldRedirectToLogin(error: APIError): boolean {
    return APIErrorHandler.isAuthError(error);
  },

  shouldShowRetryMessage(error: APIError): boolean {
    return APIErrorHandler.isRateLimitError(error);
  },
};
