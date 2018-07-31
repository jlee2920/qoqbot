CREATE TABLE regulars(
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    current_songs INT
);