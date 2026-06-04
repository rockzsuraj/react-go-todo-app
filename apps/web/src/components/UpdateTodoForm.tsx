import { useState } from 'react';
import { useUpdateTodo } from '../hooks/useTodos';
import type { TodoEditable } from '../types/todo';

interface Props {
  todo: TodoEditable;
  onCancel: () => void;
}

function UpdateTodoForm({ todo, onCancel }: Props) {
  const [description, setDescription] = useState(todo.description);
  const [assigned, setAssigned] = useState(todo.assigned_to_name);
  const updateTodoMutation = useUpdateTodo();

  function handleChangeDescription(event: { target: { value: string } }) {
    setDescription(event.target.value);
  }

  function handleChangeAssigned(event: { target: { value: string } }) {
    setAssigned(event.target.value);
  }

  function submitUpdate() {
    if (description !== '' && assigned !== '') {
      updateTodoMutation.mutate(
        {
          id: todo.id,
          payload: {
            description,
            assigned_to_name: assigned,
            completed: todo.completed,
          },
        },
        {
          onSuccess: () => {
            onCancel();
          },
        },
      );
    }
  }

  return (
    <div className="todo-edit-form">
      <div className="todo-form-grid">
        <div className="todo-form-field">
          <label htmlFor="update-assigned">
            <i className="bi bi-person" />
            Assigned to
          </label>
          <input
            id="update-assigned"
            value={assigned}
            type="text"
            required
            onChange={handleChangeAssigned}
          />
        </div>
        <div className="todo-form-field todo-form-field--wide">
          <label htmlFor="update-description">
            <i className="bi bi-text-left" />
            Task details
          </label>
          <textarea
            id="update-description"
            value={description}
            rows={3}
            required
            onChange={handleChangeDescription}
          />
        </div>
      </div>
      <div className="todo-form-actions">
        <button
          onClick={submitUpdate}
          type="button"
          className="todo-save-action"
          disabled={updateTodoMutation.isPending}
        >
          <i className="bi bi-check2" />
          {updateTodoMutation.isPending ? 'Saving...' : 'Save changes'}
        </button>
        <button onClick={onCancel} type="button" className="todo-cancel-action">
          Cancel
        </button>
      </div>
    </div>
  );
}

export default UpdateTodoForm;
