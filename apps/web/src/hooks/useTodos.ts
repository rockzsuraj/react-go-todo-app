import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { todoApi, type Todo } from '../api/supabase';

export const useTodos = () => {
  return useQuery({
    queryKey: ['todos'],
    queryFn: async () => {
      const response = await todoApi.getAll();
      return response.data.data;
    },
    staleTime: 30000,
    refetchOnWindowFocus: false,
    refetchInterval: 2000,
  });
};

export const useCreateTodo = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (todo: Omit<Todo, 'id'>) => todoApi.create(todo),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
    },
  });
};

export const useUpdateTodo = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ id, todo }: { id: number; todo: Omit<Todo, 'id'> }) => 
      todoApi.update(id, todo),
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