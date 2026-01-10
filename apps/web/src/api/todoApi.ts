import axios from 'axios';
import { APIResponse } from '../types/api';

export const apiClient = axios.create({
  baseURL: process.env.REACT_APP_API_URL || 'http://localhost:8080/api',
  timeout: 5000,
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': process.env.REACT_APP_API_KEY || 'dev-key-12345',
  },
});

export interface Todo {
  id: number;
  description: string;
  assigned: string;
}

export interface CreateTodoRequest {
  description: string;
  assigned: string;
}

export const todoApi = {
  getAll: () => apiClient.get<APIResponse<Todo[]>>('/todos'),
  create: (todo: CreateTodoRequest) => apiClient.post<APIResponse<any>>('/todos', todo),
  update: (id: number, todo: CreateTodoRequest) => apiClient.put<APIResponse<any>>(`/todos/${id}`, todo),
  delete: (id: number) => apiClient.delete<APIResponse<any>>(`/todos/${id}`),
};