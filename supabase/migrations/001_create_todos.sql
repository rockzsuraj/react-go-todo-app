-- Migration: Create todos table if not exists
-- Run this in Supabase SQL Editor

-- Drop table if exists (for clean setup)
DROP TABLE IF EXISTS todos;

-- Create todos table
CREATE TABLE todos (
  id SERIAL PRIMARY KEY,
  description TEXT NOT NULL,
  assigned TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable Row Level Security
ALTER TABLE todos ENABLE ROW LEVEL SECURITY;

-- Create policy to allow all operations (for demo purposes)
CREATE POLICY "Allow all operations on todos" ON todos
  FOR ALL USING (true) WITH CHECK (true);

-- Insert sample data
INSERT INTO todos (description, assigned) VALUES 
('Setup Supabase Database', 'Developer'),
('Deploy to GitHub Pages', 'DevOps'),
('Configure free hosting', 'Frontend Team'),
('Test production deployment', 'QA Team');

-- Grant permissions
GRANT ALL ON todos TO anon;
GRANT ALL ON todos TO authenticated;
GRANT USAGE ON SEQUENCE todos_id_seq TO anon;
GRANT USAGE ON SEQUENCE todos_id_seq TO authenticated;