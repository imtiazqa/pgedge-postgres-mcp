-- Example test data for MCP server tests

-- Insert test users
INSERT INTO test_users (username, email) VALUES
    ('alice', 'alice@example.com'),
    ('bob', 'bob@example.com'),
    ('charlie', 'charlie@example.com'),
    ('diana', 'diana@example.com'),
    ('eve', 'eve@example.com')
ON CONFLICT (username) DO NOTHING;

-- Insert test documents
INSERT INTO test_documents (title, content, user_id) VALUES
    ('Introduction to PostgreSQL', 'PostgreSQL is a powerful open-source relational database...', 1),
    ('MCP Protocol Guide', 'The Model Context Protocol (MCP) provides a standardized way...', 1),
    ('Vector Search Basics', 'Vector similarity search enables semantic search capabilities...', 2),
    ('Database Optimization', 'Query optimization is crucial for database performance...', 2),
    ('Testing Best Practices', 'Comprehensive testing ensures code quality and reliability...', 3)
ON CONFLICT DO NOTHING;

-- Insert test embeddings
INSERT INTO test_embeddings (document_id, embedding_text, vector_data) VALUES
    (1, 'database relational postgresql', decode('00000001', 'hex')),
    (2, 'protocol mcp model context', decode('00000002', 'hex')),
    (3, 'vector search semantic similarity', decode('00000003', 'hex')),
    (4, 'database optimization performance', decode('00000004', 'hex')),
    (5, 'testing quality reliability', decode('00000005', 'hex'))
ON CONFLICT DO NOTHING;
