CREATE TABLE team (
                      id SERIAL PRIMARY KEY,
                      name TEXT NOT NULL
);

CREATE TABLE player (
                        id SERIAL PRIMARY KEY,
                        gamer_name TEXT NOT NULL,
                        tag_line TEXT NOT NULL,
                        team_id INT REFERENCES team(id) ON DELETE CASCADE
);
