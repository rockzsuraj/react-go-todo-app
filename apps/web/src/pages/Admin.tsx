import { useState } from 'react';
import { useRevokeUser, useUnblockUser } from '../hooks/useAdmin';
import { useAuth } from '../hooks/useAuth';
import { Navigate } from 'react-router-dom';
import { APIErrorHandler } from '../utils/errorHandler';

export default function Admin() {
  const { data: user } = useAuth();
  const [userID, setUserID] = useState('');
  const revokeUserMutation = useRevokeUser();
  const unblockUserMutation = useUnblockUser();

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  // Simple role check; adjust if your user object has a role field
  if (user.role !== 'admin') {
    return <div className="container mt-5 text-danger">Access denied: Admins only</div>;
  }

  const handleRevoke = () => {
    if (!userID.trim()) return;
    revokeUserMutation.mutate({ userID }, {
      onSuccess: () => {
        setUserID('');
        alert('User revoked successfully');
      },
      onError: (err: unknown) => {
        const apiError = APIErrorHandler.getError(err);
        const message = apiError ? APIErrorHandler.getUserFriendlyMessage(apiError) : 'Failed to revoke user';
        alert(message);
      },
    });
  };

  const handleUnblock = () => {
    if (!userID.trim()) return;
    unblockUserMutation.mutate({ userID }, {
      onSuccess: () => {
        setUserID('');
        alert('User unblocked successfully');
      },
      onError: (err: unknown) => {
        const apiError = APIErrorHandler.getError(err);
        const message = apiError ? APIErrorHandler.getUserFriendlyMessage(apiError) : 'Failed to unblock user';
        alert(message);
      },
    });
  };

  return (
    <div className="container mt-5">
      <div className="card">
        <div className="card-header">
          <h5 className="mb-0">Admin Panel</h5>
        </div>
        <div className="card-body">
          <div className="mb-3">
            <label htmlFor="user-id" className="form-label">User ID</label>
            <input
              id="user-id"
              type="text"
              className="form-control"
              placeholder="Enter user ID to revoke/unblock"
              value={userID}
              onChange={(e) => setUserID(e.target.value)}
            />
          </div>

          <div className="d-flex gap-2">
            <button
              className="btn btn-danger"
              onClick={handleRevoke}
              disabled={!userID.trim() || revokeUserMutation.isPending}
            >
              {revokeUserMutation.isPending ? 'Revoking...' : 'Revoke User'}
            </button>

            <button
              className="btn btn-success"
              onClick={handleUnblock}
              disabled={!userID.trim() || unblockUserMutation.isPending}
            >
              {unblockUserMutation.isPending ? 'Unblocking...' : 'Unblock User'}
            </button>
          </div>

          <div className="mt-3 text-muted">
            <small>
              Revoking a user will invalidate all their JWTs immediately. Unblock restores access.
            </small>
          </div>
        </div>
      </div>
    </div>
  );
}
