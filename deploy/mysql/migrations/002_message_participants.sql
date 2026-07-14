ALTER TABLE messages
    ADD COLUMN recipient_id BIGINT UNSIGNED NULL AFTER sender_id,
    ADD KEY idx_messages_participant_created (sender_id, recipient_id, created_at),
    ADD CONSTRAINT fk_messages_recipient FOREIGN KEY (recipient_id) REFERENCES users (id);

UPDATE messages m
JOIN houses h ON h.id = m.house_id
SET m.recipient_id = h.landlord_id
WHERE m.recipient_id IS NULL;

ALTER TABLE messages MODIFY COLUMN recipient_id BIGINT UNSIGNED NOT NULL;
