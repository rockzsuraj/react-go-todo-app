import React, { useState } from 'react';
import './App.css';
import NewTodoForm from './components/NewTodoForm';
import TodoTable from './components/TodoTable';
import { useTodos, useDeleteTodo } from './hooks/useTodos';

function App() {
  const [showAddTodoForm, setShowAddTodoForm] = useState(false);
  const { data: todos = [], isLoading, error } = useTodos();
  const deleteTodoMutation = useDeleteTodo();

  const handleDeleteTodo = (id: number) => {
    deleteTodoMutation.mutate(id);
  };

  if (isLoading) return <div className="container mt-5">Loading...</div>;
  if (error) return <div className="container mt-5">Error loading todos</div>;

  return (
    <div className="mt-5 container">
      <div className="card">
        <div className="card-header">Your Todo's</div>
        <div className="card-body">
          <TodoTable todos={todos} deleteTodo={handleDeleteTodo} />
          <button
            type="button"
            onClick={() => setShowAddTodoForm(!showAddTodoForm)}
            className="btn btn-primary"
          >
            {showAddTodoForm ? 'close New Todo' : 'New todo'}
          </button>
          {showAddTodoForm && <NewTodoForm />}
        </div>
      </div>
    </div>
  );
}

export default App;
