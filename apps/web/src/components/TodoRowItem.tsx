import { useState } from 'react';
import UpdateTodoForm from './UpdateTodoForm';
import type { TodoEditable } from '../types/todo';

interface Props {
  rowNumber: number;
  rowDescription: string;
  rowAssigned: string;
  deleteTodo: (id: number) => void;
  isDeleting?: boolean; // ✅ add
}

function TodoRowItem({
  rowNumber,
  rowDescription,
  rowAssigned,
  deleteTodo,
  isDeleting = false,
}: Props) {
  const [isEditing, setIsEditing] = useState(false);

  const todo: TodoEditable = {
    id: rowNumber,
    description: rowDescription,
    assigned_to_name: rowAssigned,
  };

  return (
    <>
      <tr>
        <th scope="row">{rowNumber}</th>
        <td>{rowDescription}</td>
        <td>{rowAssigned}</td>
        <td>
          <div className="d-flex gap-2">
            <button
              type="button"
              className="btn btn-outline-primary btn-sm"
              disabled={isDeleting}
              onClick={() => setIsEditing((v) => !v)}
            >
              {isEditing ? 'Cancel' : 'Edit'}
            </button>

            <button
              type="button"
              onClick={() => deleteTodo(rowNumber)}
              className="btn btn-danger btn-sm"
              disabled={isDeleting}
            >
              {isDeleting ? '...' : 'X'}
            </button>
          </div>
        </td>
      </tr>

      {isEditing && (
        <tr>
          <td colSpan={4}>
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