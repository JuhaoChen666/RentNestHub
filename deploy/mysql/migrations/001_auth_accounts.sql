ALTER TABLE users
    MODIFY COLUMN role ENUM('admin', 'landlord', 'tenant') NOT NULL,
    ADD COLUMN username VARCHAR(80) NULL AFTER role,
    ADD COLUMN password_hash VARCHAR(100) NULL AFTER email,
    ADD UNIQUE INDEX uk_users_username (username);
