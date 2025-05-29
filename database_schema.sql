-- Premier League Simulator Database Schema
-- This file contains the complete database schema and example queries

-- =====================================================
-- TABLE DEFINITIONS
-- =====================================================

-- Teams table - stores all team information
CREATE TABLE IF NOT EXISTS teams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    short_name VARCHAR(10) NOT NULL,
    base_strength INTEGER NOT NULL,
    current_strength INTEGER NOT NULL,
    played INTEGER DEFAULT 0,
    won INTEGER DEFAULT 0,
    drawn INTEGER DEFAULT 0,
    lost INTEGER DEFAULT 0,
    goals_for INTEGER DEFAULT 0,
    goals_against INTEGER DEFAULT 0,
    goal_difference INTEGER DEFAULT 0,
    points INTEGER DEFAULT 0,
    form VARCHAR(50) DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Leagues table - stores league information
CREATE TABLE IF NOT EXISTS leagues (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    season VARCHAR(20) NOT NULL,
    current_week INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active', -- active, completed
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Matches table - stores all match information
CREATE TABLE IF NOT EXISTS matches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    league_id INTEGER NOT NULL,
    week INTEGER NOT NULL,
    home_team_id INTEGER NOT NULL,
    away_team_id INTEGER NOT NULL,
    home_goals INTEGER DEFAULT 0,
    away_goals INTEGER DEFAULT 0,
    is_played BOOLEAN DEFAULT FALSE,
    is_fixed BOOLEAN DEFAULT FALSE,
    match_date DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (league_id) REFERENCES leagues(id),
    FOREIGN KEY (home_team_id) REFERENCES teams(id),
    FOREIGN KEY (away_team_id) REFERENCES teams(id)
);

-- League teams junction table - links teams to leagues
CREATE TABLE IF NOT EXISTS league_teams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    league_id INTEGER NOT NULL,
    team_id INTEGER NOT NULL,
    position INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (league_id) REFERENCES leagues(id),
    FOREIGN KEY (team_id) REFERENCES teams(id),
    UNIQUE(league_id, team_id)
);

-- Championship probabilities table - stores historical probability data
CREATE TABLE IF NOT EXISTS championship_probabilities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    league_id INTEGER NOT NULL,
    team_id INTEGER NOT NULL,
    week INTEGER NOT NULL,
    probability REAL NOT NULL,
    calculated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (league_id) REFERENCES leagues(id),
    FOREIGN KEY (team_id) REFERENCES teams(id)
);

-- =====================================================
-- INDEXES FOR PERFORMANCE
-- =====================================================

-- Index on matches for faster league queries
CREATE INDEX IF NOT EXISTS idx_matches_league_week ON matches(league_id, week);

-- Index on league_teams for faster team lookups
CREATE INDEX IF NOT EXISTS idx_league_teams_league ON league_teams(league_id);

-- Index on championship probabilities
CREATE INDEX IF NOT EXISTS idx_probabilities_league_week ON championship_probabilities(league_id, week);

-- Index on teams name for faster lookups
CREATE INDEX IF NOT EXISTS idx_teams_name ON teams(name);

-- =====================================================
-- EXAMPLE QUERIES
-- =====================================================

-- 1. Get current league standings
-- Returns teams ordered by points, then goal difference
SELECT 
    t.name,
    t.played,
    t.won,
    t.drawn,
    t.lost,
    t.goals_for,
    t.goals_against,
    t.goal_difference,
    t.points,
    t.current_strength
FROM teams t
JOIN league_teams lt ON t.id = lt.team_id
WHERE lt.league_id = 1 -- Replace with actual league ID
ORDER BY t.points DESC, t.goal_difference DESC, t.goals_for DESC;

-- 2. Get all matches for a specific week
SELECT 
    m.week,
    ht.name AS home_team,
    at.name AS away_team,
    m.home_goals,
    m.away_goals,
    m.is_played,
    m.is_fixed
