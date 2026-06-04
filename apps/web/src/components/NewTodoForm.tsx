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
    <div className="mt-5">
      <form onSubmit={handleSubmit}>
        <div className="mb-3">
          <label htmlFor="assigned" className="form-label">
            Assigned To
          </label>
          <input
            id="assigned"
            type="text"
            className="form-control"
            value={assignedToName}
            required
            onChange={(e) => setAssignedToName(e.target.value)}
            placeholder="e.g. Mom, Dad, John"
          />
        </div>

        <div className="mb-3">
          <label htmlFor="description" className="form-label">
            Description
          </label>
          <textarea
            id="description"
            rows={3}
            className="form-control"
            value={description}
            required
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>

        <button
          type="submit"
          className="btn btn-primary mt-3"
          disabled={createTodoMutation.isPending}
        >
          {createTodoMutation.isPending ? 'Adding…' : 'Add Todo'}
        </button>
      </form>
    </div>
  );
}

export default NewTodoForm;
