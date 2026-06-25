import { useState } from 'react';
import { useUpdateTodo } from '../hooks/useTodos';
import type { TodoEditable } from '../types/todo';
import { Button, Input, TextArea } from './ui';

interface Props {
  todo: TodoEditable;
  onCancel: () => void;
}

function UpdateTodoForm({ todo, onCancel }: Props) {
  const [description, setDescription] = useState(todo.description);
  const [assigned, setAssigned] = useState(todo.assigned_to_name);
  const updateTodoMutation = useUpdateTodo();

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
          <Input
            id="update-assigned"
            label="Assigned to"
            value={assigned}
            type="text"
            required
            leftIcon={<i className="bi bi-person" />}
            onChange={(e) => setAssigned(e.target.value)}
          />
        </div>
        <div className="todo-form-field todo-form-field--wide">
          <TextArea
            id="update-description"
            label="Task details"
            value={description}
            rows={3}
            required
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>
      </div>
      <div className="todo-form-actions">
        <Button
          onClick={submitUpdate}
          type="button"
          isLoading={updateTodoMutation.isPending}
          leftIcon={<i className="bi bi-check2" />}
        >
          Save changes
        </Button>
        <Button onClick={onCancel} type="button" variant="secondary">
          Cancel
        </Button>
      </div>
    </div>
  );
}

export default UpdateTodoForm;

