import { useState } from 'react';
import { useCreateTodo } from '../hooks/useTodos';
import { Button } from './ui/Button';
import { Input } from './ui/Input';
import { TextArea } from './ui/TextArea';

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
          <Input
            id="assigned"
            label="Assigned to"
            type="text"
            value={assignedToName}
            required
            onChange={(e) => setAssignedToName(e.target.value)}
            placeholder="e.g. Mom, Dad, John"
            leftIcon={<i className="bi bi-person" />}
          />
        </div>

        <div className="todo-form-field todo-form-field--wide">
          <TextArea
            id="description"
            label="Task details"
            rows={3}
            value={description}
            required
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Describe what needs to be done..."
          />
        </div>
      </div>

      <div className="todo-form-actions">
        <Button
          type="submit"
          isLoading={createTodoMutation.isPending}
          leftIcon={<i className="bi bi-plus-lg" />}
        >
          Add task
        </Button>
      </div>
    </form>
  );
}

export default NewTodoForm;
