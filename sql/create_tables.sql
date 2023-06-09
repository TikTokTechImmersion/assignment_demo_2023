-- See https://www.postgresqltutorial.com/postgresql-date-functions/postgresql-current_timestamp/
-- for how to set default timestamp as current time
CREATE TABLE messages (
	message_id SERIAL PRIMARY KEY,
	chat VARCHAR(50) NOT NULL,
	sender VARCHAR(50) NOT NULL,
	message TEXT NOT NULL,
	message_send_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (chat, message_send_time),
	CHECK ((chat LIKE CONCAT(sender, ':_%')) OR (chat LIKE CONCAT('_%:', sender)))
);
