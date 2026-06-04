export interface UserResponse {
  user: User;
}

export interface User {
  id: string;
  email: string;
  name: string;
  picture: string;
  role: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
}
