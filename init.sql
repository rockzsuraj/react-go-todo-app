CREATE TABLE IF NOT EXISTS todos (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    assigned TEXT
);

-- Insert sample data
INSERT INTO todos (description, assigned) VALUES 
('Setup project', 'John'),
('Create API', 'Jane'),
('Build frontend', 'Bob')
ON CONFLICT DO NOTHING;