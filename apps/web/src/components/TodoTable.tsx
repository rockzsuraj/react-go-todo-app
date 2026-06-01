import TodoRowItem from './TodoRowItem';
import type { Todo } from '../types/todo';

interface Props {
  todos: Todo[];
  deleteTodo: (id: number) => void;
  toggleTodoCompleted: (todo: Todo) => void;
  isDeleting?: boolean;
  sortBy?: string;
  sortOrder?: 'ASC' | 'DESC';
  onSort?: (field: string) => void;
}

function TodoTable({
  todos,
  deleteTodo,
  toggleTodoCompleted,
  isDeleting = false,
  sortBy,
  sortOrder,
  onSort
}: Props) {
  const getSortIcon = (field: string) => {
    if (sortBy !== field) return 'bi bi-arrow-down-up';
    return sortOrder === 'ASC' ? 'bi bi-arrow-up' : 'bi bi-arrow-down';
  };

  if (todos.length === 0) {
    return (
      <div className="text-center py-5">
        <div className="mb-3">
          <i className="bi bi-inbox display-1 text-muted"></i>
        </div>
        <h5 className="text-muted">No todos found</h5>
        <p className="text-muted">
          {isDeleting ? 'Deleting...' : 'Create your first todo to get started!'}
        </p>
      </div>
    );
  }

  return (
    <div className="table-responsive">
      <table className="table table-hover align-middle">
        <thead className="table-light">
          <tr>
            <th className="border-0" style={{ width: '40px' }}></th>
            <th
              className="border-0 fw-semibold text-primary"
              onClick={() => onSort?.('description')}
              style={{ cursor: onSort ? 'pointer' : 'default' }}
            >
              <i className="bi bi-text-paragraph me-1"></i>
              Description
              {onSort && <i className={`ms-1 ${getSortIcon('description')}`}></i>}
            </th>
            <th
              className="border-0 fw-semibold text-primary"
              onClick={() => onSort?.('assigned_to_name')}
              style={{ cursor: onSort ? 'pointer' : 'default' }}
            >
              <i className="bi bi-person me-1"></i>
              Assigned To
              {onSort && <i className={`ms-1 ${getSortIcon('assigned_to_name')}`}></i>}
            </th>
            <th
              className="border-0 fw-semibold text-primary"
              onClick={() => onSort?.('created_at')}
              style={{ cursor: onSort ? 'pointer' : 'default' }}
            >
              <i className="bi bi-calendar-plus me-1"></i>
              Created
              {onSort && <i className={`ms-1 ${getSortIcon('created_at')}`}></i>}
            </th>
            <th className="border-0 fw-semibold text-primary">Actions</th>
          </tr>
        </thead>
        <tbody>
          {todos.map((todo) => (
            <TodoRowItem
              key={todo.id}
              todo={todo}
              deleteTodo={() => deleteTodo(todo.id)}
              toggleTodoCompleted={() => toggleTodoCompleted(todo)}
              isDeleting={isDeleting}
            />
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default TodoTable;