import type { User } from "../hooks/useAuth";
import { apiClient } from "./client";


export const authApi = {
  getMe: () => apiClient.get<{ user: User }>('/auth/me').then(res => res.data),
  logout: () => apiClient.post('/auth/logout'),
};