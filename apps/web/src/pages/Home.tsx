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
    return <div className="container mt-5">Loading todos...</div>;
  }

  // ❌ Error
  if (isError) {
    return (
      <div className="container mt-5 text-danger">Failed to load todos</div>
    );
  }

  // ✅ LOGGED IN UI
  return (
    <div className="container mt-5">
      <div className="card glass-panel border-0">
        <div className="card-header premium-header text-white">
          <div className="d-flex justify-content-between align-items-center">
            <h5 className="mb-0 fw-semibold d-flex align-items-center gap-2">
              <i className="bi bi-list-task fs-4"></i>
              My Todos
            </h5>

            <button
              type="button"
              className="btn btn-light btn-sm fw-semibold rounded-pill px-3 py-1.5 shadow-sm"
              onClick={() => setShowAddTodoForm((prev) => !prev)}
            >
              <i
                className={`bi ${showAddTodoForm ? 'bi-x-lg' : 'bi-plus-lg'} me-1`}
              ></i>
              {showAddTodoForm ? 'Close' : 'New Todo'}
            </button>
          </div>
        </div>

        <div className="card-body p-4">
          {/* New Todo Form */}
          {showAddTodoForm && (
            <div className="mb-4 p-3 bg-light rounded-4 border border-light-subtle animate-hover">
              <NewTodoForm onSuccess={() => setShowAddTodoForm(false)} />
            </div>
          )}

          {/* Modern Filters Section */}
          <div className="mb-5 p-4 rounded-4 bg-light bg-opacity-50 border border-light-subtle">
            <div className="d-flex justify-content-between align-items-center mb-4">
              <h6 className="filter-section-title mb-0">
                <i className="bi bi-funnel-fill me-2 text-indigo"></i>
                Filter & Sort Tasks
              </h6>

              {(filterCompleted !== undefined || searchInput !== '') && (
                <button
                  type="button"
                  className="btn-reset-filters"
                  onClick={clearFilters}
                >
                  <i className="bi bi-arrow-counterclockwise"></i>
                  Reset Filters
                </button>
              )}
            </div>

            <div className="row g-4 align-items-end">
              {/* Status Tabs */}
              <div className="col-lg-4 col-md-6">
                <span className="form-label fw-semibold text-secondary small mb-2 d-block">
                  Task Status
                </span>
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
              <div className="col-lg-4 col-md-6">
                <label
                  className="form-label fw-semibold text-secondary small mb-2"
                  htmlFor="assignee-search"
                >
                  Assigned To
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
              <div className="col-lg-4 col-md-12">
                <label
                  className="form-label fw-semibold text-secondary small mb-2"
                  htmlFor="sort-order-select"
                >
                  Sort Order
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

            {/* Results & Summary */}
            <div className="d-flex justify-content-between align-items-center mt-4 pt-3 border-top border-light-subtle">
              <div className="results-info-badge">
                <i className="bi bi-info-circle-fill me-2"></i>
                Total Tasks: {total}
              </div>

              {total > limit && (
                <div className="text-secondary small">
                  Page <strong>{page}</strong> of {Math.ceil(total / limit)}
                </div>
              )}
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
