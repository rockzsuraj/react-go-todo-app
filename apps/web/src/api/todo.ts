import type { APIResponse } from '../types/api';
import type { Todo, CreateTodoInput } from '../types/todo';
import { apiClient } from './client';

export const todoApi = {
  getAll: (params?: { page?: number; limit?: number }) =>
    apiClient.get<APIResponse<Todo[]>>('/todos', { params }),

  create: (payload: CreateTodoInput) =>
    apiClient.post<APIResponse<Todo>>('/todos', payload),

  update: (id: number, payload: CreateTodoInput) =>
    apiClient.put<APIResponse<Todo>>(`/todos/${id}`, payload),

  delete: (id: number) =>
    apiClient.delete<APIResponse<null>>(`/todos/${id}`),
};