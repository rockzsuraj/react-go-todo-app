import { useState } from 'react';
import NewTodoForm from '../components/NewTodoForm';
import TodoTable from '../components/TodoTable';
import LoginCard from '../components/LoginCard';
import { useTodos, useDeleteTodo } from '../hooks/useTodos';
import { useAuth } from '../hooks/useAuth';
import { usePageTitle } from '../hooks/usePageTitle';
import { Navigate } from 'react-router-dom';

export default function Home() {
  usePageTitle('Home');
  const [showAddTodoForm, setShowAddTodoForm] = useState(false);
  const [page, setPage] = useState(1);
  const [limit] = useState(25);

  // 🔐 Auth state
  const { data: authData, isLoading: authLoading } = useAuth();
  const user = authData?.user ?? null;

  // 📝 Todos (only when logged in)
  const {
    data: todosData,
    isLoading,
    isError,
  } = useTodos(!!user, page, limit);

  const todos = todosData?.todos ?? [];

  // ✅ SAFE META
  const { total = 0 } = todosData?.meta ?? {};

  const deleteTodoMutation = useDeleteTodo();

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

  // ⏳ Checking auth
  if (authLoading) {
    return <div className="container mt-5">Checking login…</div>;
  }


  // ...

  // 🔐 NOT LOGGED IN UI
  if (!user) {
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
      <div className="card">
        <div className="card-header d-flex justify-content-between align-items-center">
          <h5 className="mb-0">My Todos</h5>

          <button
            type="button"
            className="btn btn-sm btn-primary"
            onClick={() => setShowAddTodoForm((prev) => !prev)}
          >
            {showAddTodoForm ? 'Close' : 'New Todo'}
          </button>
        </div>

        <div className="card-body">
          <div className="d-flex justify-content-between mb-3 align-items-center">
            <div>
              <strong>Total:</strong> {total}
            </div>

            <div className="d-flex gap-2 align-items-center">
              <button
                type="button"
                className="btn btn-sm btn-outline-secondary"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
              >
                Prev
              </button>

              <div>Page {page}</div>

              <button
                type="button"
                className="btn btn-sm btn-outline-secondary"
                onClick={() => setPage((p) => p + 1)}
                disabled={(page * limit) >= total}
              >
                Next
              </button>
            </div>
          </div>
          {showAddTodoForm && (
            <div className="mt-4">
              <NewTodoForm onSuccess={() => setPage(1)} />
            </div>
          )}

          <TodoTable
            todos={todos}
            deleteTodo={handleDeleteTodo}
            isDeleting={deleteTodoMutation.isPending}
          />
        </div>
      </div>
    </div>
  );
}