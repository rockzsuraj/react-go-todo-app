import type { Todo } from '../types/todo';
import TodoRowItem from './TodoRowItem';

interface Props {
  todos: Todo[];
  deleteTodo: (id: number) => void;
  toggleTodoCompleted: (todo: Todo) => void;
  isDeleting?: boolean;
}

function TodoTable({
  todos,
  deleteTodo,
  toggleTodoCompleted,
  isDeleting = false,
}: Props) {
  if (todos.length === 0) {
    return (
      <div className="todo-empty-state">
        <span className="todo-empty-icon">
          <i className="bi bi-check2-circle" />
        </span>
        <h3>No tasks found</h3>
        <p>
          {isDeleting
            ? 'Deleting...'
            : 'Try another filter or add a new task to get started.'}
        </p>
      </div>
    );
  }

  return (
    <div className="todo-card-list">
      {todos.map((todo) => (
        <TodoRowItem
          key={todo.id}
          todo={todo}
          deleteTodo={() => deleteTodo(todo.id)}
          toggleTodoCompleted={() => toggleTodoCompleted(todo)}
          isDeleting={isDeleting}
        />
      ))}
    </div>
  );
}

export default TodoTable;
