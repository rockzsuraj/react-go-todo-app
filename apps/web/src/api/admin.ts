import type { APIResponse } from '../types/api';
import { apiClient } from './client';

export const adminApi = {
  revokeUser: (userID: string) =>
    apiClient.post<APIResponse<null>>('/admin/revoke-user', { user_id: userID }),
    
  unblockUser: (userID: string) =>
    apiClient.post<APIResponse<null>>('/admin/unblock-user', { user_id: userID }),
};
