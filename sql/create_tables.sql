DROP DATABASE IF EXISTS assignment_demo_2023;
CREATE DATABASE assignment_demo_2023;

-- Use database assignment_demo_2023
\c assignment_demo_2023;

DROP TABLE IF EXISTS messages;

CREATE TABLE messages (
	message_id SERIAL PRIMARY KEY,
	sender VARCHAR(50) NOT NULL,
	receiver VARCHAR(50) NOT NULL,
	message_send_time TIMESTAMP NOT NULL,
	message TEXT NOT NULL,
	UNIQUE (sender, receiver, message_send_time)
);
