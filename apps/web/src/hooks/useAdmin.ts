import { useMutation, useQueryClient } from '@tanstack/react-query';
import { AxiosResponse } from 'axios';
import { adminApi } from '../api/admin';
import { APIErrorHandler } from '../utils/errorHandler';
import { logger } from '../services/logger';
import type { APIResponse } from '../types/api';

// Define proper types for TanStack Query mutations
interface RevokeUserVariables {
  userID: string;
}

interface UnblockUserVariables {
  userID: string;
}

export const useRevokeUser = () => {
  const queryClient = useQueryClient();

  return useMutation<AxiosResponse<APIResponse<null>>, Error, RevokeUserVariables>({
    mutationFn: (variables: RevokeUserVariables) => {
      if (!variables.userID.trim()) {
        throw new Error('User ID is required');
      }
      return adminApi.revokeUser(variables.userID);
    },
    onSuccess: (_data: AxiosResponse<APIResponse<null>>, variables: RevokeUserVariables) => {
      // Invalidate any user list or auth state if needed
      queryClient.invalidateQueries({ queryKey: ['admin'] });
      if (variables?.userID) {
        logger.info('User revoked successfully:', { userID: variables.userID });
      }
    },
    onError: (error: unknown) => {
      const apiError = APIErrorHandler.getError(error);
      if (apiError) {
        const userMessage = APIErrorHandler.getUserFriendlyMessage(apiError);
        logger.error('Revoke user failed:', apiError);
        console.error(userMessage);
      } else {
        logger.error('Unknown revoke user error:', error);
      }
    },
  });
};

export const useUnblockUser = () => {
  const queryClient = useQueryClient();

  return useMutation<AxiosResponse<APIResponse<null>>, Error, UnblockUserVariables>({
    mutationFn: (variables: UnblockUserVariables) => adminApi.unblockUser(variables.userID),
    onSuccess: (_data: AxiosResponse<APIResponse<null>>, variables: UnblockUserVariables) => {
      // Invalidate any user list or auth state if needed
      queryClient.invalidateQueries({ queryKey: ['admin'] });
      if (variables?.userID) {
        logger.info('User unblocked successfully:', { userID: variables.userID });
      }
    },
    onError: (error: unknown) => {
      const apiError = APIErrorHandler.getError(error);
      if (apiError) {
        const userMessage = APIErrorHandler.getUserFriendlyMessage(apiError);
        logger.error('Unblock user failed:', apiError);
        console.error(userMessage);
      } else {
        logger.error('Unknown unblock user error:', error);
      }
    },
  });
};
