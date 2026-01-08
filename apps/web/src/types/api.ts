export interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: string;
  };
  meta?: {
    total?: number;
    page?: number;
    limit?: number;
    offset?: number;
  };
  timestamp: string;
}