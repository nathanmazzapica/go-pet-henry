CREATE TABLE users (
	id INTEGER PRIMARY KEY,
	pets INTEGER NOT NULL DEFAULT 0,
	user_id TEXT NOT NULL,
	display_name TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