FROM matches m
JOIN teams ht ON m.home_team_id = ht.id
JOIN teams at ON m.away_team_id = at.id
WHERE m.league_id = 1 AND m.week = 5 -- Replace with actual league ID and week
ORDER BY m.id;

-- 3. Get team statistics with form
SELECT 
    t.name,
    t.points,
    t.goal_difference,
    t.form,
    CASE 
        WHEN t.points >= 20 THEN 'Excellent'
        WHEN t.points >= 15 THEN 'Good'
        WHEN t.points >= 10 THEN 'Average'
        ELSE 'Poor'
    END AS performance_rating
FROM teams t
JOIN league_teams lt ON t.id = lt.team_id
WHERE lt.league_id = 1
ORDER BY t.points DESC;

-- 4. Get championship probabilities for latest week
SELECT 
    t.name,
    cp.probability,
    cp.week
FROM championship_probabilities cp
JOIN teams t ON cp.team_id = t.id
WHERE cp.league_id = 1 
    AND cp.week = (
        SELECT MAX(week) 
        FROM championship_probabilities 
        WHERE league_id = 1
    )
ORDER BY cp.probability DESC;

-- 5. Get match results with goal difference
SELECT 
    m.week,
    ht.name AS home_team,
    at.name AS away_team,
    m.home_goals,
    m.away_goals,
    (m.home_goals - m.away_goals) AS goal_difference,
    CASE 
        WHEN m.home_goals > m.away_goals THEN ht.name
        WHEN m.away_goals > m.home_goals THEN at.name
        ELSE 'Draw'
    END AS winner
FROM matches m
JOIN teams ht ON m.home_team_id = ht.id
JOIN teams at ON m.away_team_id = at.id
WHERE m.league_id = 1 AND m.is_played = TRUE
ORDER BY m.week, m.id;

-- 6. Get team head-to-head record
SELECT 
    t1.name AS team1,
    t2.name AS team2,
    COUNT(*) AS matches_played,
    SUM(CASE 
        WHEN (m.home_team_id = t1.id AND m.home_goals > m.away_goals) OR 
             (m.away_team_id = t1.id AND m.away_goals > m.home_goals) 
        THEN 1 ELSE 0 
    END) AS team1_wins,
    SUM(CASE 
        WHEN m.home_goals = m.away_goals 
        THEN 1 ELSE 0 
    END) AS draws,
    SUM(CASE 
        WHEN (m.home_team_id = t2.id AND m.home_goals > m.away_goals) OR 
             (m.away_team_id = t2.id AND m.away_goals > m.home_goals) 
        THEN 1 ELSE 0 
    END) AS team2_wins
FROM matches m
JOIN teams t1 ON t1.id IN (m.home_team_id, m.away_team_id)
JOIN teams t2 ON t2.id IN (m.home_team_id, m.away_team_id)
WHERE m.league_id = 1 
    AND t1.id != t2.id 
    AND m.is_played = TRUE
    AND t1.name = 'Manchester City' -- Replace with team names
    AND t2.name = 'Arsenal'
GROUP BY t1.id, t2.id;

-- 7. Get top scorers (teams with most goals)
SELECT 
    t.name,
    t.goals_for,
    t.played,
    ROUND(CAST(t.goals_for AS FLOAT) / NULLIF(t.played, 0), 2) AS goals_per_game
FROM teams t
JOIN league_teams lt ON t.id = lt.team_id
WHERE lt.league_id = 1
ORDER BY t.goals_for DESC;

-- 8. Get best defensive teams (least goals conceded)
SELECT 
    t.name,
    t.goals_against,
    t.played,
    ROUND(CAST(t.goals_against AS FLOAT) / NULLIF(t.played, 0), 2) AS goals_conceded_per_game
FROM teams t
JOIN league_teams lt ON t.id = lt.team_id
WHERE lt.league_id = 1
ORDER BY t.goals_against ASC;

-- 9. Get probability trends for a specific team
SELECT 
    cp.week,
    cp.probability,
    t.name
