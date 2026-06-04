import { useEffect, useState } from 'react';
import NewTodoForm from '../components/NewTodoForm';
import TodoTable from '../components/TodoTable';
import { usePageTitle } from '../hooks/usePageTitle';
import {
  useDeleteTodo,
  useTodos,
  useToggleTodoCompleted,
} from '../hooks/useTodos';

export default function Home() {
  usePageTitle('Home');
  const [showAddTodoForm, setShowAddTodoForm] = useState(false);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [sortBy, setSortBy] = useState<string>('created_at');
  const [sortOrder, setSortOrder] = useState<'ASC' | 'DESC'>('DESC');
  const [filterCompleted, setFilterCompleted] = useState<boolean | undefined>(
    undefined,
  );
  const [filterAssigned, setFilterAssigned] = useState<string>('');
  const [searchInput, setSearchInput] = useState<string>('');

  // Debounce search assignee input to prevent fetching on every keypress
  useEffect(() => {
    if (searchInput === filterAssigned) return;

    const handler = setTimeout(() => {
      setFilterAssigned(searchInput);
      setPage(1);
    }, 300);

    return () => clearTimeout(handler);
  }, [searchInput, filterAssigned]);

  // 📝 Todos
  const {
    data: todosData,
    isLoading,
    isError,
    isFetching,
  } = useTodos(
    true,
    page,
    limit,
    sortBy,
    sortOrder,
    filterCompleted,
    filterAssigned,
  );

  const todos = todosData?.todos ?? [];

  // ✅ SAFE META
  const { total = 0 } = todosData?.meta ?? {};
  const totalPages = Math.max(1, Math.ceil(total / limit));
  const completedCount = todos.filter((todo) => todo.completed).length;
  const activeCount = todos.length - completedCount;

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

  const handleToggleTodoCompleted = (todo: {
    id: number;
    description: string;
    assigned_to_name: string;
    completed: boolean;
  }) => {
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
    setSearchInput('');
    setFilterAssigned('');
    setPage(1);
  };
  // ⏳ Todos loading
  if (isLoading) {
    return (
      <div className="todo-page-shell">
        <div className="todo-loading-state">
          <span className="spinner-border text-primary" aria-hidden="true" />
          <span>Loading your tasks...</span>
        </div>
      </div>
    );
  }

  // ❌ Error
  if (isError) {
    return (
      <div className="todo-page-shell">
        <div className="todo-error-state">
          <i className="bi bi-exclamation-triangle" />
          <div>
            <strong>Could not load your tasks</strong>
            <span>Please refresh the page and try again.</span>
          </div>
        </div>
      </div>
    );
  }

  // ✅ LOGGED IN UI
  return (
    <div className="todo-page-shell">
      <section className="todo-hero">
        <div>
          <span className="todo-eyebrow">Personal workspace</span>
          <h1>My tasks</h1>
          <p>Keep work moving and make today feel manageable.</p>
        </div>
        <button
          type="button"
          className="todo-primary-action"
          onClick={() => setShowAddTodoForm((prev) => !prev)}
          aria-expanded={showAddTodoForm}
        >
          <i className={`bi ${showAddTodoForm ? 'bi-x-lg' : 'bi-plus-lg'}`} />
          {showAddTodoForm ? 'Close form' : 'Add task'}
        </button>
      </section>

      <section className="todo-stats" aria-label="Task summary">
        <div className="todo-stat-card todo-stat-card--total">
          <span className="todo-stat-icon">
            <i className="bi bi-collection" />
          </span>
          <div>
            <strong>{total}</strong>
            <span>Total tasks</span>
          </div>
        </div>
        <div className="todo-stat-card todo-stat-card--active">
          <span className="todo-stat-icon">
            <i className="bi bi-lightning-charge" />
          </span>
          <div>
            <strong>{activeCount}</strong>
            <span>Active on this page</span>
          </div>
        </div>
        <div className="todo-stat-card todo-stat-card--done">
          <span className="todo-stat-icon">
            <i className="bi bi-check2-circle" />
          </span>
          <div>
            <strong>{completedCount}</strong>
            <span>Done on this page</span>
          </div>
        </div>
      </section>

      {showAddTodoForm && (
        <section className="todo-create-panel">
          <div className="todo-section-heading">
            <div>
              <span className="todo-eyebrow">New task</span>
              <h2>What needs to get done?</h2>
            </div>
          </div>
          <NewTodoForm onSuccess={() => setShowAddTodoForm(false)} />
        </section>
      )}

      <section className="todo-workspace">
        <div className="todo-toolbar">
          <div className="todo-toolbar-heading">
            <div>
              <span className="todo-eyebrow">Task list</span>
              <h2>Stay on top of it</h2>
            </div>

            {(filterCompleted !== undefined || searchInput !== '') && (
              <button
                type="button"
                className="btn-reset-filters"
                onClick={clearFilters}
              >
                <i className="bi bi-arrow-counterclockwise" />
                Reset
              </button>
            )}
          </div>

          <div className="todo-filter-grid">
            {/* Status Tabs */}
            <div className="todo-filter-field">
              <span className="todo-filter-label">Status</span>
              <div className="segmented-control">
                <button
                  type="button"
                  className={`segmented-pill ${filterCompleted === undefined ? 'active' : ''}`}
                  onClick={() => {
                    setFilterCompleted(undefined);
                    handleFilterChange();
                  }}
                >
                  All
                </button>
                <button
                  type="button"
                  className={`segmented-pill ${filterCompleted === false ? 'active' : ''}`}
                  onClick={() => {
                    setFilterCompleted(false);
                    handleFilterChange();
                  }}
                >
                  Active
                </button>
                <button
                  type="button"
                  className={`segmented-pill ${filterCompleted === true ? 'active' : ''}`}
                  onClick={() => {
                    setFilterCompleted(true);
                    handleFilterChange();
                  }}
                >
                  Completed
                </button>
              </div>
            </div>

            {/* Search Assignee */}
            <div className="todo-filter-field">
              <label className="todo-filter-label" htmlFor="assignee-search">
                Assignee
              </label>
              <div className="modern-input-group">
                <input
                  id="assignee-search"
                  type="text"
                  className="form-control modern-input"
                  placeholder="Search by assignee name..."
                  value={searchInput}
                  onChange={(e) => {
                    setSearchInput(e.target.value);
                  }}
                />
                {isFetching ? (
                  <output
                    className="spinner-border spinner-border-sm text-indigo modern-input-icon"
                    style={{ borderWidth: '2px' }}
                  >
                    <span className="visually-hidden">Loading...</span>
                  </output>
                ) : (
                  <i className="bi bi-search modern-input-icon"></i>
                )}
              </div>
            </div>

            {/* Sort Controls */}
            <div className="todo-filter-field">
              <label className="todo-filter-label" htmlFor="sort-order-select">
                Sort by
              </label>
              <div className="d-flex gap-2">
                <div className="modern-input-group flex-grow-1">
                  <select
                    id="sort-order-select"
                    className="form-select modern-input modern-select"
                    value={sortBy}
                    onChange={(e) => handleSort(e.target.value)}
                  >
                    <option value="created_at">Created Date</option>
                    <option value="updated_at">Updated Date</option>
                    <option value="description">Description</option>
                    <option value="assigned_to_name">Assigned To</option>
                  </select>
                  <i className="bi bi-sort-down modern-input-icon"></i>
                </div>

                <button
                  type="button"
                  className="sort-direction-btn"
                  title={
                    sortOrder === 'ASC' ? 'Sort Ascending' : 'Sort Descending'
                  }
                  onClick={() => {
                    setSortOrder(sortOrder === 'ASC' ? 'DESC' : 'ASC');
                    setPage(1);
                  }}
                >
                  <i
                    className={`bi ${sortOrder === 'ASC' ? 'bi-sort-alpha-up' : 'bi-sort-alpha-down'}`}
                  ></i>
                </button>
              </div>
            </div>
          </div>
        </div>

        <div className="todo-list-region">
          <TodoTable
            todos={todos}
            deleteTodo={handleDeleteTodo}
            toggleTodoCompleted={handleToggleTodoCompleted}
            isDeleting={deleteTodoMutation.isPending}
          />
        </div>

        {totalPages > 1 && (
          <nav className="todo-pagination" aria-label="Task pages">
            <button
              type="button"
              onClick={() => setPage((current) => Math.max(1, current - 1))}
              disabled={page === 1 || isFetching}
            >
              <i className="bi bi-arrow-left" />
              Previous
            </button>
            <span>
              Page <strong>{page}</strong> of <strong>{totalPages}</strong>
            </span>
            <button
              type="button"
              onClick={() =>
                setPage((current) => Math.min(totalPages, current + 1))
              }
              disabled={page === totalPages || isFetching}
            >
              Next
              <i className="bi bi-arrow-right" />
            </button>
          </nav>
        )}
      </section>
    </div>
  );
}
