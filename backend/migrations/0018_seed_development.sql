INSERT INTO departments (code, name, description)
VALUES
    ('IT', 'IT Department', 'Development department seed data'),
    ('HK', 'Housekeeping', 'Development department seed data'),
    ('FO', 'Front Office', 'Development department seed data'),
    ('ENG', 'Engineering', 'Development department seed data'),
    ('HR', 'Human Resources', 'Development department seed data'),
    ('FB', 'Food & Beverage', 'Development department seed data')
ON CONFLICT (code) DO NOTHING;

INSERT INTO locations (code, name, description)
VALUES
    ('LOBBY', 'Lobby', 'Development location seed data'),
    ('BALLROOM', 'Ballroom', 'Development location seed data'),
    ('BACK-OFFICE', 'Back Office', 'Development location seed data'),
    ('ROOM-8020', 'Room 8020', 'Development location seed data'),
    ('ROOM-7309', 'Room 7309', 'Development location seed data'),
    ('VILLA', 'Villa', 'Development location seed data')
ON CONFLICT (code) DO NOTHING;

