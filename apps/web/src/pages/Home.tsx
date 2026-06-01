import { useState } from 'react';
import NewTodoForm from '../components/NewTodoForm';
import TodoTable from '../components/TodoTable';
import LoginCard from '../components/LoginCard';
import { useTodos, useDeleteTodo, useToggleTodoCompleted } from '../hooks/useTodos';
import { useAuth } from '../hooks/useAuth';
import { usePageTitle } from '../hooks/usePageTitle';
import { Navigate } from 'react-router-dom';

export default function Home() {
  usePageTitle('Home');
  const [showAddTodoForm, setShowAddTodoForm] = useState(false);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [sortBy, setSortBy] = useState<string>('created_at');
  const [sortOrder, setSortOrder] = useState<'ASC' | 'DESC'>('DESC');
  const [filterCompleted, setFilterCompleted] = useState<boolean | undefined>(undefined);
  const [filterAssigned, setFilterAssigned] = useState<string>('');

  // 🔐 Auth state
  const { data: user, isLoading: authLoading } = useAuth();

  console.log('🏠 Home: Auth state:', { user, authLoading });

  // 📝 Todos (only when logged in)
  const {
    data: todosData,
    isLoading,
    isError,
  } = useTodos(!!user, page, limit, sortBy, sortOrder, filterCompleted, filterAssigned);

  const todos = todosData?.todos ?? [];

  // ✅ SAFE META
  const { total = 0 } = todosData?.meta ?? {};

  const deleteTodoMutation = useDeleteTodo();
  const toggleTodoCompletedMutation = useToggleTodoCompleted();

  const handleDeleteTodo = (id: number) => {
    deleteTodoMutation.mutate(id, {
      onSuccess: () => {
        // Go back a page if last item is deleted
        if (todos.length === 1 && page > 1) {
          setPage((p) => p - 1);
        }
      },
    });
  };

  const handleToggleTodoCompleted = (todo: { id: number; description: string; assigned_to_name: string; completed: boolean }) => {
    toggleTodoCompletedMutation.mutate({
      id: todo.id,
      description: todo.description,
      assigned_to_name: todo.assigned_to_name,
      completed: !todo.completed,
    });
  };

  const handleSort = (field: string) => {
    if (sortBy === field) {
      setSortOrder(sortOrder === 'ASC' ? 'DESC' : 'ASC');
    } else {
      setSortBy(field);
      setSortOrder('ASC');
    }
    setPage(1); // Reset to first page when sorting
  };

  const handleFilterChange = () => {
    setPage(1); // Reset to first page when filtering
  };

  const clearFilters = () => {
    setFilterCompleted(undefined);
    setFilterAssigned('');
    setPage(1);
  };
  // ⏳ Checking auth
  if (authLoading) {
    return <div className="container mt-5">Checking login…</div>;
  }


  // ...

  // 🔐 NOT LOGGED IN UI
  if (!user) {
    console.log('🚪 Home: No user found, redirecting to login');
    return <Navigate to="/login" replace />;
  }

  // ⏳ Todos loading
  if (isLoading) {
    return <div className="container mt-5">Loading todos...</div>;
  }

  // ❌ Error
  if (isError) {
    return (
      <div className="container mt-5 text-danger">
        Failed to load todos
      </div>
    );
  }

  // ✅ LOGGED IN UI
  return (
    <div className="container mt-5">
      <div className="card border-0 shadow-lg">
        <div className="card-header bg-gradient text-white">
          <div className="d-flex justify-content-between align-items-center">
            <h5 className="mb-0 fw-semibold">
              <i className="bi bi-list-task me-2"></i>
              My Todos
            </h5>

            <button
              type="button"
              className="btn btn-light btn-sm"
              onClick={() => setShowAddTodoForm((prev) => !prev)}
            >
              <i className="bi bi-plus-lg me-1"></i>
              {showAddTodoForm ? 'Close' : 'New Todo'}
            </button>
          </div>
        </div>

        <div className="card-body">
          {/* New Todo Form */}
          {showAddTodoForm && (
            <NewTodoForm onSuccess={() => setShowAddTodoForm(false)} />
          )}

          {/* Filters Section */}
          <div className="mb-4">
            <div className="d-flex justify-content-between align-items-center mb-3">
              <h5 className="mb-0 fw-semibold">
                <i className="bi bi-funnel me-2"></i>
                Filters
              </h5>
              <button
                type="button"
                className="btn btn-outline-secondary btn-sm"
                onClick={clearFilters}
              >
                <i className="bi bi-arrow-clockwise me-1"></i>
                Clear All
              </button>
            </div>

            <div className="row g-3">
              <div className="col-md-4">
                <label className="form-label fw-semibold">
                  <i className="bi bi-check-circle me-1"></i>
                  Status
                </label>
                <select 
                  className="form-select"
                  value={filterCompleted === undefined ? '' : filterCompleted.toString()}
                  onChange={(e) => {
                    const value = e.target.value;
                    setFilterCompleted(value === '' ? undefined : value === 'true');
                    handleFilterChange();
                  }}
                >
                  <option value="">All Tasks</option>
                  <option value="false">Incomplete</option>
                  <option value="true">Completed</option>
                </select>
              </div>
              
              <div className="col-md-4">
                <label className="form-label fw-semibold">
                  <i className="bi bi-person me-1"></i>
                  Assigned To
                </label>
                <input
                  type="text"
                  className="form-control"
                  placeholder="Search by name..."
                  value={filterAssigned}
                  onChange={(e) => {
                    setFilterAssigned(e.target.value);
                    handleFilterChange();
                  }}
                />
              </div>
              
              <div className="col-md-4">
                <label className="form-label fw-semibold">
                  <i className="bi bi-sort-down me-1"></i>
                  Sort By
                </label>
                <select 
                  className="form-select"
                  value={sortBy}
                  onChange={(e) => handleSort(e.target.value)}
                >
                  <option value="created_at">Created Date</option>
                  <option value="updated_at">Updated Date</option>
                  <option value="description">Description</option>
                  <option value="assigned_to_name">Assigned To</option>
                </select>
              </div>
            </div>

            {/* Sort Order and Results */}
            <div className="row g-3">
              <div className="col-md-6">
                <label className="form-label fw-semibold">
                  <i className="bi bi-sort-up me-1"></i>
                  Order
                </label>
                <select 
                  className="form-select"
                  value={sortOrder}
                  onChange={(e) => {
                    setSortOrder(e.target.value as 'ASC' | 'DESC');
                    setPage(1);
                  }}
                >
                  <option value="ASC">Ascending</option>
                  <option value="DESC">Descending</option>
                </select>
              </div>
              
              <div className="col-md-6">
                <label className="form-label fw-semibold">
                  <i className="bi bi-info-circle me-1"></i>
                  Results
                </label>
                <div className="form-control-plaintext bg-light">
                  <strong>Total: {total}</strong>
                  {total > limit && (
                    <span className="text-muted ms-2">
                      (Page {page} of {Math.ceil(total / limit)})
                    </span>
                  )}
                </div>
              </div>
            </div>
          </div>
          <TodoTable
            todos={todos}
            deleteTodo={handleDeleteTodo}
            toggleTodoCompleted={handleToggleTodoCompleted}
            isDeleting={deleteTodoMutation.isPending}
            sortBy={sortBy}
            sortOrder={sortOrder}
            onSort={handleSort}
          />
        </div>
      </div>
    </div>
  );
}