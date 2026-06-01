import { useState } from 'react';
import UpdateTodoForm from './UpdateTodoForm';
import type { Todo } from '../types/todo';

interface Props {
  todo: Todo;
  deleteTodo: () => void;
  toggleTodoCompleted: () => void;
  isDeleting?: boolean;
}

function TodoRowItem({
  todo,
  deleteTodo,
  toggleTodoCompleted,
  isDeleting = false,
}: Props) {
  const [isEditing, setIsEditing] = useState(false);

  return (
    <>
      <tr className="align-middle">
        <td>
          <div className="form-check">
            <input
              type="checkbox"
              className="form-check-input"
              checked={todo.completed}
              disabled={isDeleting}
              onChange={toggleTodoCompleted}
            />
          </div>
        </td>
        <td>
          <div className={`fw-semibold ${todo.completed ? 'text-decoration-line-through text-muted' : ''}`}>
            {todo.description}
          </div>
        </td>
        <td>
          <div className="d-flex align-items-center gap-2">
            <div className="rounded-circle bg-primary text-white d-flex align-items-center justify-content-center" 
                 style={{ width: '32px', height: '32px', fontSize: '14px' }}>
              {todo.assigned_to_name.charAt(0).toUpperCase()}
            </div>
            <span>{todo.assigned_to_name}</span>
          </div>
        </td>
        <td>
          <small className="text-muted">
            {new Date(todo.created_at).toLocaleDateString()}
          </small>
        </td>
        <td>
          <div className="d-flex gap-2">
            <button
              type="button"
              className="btn btn-outline-primary btn-sm"
              disabled={isDeleting}
              onClick={() => setIsEditing((v) => !v)}
            >
              <i className={`bi ${isEditing ? 'bi-x-lg' : 'bi-pencil'}`}></i>
            </button>

            <button
              type="button"
              onClick={deleteTodo}
              className="btn btn-outline-danger btn-sm"
              disabled={isDeleting}
            >
              <i className={`bi ${isDeleting ? 'bi-hourglass-split' : 'bi-trash'}`}></i>
            </button>
          </div>
        </td>
      </tr>

      {isEditing && (
        <tr>
          <td colSpan={5} className="bg-light">
            <UpdateTodoForm
              todo={todo}
              onCancel={() => setIsEditing(false)}
            />
          </td>
        </tr>
      )}
    </>
  );
}

export default TodoRowItem;