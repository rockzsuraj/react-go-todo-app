import React from "react";
import TodoRowItem from "./TodoRowItem";
import { Todo } from "../types/todo";

interface Props {
    todos: Todo[],
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
                {
                    props.todos.map(todo =>
                        <TodoRowItem
                            key={todo.rowNumber}
                            rowNumber={todo.rowNumber}
                            rowDescription={todo.rowDescription}
                            rowAssigned={todo.rowAssigned}
                            deleteTodo={props.deleteTodo}
                        />
                    )
                }
            </tbody>
        </table>
    )
}

export default TodoTable;