CREATE TABLE IF NOT EXISTS players (
     player_id VARCHAR(255) PRIMARY KEY,
     nickname VARCHAR(255) UNIQUE NOT NULL,
     updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS matches (
     match_id VARCHAR(255) NOT NULL,
     player_id VARCHAR(255) NOT NULL
     REFERENCES players(player_id)
     ON DELETE CASCADE,
     map VARCHAR(100) NOT NULL,
     result VARCHAR(50) NOT NULL,
     score VARCHAR(50) NOT NULL,
     kills INT NOT NULL CHECK (kills >= 0),
     deaths INT NOT NULL CHECK (deaths >= 0),
     kd_ratio NUMERIC(5, 2) NOT NULL,
     saved_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
     PRIMARY KEY (match_id, player_id)
);