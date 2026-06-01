import { UserResponse } from "../types/user";
import { APIResponse } from "../types/api";
import { apiClient } from "./client";

export const authApi = {
  getMe: async (): Promise<UserResponse> => {
    const response = await apiClient.get<APIResponse<UserResponse>>('/auth/me');
    if (!response.data.data) {
      throw new Error('No user data in response');
    }
    return response.data.data;
  },
  logout: () => apiClient.post('/auth/logout'),
};