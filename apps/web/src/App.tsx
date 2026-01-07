import { useState } from 'react';
import './App.css';
import TodoTable from './components/TodoTable';
import NewTodoForm from './components/NewTodoForm';
import React from 'react';

function App() {

  const [todos, setTodos] = useState([
    {
      rowNumber: 1,
      rowDescription: "Feed puppy",
      rowAssigned: 'User One'
    },
    {
      rowNumber: 2,
      rowDescription: "Water plants",
      rowAssigned: 'User Two'
    },
    {
      rowNumber: 3,
      rowDescription: "make dinner",
      rowAssigned: 'User Three'
    }
  ]);

  const [showAddTodoForm, setShowAddTodoForm] = useState(false);

  function handleAddTodo(description: string, assigned: string) {
    let rowNumber = 0;
    if (todos.length === 0) {
      rowNumber = 1;
    } else {
      rowNumber = todos[todos.length - 1].rowNumber + 1;
    }
    setTodos(prevTodo => [...prevTodo, {
      rowNumber: rowNumber,
      rowDescription: description,
      rowAssigned: assigned
    }]);
  }

  function deleteTodo(id: number) {
    setTodos(prevTodos => prevTodos.filter(todo => todo.rowNumber !== id))
  }

  return (
    <div className='mt-5 container'>
      <div className='card'>
        <div className='card-header'>
          Your Todo's
        </div>
        <div className='card-body'>
          <TodoTable todos={todos} deleteTodo={deleteTodo} />
          <button
            onClick={() => setShowAddTodoForm(!showAddTodoForm)}
            className='btn btn-primary'
          >{showAddTodoForm ? 'close New Todo' : 'New todo'}</button>
          {
            showAddTodoForm && <NewTodoForm handleAddTodo={handleAddTodo} />
          }
        </div>
      </div>
    </div>
  );
}

export default App;
