-- Create todos table
CREATE TABLE IF NOT EXISTS todos (
  id SERIAL PRIMARY KEY,
  description TEXT NOT NULL,
  assigned TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert sample data
INSERT INTO todos (description, assigned) VALUES 
('Setup Database', 'Developer'),
('Deploy Backend', 'DevOps'),
('Test API', 'QA Team')
ON CONFLICT DO NOTHING;