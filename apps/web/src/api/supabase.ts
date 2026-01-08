import { createClient } from '@supabase/supabase-js';

const supabaseUrl = 'https://qnlhgaymddnazecbtsau.supabase.co';
const supabaseKey = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InFubGhnYXltZGRuYXplY2J0c2F1Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3MzY0MzE0NzQsImV4cCI6MjA1MjAwNzQ3NH0.sb_publishable_kw655WpHOEH67cfb75d09g__KVNNChK';

export const supabase = createClient(supabaseUrl, supabaseKey);

export interface Todo {
  id: number;
  description: string;
  assigned: string;
  created_at?: string;
}

export const todoApi = {
  getAll: async () => {
    const { data, error } = await supabase
      .from('todos')
      .select('*')
      .order('id', { ascending: true });
    
    if (error) throw error;
    return { data: { success: true, data } };
  },

  create: async (todo: Omit<Todo, 'id'>) => {
    const { data, error } = await supabase
      .from('todos')
      .insert([todo])
      .select();
    
    if (error) throw error;
    return { data: { success: true, data } };
  },

  update: async (id: number, todo: Omit<Todo, 'id'>) => {
    const { data, error } = await supabase
      .from('todos')
      .update(todo)
      .eq('id', id)
      .select();
    
    if (error) throw error;
    return { data: { success: true, data } };
  },

  delete: async (id: number) => {
    const { error } = await supabase
      .from('todos')
      .delete()
      .eq('id', id);
    
    if (error) throw error;
    return { data: { success: true } };
  },
};