export interface Todo {
  id: number;
  description: string;
  assigned_to_name: string;
  completed: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateTodoInput {
  description: string;
  assigned_to_name: string;
}

export type TodoEditable = Pick<Todo, 'id' | 'description' | 'assigned_to_name'>;