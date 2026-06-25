import { useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { logger } from '../services/logger';
import { useAuth } from './useAuth';

export const useTodoSSE = () => {
  const queryClient = useQueryClient();
  const { data: user } = useAuth();

  useEffect(() => {
    if (!user) return;

    logger.info('Initializing SSE connection for live updates');
    const eventSource = new EventSource('/api/todos/events', {
      withCredentials: true,
    });

    eventSource.onmessage = (event) => {
      logger.info('SSE event received:', event.data);
      if (event.data === 'todo_update') {
        queryClient.invalidateQueries({ queryKey: ['todos'] });
      }
    };

    eventSource.onerror = (error) => {
      logger.error('SSE connection error:', error);
    };

    return () => {
      logger.info('Closing SSE connection');
      eventSource.close();
    };
  }, [user, queryClient]);
};
