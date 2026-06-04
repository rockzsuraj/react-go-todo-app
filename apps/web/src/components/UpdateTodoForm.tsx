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
    <div className="mt-3 p-3 border rounded">
      <h6>Update Todo</h6>
      <div className="mb-3">
        <label htmlFor="update-assigned" className="form-label">
          Assigned
        </label>
        <input
          id="update-assigned"
          value={assigned}
          type="text"
          className="form-control"
          required
          onChange={handleChangeAssigned}
        />
      </div>
      <div className="mb-3">
        <label htmlFor="update-description" className="form-label">
          Description
        </label>
        <textarea
          id="update-description"
          value={description}
          rows={3}
          className="form-control"
          required
          onChange={handleChangeDescription}
        />
      </div>
      <div className="d-flex gap-2">
        <button
          onClick={submitUpdate}
          type="button"
          className="btn btn-success"
          disabled={updateTodoMutation.isPending}
        >
          {updateTodoMutation.isPending ? 'Updating...' : 'Update'}
        </button>
        <button onClick={onCancel} type="button" className="btn btn-secondary">
          Cancel
        </button>
      </div>
    </div>
  );
}

export default UpdateTodoForm;
