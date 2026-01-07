import React from "react";

interface Props {
    rowNumber: TodoModel['rowNumber'],
    rowDescription: TodoModel['rowDescription'],
    rowAssigned: TodoModel['rowAssigned'],
    deleteTodo: Function
}

function TodoRowItem(props: Props) {
    const { rowNumber, rowDescription, rowAssigned, deleteTodo } = props;
    return (
        <tr>
            <th scope="row">{rowNumber}</th>
            <td>{rowDescription}</td>
            <td>{rowAssigned}</td>
            <td>
                <button
                    onClick={() => deleteTodo(rowNumber)}
                    className="btn btn-danger"
                >
                    X
                </button>
            </td>
        </tr>
    )
}

export default TodoRowItem;