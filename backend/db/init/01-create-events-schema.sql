-- Create a schema for events backup
CREATE SCHEMA IF NOT EXISTS events;

-- Create a table for events in the events schema
CREATE TABLE IF NOT EXISTS events.events (
    id VARCHAR(36) PRIMARY KEY,
    date TIMESTAMP NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(255)
);

-- Grant privileges to the Keycloak user (since that's the user we're using)
GRANT ALL PRIVILEGES ON SCHEMA events TO bn_keycloak;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA events TO bn_keycloak;
