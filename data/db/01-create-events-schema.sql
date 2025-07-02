-- Create a schema for events
CREATE SCHEMA IF NOT EXISTS events;

-- Create a dedicated user for events schema
CREATE USER events_user WITH PASSWORD 'events_password';

-- Create a table for events in the events schema
CREATE TABLE IF NOT EXISTS events.events (
    id VARCHAR(36) PRIMARY KEY,
    date TIMESTAMP NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(255)
);

-- Grant privileges to the events user
GRANT ALL PRIVILEGES ON SCHEMA events TO events_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA events TO events_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA events GRANT ALL ON TABLES TO events_user;

-- Insert sample events into the events schema
INSERT INTO events.events (id, date, title, description, location)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000', '2023-12-15 18:00:00', 'Annual Soccer Tournament', 'Join us for the annual soccer tournament with teams from all over the region.', 'Central Park Stadium'),
    ('550e8400-e29b-41d4-a716-446655440001', '2023-12-20 09:00:00', 'Morning Yoga Session', 'Start your day with a refreshing yoga session suitable for all experience levels.', 'Community Center'),
    ('550e8400-e29b-41d4-a716-446655440002', '2024-01-05 14:30:00', 'Basketball Workshop', 'Learn basketball fundamentals from professional coaches in this interactive workshop.', 'Downtown Sports Complex'),
    ('550e8400-e29b-41d4-a716-446655440003', '2024-01-10 19:00:00', 'Swimming Competition', 'Annual swimming competition with categories for all age groups.', 'Olympic Pool'),
    ('550e8400-e29b-41d4-a716-446655440004', '2024-01-15 16:00:00', 'Tennis Tournament', 'Singles and doubles tennis tournament with prizes for winners.', 'Tennis Club'),
    ('550e8400-e29b-41d4-a716-446655440005', '2024-02-01 10:00:00', 'Marathon Preparation Workshop', 'Get tips and training advice for upcoming marathon events.', 'Running Track'),
    ('550e8400-e29b-41d4-a716-446655440007', '2024-02-20 15:00:00', 'Kids Sports Day', 'Fun sports activities for children aged 5-12 years.', 'Elementary School Grounds'),
    ('550e8400-e29b-41d4-a716-446655440008', '2024-03-05 13:00:00', 'Cycling Tour', 'Scenic cycling tour through the countryside, 30km route.', 'Bike Shop Starting Point'),
    ('550e8400-e29b-41d4-a716-446655440009', '2024-03-15 11:00:00', 'Fitness Challenge', 'Test your fitness with various challenges and compete for prizes.', 'Fitness Center');

-- Create a dedicated user for keycloak schema
CREATE USER keycloak_user WITH PASSWORD 'keycloak_password';

-- Since we can't easily switch databases in a single initialization script,
-- we'll create the keycloak schema in the events_demo database
-- Keycloak will be configured to use this schema in the bitnami_keycloak database

-- Create schema for keycloak
CREATE SCHEMA IF NOT EXISTS keycloak;

-- Grant privileges to the keycloak user
GRANT ALL PRIVILEGES ON SCHEMA keycloak TO keycloak_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA keycloak TO keycloak_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA keycloak GRANT ALL ON TABLES TO keycloak_user;