FROM championship_probabilities cp
JOIN teams t ON cp.team_id = t.id
WHERE cp.league_id = 1 
    AND t.name = 'Manchester City' -- Replace with team name
ORDER BY cp.week;

-- 10. Get upcoming fixtures
SELECT 
    m.week,
    ht.name AS home_team,
    at.name AS away_team
FROM matches m
JOIN teams ht ON m.home_team_id = ht.id
JOIN teams at ON m.away_team_id = at.id
WHERE m.league_id = 1 
    AND m.is_played = FALSE
ORDER BY m.week, m.id
LIMIT 10;

-- 11. Get league summary statistics
SELECT 
    l.name AS league_name,
    l.season,
    l.current_week,
    COUNT(DISTINCT lt.team_id) AS total_teams,
    COUNT(m.id) AS total_matches,
    SUM(CASE WHEN m.is_played = TRUE THEN 1 ELSE 0 END) AS played_matches,
    SUM(m.home_goals + m.away_goals) AS total_goals,
    l.status
FROM leagues l
LEFT JOIN league_teams lt ON l.id = lt.league_id
LEFT JOIN matches m ON l.id = m.league_id
WHERE l.id = 1 -- Replace with actual league ID
GROUP BY l.id;

-- 12. Get form table (last 5 matches performance)
SELECT 
    t.name,
    t.form,
    LENGTH(t.form) - LENGTH(REPLACE(t.form, 'W', '')) AS wins_in_form,
    LENGTH(t.form) - LENGTH(REPLACE(t.form, 'D', '')) AS draws_in_form,
    LENGTH(t.form) - LENGTH(REPLACE(t.form, 'L', '')) AS losses_in_form,
    t.points
FROM teams t
JOIN league_teams lt ON t.id = lt.team_id
WHERE lt.league_id = 1
ORDER BY t.points DESC;

-- =====================================================
-- SAMPLE DATA INSERT STATEMENTS
-- =====================================================

-- Insert sample league
INSERT INTO leagues (name, season, current_week, status) 
VALUES ('Premier League Mini', '2024-25', 0, 'active');

-- Insert sample teams (this would be done programmatically)
INSERT OR IGNORE INTO teams (name, short_name, base_strength, current_strength, form) VALUES
('Manchester City', 'MCI', 85, 85, ''),
('Arsenal', 'ARS', 82, 82, ''),
('Liverpool', 'LIV', 83, 83, ''),
('Manchester United', 'MUN', 80, 80, '');

-- Link teams to league (assuming league_id = 1)
INSERT OR IGNORE INTO league_teams (league_id, team_id) 
SELECT 1, id FROM teams WHERE name IN ('Manchester City', 'Arsenal', 'Liverpool', 'Manchester United');

-- =====================================================
-- VIEWS FOR COMMON QUERIES
-- =====================================================

-- View for current league standings
CREATE VIEW IF NOT EXISTS current_standings AS
SELECT 
    ROW_NUMBER() OVER (ORDER BY t.points DESC, t.goal_difference DESC, t.goals_for DESC) AS position,
    t.name,
    t.played,
    t.won,
    t.drawn,
    t.lost,
    t.goals_for,
    t.goals_against,
    t.goal_difference,
    t.points,
    t.form
FROM teams t
JOIN league_teams lt ON t.id = lt.team_id;

-- View for match results with team names
CREATE VIEW IF NOT EXISTS match_results AS
SELECT 
    m.league_id,
    m.week,
    ht.name AS home_team,
    at.name AS away_team,
    m.home_goals,
    m.away_goals,
    CASE 
        WHEN m.home_goals > m.away_goals THEN ht.name
        WHEN m.away_goals > m.home_goals THEN at.name
        ELSE 'Draw'
    END AS result,
    m.is_played,
    m.is_fixed
FROM matches m
JOIN teams ht ON m.home_team_id = ht.id
JOIN teams at ON m.away_team_id = at.id; 