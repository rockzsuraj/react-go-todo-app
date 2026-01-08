import { Todo } from '../api/supabase';
import TodoRowItem from './TodoRowItem';

interface Props {
  todos: Todo[];
  deleteTodo: (id: number) => void;
}

function TodoTable(props: Props) {
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
        {props.todos.map((todo) => (
          <TodoRowItem
            key={todo.id}
            rowNumber={todo.id}
            rowDescription={todo.description}
            rowAssigned={todo.assigned}
            deleteTodo={props.deleteTodo}
          />
        ))}
      </tbody>
    </table>
  );
}

export default TodoTable;
