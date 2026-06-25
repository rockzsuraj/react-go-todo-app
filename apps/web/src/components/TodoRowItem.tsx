import { useState } from 'react';
import type { Todo } from '../types/todo';
import UpdateTodoForm from './UpdateTodoForm';
import { Badge } from './ui/Badge';
import { Button } from './ui/Button';

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
            <Badge variant={todo.completed ? 'success' : 'info'}>
              {todo.completed ? 'Completed' : 'Active'}
            </Badge>
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
        <Button
          variant="secondary"
          size="sm"
          disabled={isDeleting}
          onClick={() => setIsEditing((v) => !v)}
          leftIcon={<i className={`bi ${isEditing ? 'bi-x-lg' : 'bi-pencil'}`} />}
          aria-label={isEditing ? 'Close edit form' : 'Edit task'}
        >
          {isEditing ? 'Close' : 'Edit'}
        </Button>
        <Button
          variant="danger"
          size="sm"
          onClick={deleteTodo}
          isLoading={isDeleting}
          leftIcon={<i className="bi bi-trash3" />}
          aria-label="Delete task"
        >
          Delete
        </Button>
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
