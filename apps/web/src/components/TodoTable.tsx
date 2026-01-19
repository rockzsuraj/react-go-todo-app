import TodoRowItem from './TodoRowItem';
import type { Todo } from '../types/todo';

interface Props {
  todos: Todo[];
  deleteTodo: (id: number) => void;
  isDeleting?: boolean;
}

function TodoTable({ todos, deleteTodo, isDeleting = false }: Props) {
  return (
    <table className="table table-hover">
      <thead>
        <tr>
          <th>#</th>
          <th>Description</th>
          <th>Assigned</th>
          <th>Delete</th>
        </tr>
      </thead>

      <tbody>
        {todos.map((todo) => (
          <TodoRowItem
            key={todo.id}
            rowNumber={todo.id}
            rowDescription={todo.description}
            rowAssigned={todo.assigned_to_name}
            deleteTodo={deleteTodo}
            isDeleting={isDeleting} // ✅ now valid
          />
        ))}
      </tbody>
    </table>
  );
}

export default TodoTable;