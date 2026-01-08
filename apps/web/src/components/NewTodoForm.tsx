import { useState } from 'react';
import { useCreateTodo } from '../hooks/useTodos';

function NewTodoForm() {
  const [description, setDescription] = useState('');
  const [assigned, setAssigned] = useState('');
  const createTodoMutation = useCreateTodo();

  function handleChangeDescription(event: { target: { value: string } }) {
    setDescription(event.target.value);
  }

  function handleChangeAssigned(event: { target: { value: string } }) {
    setAssigned(event.target.value);
  }

  function submitTodo() {
    if (description !== '' && assigned !== '') {
      createTodoMutation.mutate(
        { description, assigned },
        {
          onSuccess: () => {
            setAssigned('');
            setDescription('');
          },
        }
      );
    }
  }

  return (
    <div className="mt-5">
      <form onSubmit={submitTodo}>
        <div className="mb-3">
          <label htmlFor="assigned" className="form-label">
            Assigned
          </label>
          <input
            id="assigned"
            value={assigned}
            type="text"
            className="form-control"
            required
            onChange={handleChangeAssigned}
          />
        </div>
        <div className="mb-3">
          <label htmlFor="description" className="form-label">
            Description
          </label>
          <textarea
            id="description"
            value={description}
            rows={3}
            className="form-control"
            required
            onChange={handleChangeDescription}
          />
        </div>
        <button
          onClick={submitTodo}
          type="button"
          className="btn btn-primary mt-3"
          disabled={createTodoMutation.isPending}
        >
          {createTodoMutation.isPending ? 'Adding...' : 'Add Todo'}
        </button>
      </form>
    </div>
  );
}

export default NewTodoForm;
