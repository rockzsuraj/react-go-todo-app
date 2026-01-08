import { useState } from 'react';
import UpdateTodoForm from './UpdateTodoForm';
import { Todo } from '../api/supabase';

interface Props {
  rowNumber: number;
  rowDescription: string;
  rowAssigned: string;
  deleteTodo: (id: number) => void;
}

function TodoRowItem(props: Props) {
  const { rowNumber, rowDescription, rowAssigned, deleteTodo } = props;
  const [isEditing, setIsEditing] = useState(false);
  
  const todo: Todo = {
    id: rowNumber,
    description: rowDescription,
    assigned: rowAssigned,
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
              onClick={() => setIsEditing(!isEditing)}
            >
              {isEditing ? 'Cancel' : 'Edit'}
            </button>
            <button
              type="button"
              onClick={() => deleteTodo(rowNumber)}
              className="btn btn-danger btn-sm"
            >
              X
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
