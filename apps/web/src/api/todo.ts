import type { APIResponse } from '../types/api';
import type { Todo, CreateTodoInput, UpdateTodoInput } from '../types/todo';
import { apiClient } from './client';

export const todoApi = {
  getAll: (params?: { 
    page?: number; 
    limit?: number;
    sortBy?: string;
    sortOrder?: 'ASC' | 'DESC';
    completed?: boolean;
    assigned?: string;
  }) => {
    const queryParams = params
      ? {
          page: params.page,
          limit: params.limit,
          sort_by: params.sortBy,
          sort_order: params.sortOrder,
          completed: params.completed,
          assigned: params.assigned,
        }
      : undefined;

    return apiClient.get<APIResponse<Todo[]>>('/todos', { params: queryParams });
  },

  create: (payload: CreateTodoInput) =>
    apiClient.post<APIResponse<Todo>>('/todos', payload),

  update: (id: number, payload: UpdateTodoInput) =>
    apiClient.put<APIResponse<Todo>>(`/todos/${id}`, payload),

  delete: (id: number) =>
    apiClient.delete<APIResponse<null>>(`/todos/${id}`),
};