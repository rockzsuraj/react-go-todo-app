import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { CreateTodoInput, Todo } from '../types/todo';
import { todoApi } from '../api';

export const useTodos = (enabled: boolean, page = 1, limit = 10) =>
  useQuery({
    queryKey: ['todos', page, limit],
    queryFn: async () => {
      const res = await todoApi.getAll({ page, limit });
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
  });
};

export const useUpdateTodo = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: CreateTodoInput }) =>
      todoApi.update(id, payload),

    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
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
  });
};