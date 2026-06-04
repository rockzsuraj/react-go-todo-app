import { useState } from 'react';
import type { Todo } from '../types/todo';
import UpdateTodoForm from './UpdateTodoForm';

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
    <article
      className={`todo-item ${todo.completed ? 'todo-item--completed' : ''}`}
    >
      <div className="todo-item-main">
        <button
          type="button"
          className="todo-check-button"
          aria-label={
            todo.completed ? 'Mark task active' : 'Mark task complete'
          }
          aria-pressed={todo.completed}
          disabled={isDeleting}
          onClick={toggleTodoCompleted}
        >
          <i className={`bi ${todo.completed ? 'bi-check-lg' : ''}`} />
        </button>

        <div className="todo-item-content">
          <div className="todo-item-topline">
            <span
              className={`todo-status ${todo.completed ? 'is-done' : 'is-active'}`}
            >
              {todo.completed ? 'Completed' : 'Active'}
            </span>
            <time dateTime={todo.created_at}>
              <i className="bi bi-calendar3" />
              {new Date(todo.created_at).toLocaleDateString(undefined, {
                month: 'short',
                day: 'numeric',
                year: 'numeric',
              })}
            </time>
          </div>
          <h3>{todo.description}</h3>
          <div className="todo-assignee">
            <span className="todo-assignee-avatar">
              {todo.assigned_to_name.charAt(0).toUpperCase()}
            </span>
            <span>{todo.assigned_to_name}</span>
          </div>
        </div>
      </div>

      <div className="todo-item-actions">
        <button
          type="button"
          className="todo-icon-action"
          disabled={isDeleting}
          onClick={() => setIsEditing((v) => !v)}
          aria-label={isEditing ? 'Close edit form' : 'Edit task'}
        >
          <i className={`bi ${isEditing ? 'bi-x-lg' : 'bi-pencil'}`} />
          <span>{isEditing ? 'Close' : 'Edit'}</span>
        </button>
        <button
          type="button"
          onClick={deleteTodo}
          className="todo-icon-action todo-icon-action--danger"
          disabled={isDeleting}
          aria-label="Delete task"
        >
          <i
            className={`bi ${isDeleting ? 'bi-hourglass-split' : 'bi-trash3'}`}
          />
          <span>Delete</span>
        </button>
      </div>

      {isEditing && (
        <div className="todo-item-editor">
          <UpdateTodoForm todo={todo} onCancel={() => setIsEditing(false)} />
        </div>
      )}
    </article>
  );
}

export default TodoRowItem;
