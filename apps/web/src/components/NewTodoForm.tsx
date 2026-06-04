import { useState } from 'react';
import { useCreateTodo } from '../hooks/useTodos';

function NewTodoForm({ onSuccess }: { onSuccess?: () => void }) {
  const [description, setDescription] = useState('');
  const [assignedToName, setAssignedToName] = useState('');

  const createTodoMutation = useCreateTodo();

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (!description.trim() || !assignedToName.trim()) return;

    createTodoMutation.mutate(
      {
        description,
        assigned_to_name: assignedToName,
      },
      {
        onSuccess: () => {
          setDescription('');
          setAssignedToName('');
          if (onSuccess) onSuccess();
        },
      },
    );
  }

  return (
    <form className="todo-form" onSubmit={handleSubmit}>
      <div className="todo-form-grid">
        <div className="todo-form-field">
          <label htmlFor="assigned">
            <i className="bi bi-person" />
            Assigned to
          </label>
          <input
            id="assigned"
            type="text"
            value={assignedToName}
            required
            onChange={(e) => setAssignedToName(e.target.value)}
            placeholder="e.g. Mom, Dad, John"
          />
        </div>

        <div className="todo-form-field todo-form-field--wide">
          <label htmlFor="description">
            <i className="bi bi-text-left" />
            Task details
          </label>
          <textarea
            id="description"
            rows={3}
            value={description}
            required
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Describe what needs to be done..."
          />
        </div>
      </div>

      <div className="todo-form-actions">
        <button
          type="submit"
          className="todo-primary-action"
          disabled={createTodoMutation.isPending}
        >
          <i className="bi bi-plus-lg" />
          {createTodoMutation.isPending ? 'Adding...' : 'Add task'}
        </button>
      </div>
    </form>
  );
}

export default NewTodoForm;
