import type { Todo } from '../types/todo';
import TodoRowItem from './TodoRowItem';
import { EmptyState } from './ui';

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
      <EmptyState
        icon="bi-collection"
        title="No tasks found"
        description={
          isDeleting
            ? 'Deleting...'
            : 'Try a different filter, or add a new task to get started.'
        }
      />
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

