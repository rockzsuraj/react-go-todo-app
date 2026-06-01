import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { CreateTodoInput, Todo, UpdateTodoInput } from '../types/todo';
import { todoApi } from '../api';
import { APIErrorHandler } from '../utils/errorHandler';
import { logger } from '../services/logger';

export const useTodos = (
	enabled: boolean, 
	page = 1, 
	limit = 10,
	sortBy?: string,
	sortOrder: 'ASC' | 'DESC' = 'ASC',
	completed?: boolean,
	assigned?: string
) =>
  useQuery({
    queryKey: ['todos', page, limit, sortBy, sortOrder, completed, assigned],
    queryFn: async () => {
      const res = await todoApi.getAll({ 
        page, 
        limit, 
        sortBy, 
        sortOrder, 
        completed, 
        assigned 
      });
      return {
        todos: (res.data?.data as Todo[]) ?? [],
        meta: res.data?.meta,
      };
    },
    enabled,
  });

export const useCreateTodo = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateTodoInput) =>
      todoApi.create(payload),

    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
    },
    onError: (error: unknown) => {
      const apiError = APIErrorHandler.getError(error);
      if (apiError) {
        const userMessage = APIErrorHandler.getUserFriendlyMessage(apiError);
        logger.error('Create todo failed:', apiError);
        // You could show a toast notification here
        console.error(userMessage);
      } else {
        logger.error('Unknown create todo error:', error);
      }
    },
  });
};

export const useUpdateTodo = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: UpdateTodoInput }) =>
      todoApi.update(id, payload),

    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
    },
    onError: (error: unknown) => {
      const apiError = APIErrorHandler.getError(error);
      if (apiError) {
        const userMessage = APIErrorHandler.getUserFriendlyMessage(apiError);
        logger.error('Update todo failed:', apiError);
        console.error(userMessage);
      } else {
        logger.error('Unknown update todo error:', error);
      }
    },
  });
};

export const useToggleTodoCompleted = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      description,
      assigned_to_name,
      completed,
    }: {
      id: number;
      description: string;
      assigned_to_name: string;
      completed: boolean;
    }) =>
      todoApi.update(id, { description, assigned_to_name, completed }),

    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
    },
    onError: (error: unknown) => {
      const apiError = APIErrorHandler.getError(error);
      if (apiError) {
        const userMessage = APIErrorHandler.getUserFriendlyMessage(apiError);
        logger.error('Toggle todo completed failed:', apiError);
        console.error(userMessage);
      } else {
        logger.error('Unknown toggle todo error:', error);
      }
    },
  });
};

export const useDeleteTodo = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => todoApi.delete(id),

    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
    },
    onError: (error: unknown) => {
      const apiError = APIErrorHandler.getError(error);
      if (apiError) {
        const userMessage = APIErrorHandler.getUserFriendlyMessage(apiError);
        logger.error('Delete todo failed:', apiError);
        console.error(userMessage);
      } else {
        logger.error('Unknown delete todo error:', error);
      }
    },
  });
};