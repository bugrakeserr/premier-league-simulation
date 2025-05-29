package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// database connection and all the operations we need
type Database struct {
	db *sql.DB
}

// set up the database and create all the tables we need
func InitDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	database := &Database{db: db}

	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	return database, nil
}

// create all the tables we need for the database
func (d *Database) createTables() error {
	// teams table with all their stats
	teamsTable := `
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
	);`

	// leagues table to track different seasons
	leaguesTable := `
	CREATE TABLE IF NOT EXISTS leagues (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(100) NOT NULL,
		season VARCHAR(20) NOT NULL,
		current_week INTEGER DEFAULT 0,
		status VARCHAR(20) DEFAULT 'active', -- active, completed
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// matches table with all the game results
	matchesTable := `
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
	);`

	// junction table to link leagues and teams
	leagueTeamsTable := `
	CREATE TABLE IF NOT EXISTS league_teams (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		league_id INTEGER NOT NULL,
		team_id INTEGER NOT NULL,
		position INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (league_id) REFERENCES leagues(id),
		FOREIGN KEY (team_id) REFERENCES teams(id),
		UNIQUE(league_id, team_id)
	);`

	// table to store championship probability history
	probabilitiesTable := `
	CREATE TABLE IF NOT EXISTS championship_probabilities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		league_id INTEGER NOT NULL,
		team_id INTEGER NOT NULL,
		week INTEGER NOT NULL,
		probability REAL NOT NULL,
		calculated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (league_id) REFERENCES leagues(id),
		FOREIGN KEY (team_id) REFERENCES teams(id)
	);`

	tables := []string{teamsTable, leaguesTable, matchesTable, leagueTeamsTable, probabilitiesTable}

	for _, table := range tables {
		if _, err := d.db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	return nil
}

// SaveTeam saves or updates a team in the database
func (d *Database) SaveTeam(team *Team) (int64, error) {
	// make sure we have valid team data
	if team == nil || team.Name == "" {
		return 0, fmt.Errorf("invalid team data: team name is required")
	}

	query := `
	INSERT OR REPLACE INTO teams 
	(name, short_name, base_strength, current_strength, played, won, drawn, lost, 
	 goals_for, goals_against, goal_difference, points, form, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	formStr := ""
	if team.Form != nil {
		for i, f := range team.Form {
			if f != "" {
				if i > 0 && formStr != "" {
					formStr += ","
				}
				formStr += f
			}
		}
	}

	result, err := d.db.Exec(query,
		team.Name, getShortName(team.Name), team.BaseStrength, team.CurrentStrength,
		team.Played, team.Won, team.Drawn, team.Lost,
		team.GoalsFor, team.GoalsAgainst, team.GoalDifference, team.Points, formStr)

	if err != nil {
		return 0, fmt.Errorf("failed to save team %s: %v", team.Name, err)
	}

	return result.LastInsertId()
}

// save a new league to the database
func (d *Database) SaveLeague(league *League, name, season string) (int64, error) {
	query := `
	INSERT INTO leagues (name, season, current_week, status)
	VALUES (?, ?, ?, 'active')`

	result, err := d.db.Exec(query, name, season, league.Week)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// save a match result to the database
func (d *Database) SaveMatch(leagueID int64, match *Match) error {
	// get the team ids first
	homeTeamID, err := d.getTeamID(match.HomeTeam.Name)
	if err != nil {
		return err
	}

	awayTeamID, err := d.getTeamID(match.AwayTeam.Name)
	if err != nil {
		return err
	}

	query := `
	INSERT OR REPLACE INTO matches 
	(league_id, week, home_team_id, away_team_id, home_goals, away_goals, is_played, is_fixed, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	_, err = d.db.Exec(query, leagueID, match.Week, homeTeamID, awayTeamID,
		match.HomeGoals, match.AwayGoals, match.IsPlayed, match.IsFixed)

	return err
}

// save championship probabilities for a specific week
func (d *Database) SaveChampionshipProbabilities(leagueID int64, week int, probabilities map[string]float64) error {
	// clear out old probabilities for this league and week
	deleteQuery := "DELETE FROM championship_probabilities WHERE league_id = ? AND week = ?"
	_, err := d.db.Exec(deleteQuery, leagueID, week)
	if err != nil {
		return err
	}

	// add the new probabilities
	insertQuery := `
	INSERT INTO championship_probabilities (league_id, team_id, week, probability)
	VALUES (?, ?, ?, ?)`

	for teamName, prob := range probabilities {
		teamID, err := d.getTeamID(teamName)
		if err != nil {
			continue // skip if we can't find the team
		}

		_, err = d.db.Exec(insertQuery, leagueID, teamID, week, prob)
		if err != nil {
			return err
		}
	}

	return nil
}

// getTeamID gets a team's ID by name
func (d *Database) getTeamID(teamName string) (int64, error) {
	var teamID int64
	query := "SELECT id FROM teams WHERE name = ?"
	err := d.db.QueryRow(query, teamName).Scan(&teamID)
	return teamID, err
}

// get the current league standings sorted by points
func (d *Database) GetLeagueStandings(leagueID int64) ([]*Team, error) {
	query := `
	SELECT t.name, t.base_strength, t.current_strength, t.played, t.won, t.drawn, t.lost,
	       t.goals_for, t.goals_against, t.goal_difference, t.points, t.form
	FROM teams t
	JOIN league_teams lt ON t.id = lt.team_id
	WHERE lt.league_id = ?
	ORDER BY t.points DESC, t.goal_difference DESC, t.goals_for DESC`

	rows, err := d.db.Query(query, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		team := &Team{}
		var formStr string

		err := rows.Scan(&team.Name, &team.BaseStrength, &team.CurrentStrength,
			&team.Played, &team.Won, &team.Drawn, &team.Lost,
			&team.GoalsFor, &team.GoalsAgainst, &team.GoalDifference,
			&team.Points, &formStr)
		if err != nil {
			return nil, err
		}

		// convert form string back to array
		if formStr != "" {
			team.Form = parseFormString(formStr)
		} else {
			team.Form = make([]string, 5)
		}

		teams = append(teams, team)
	}

	return teams, nil
}

// get all matches for a league organized by week
func (d *Database) GetLeagueMatches(leagueID int64) ([][]Match, error) {
	query := `
	SELECT m.week, ht.name, at.name, m.home_goals, m.away_goals, m.is_played, m.is_fixed
	FROM matches m
	JOIN teams ht ON m.home_team_id = ht.id
	JOIN teams at ON m.away_team_id = at.id
	WHERE m.league_id = ?
	ORDER BY m.week, m.id`

	rows, err := d.db.Query(query, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matchesByWeek := make(map[int][]Match)
	maxWeek := 0

	for rows.Next() {
		var week int
		var homeTeamName, awayTeamName string
		var homeGoals, awayGoals int
		var isPlayed, isFixed bool

		err := rows.Scan(&week, &homeTeamName, &awayTeamName,
			&homeGoals, &awayGoals, &isPlayed, &isFixed)
		if err != nil {
			return nil, err
		}

		match := Match{
			HomeTeam:  &Team{Name: homeTeamName},
			AwayTeam:  &Team{Name: awayTeamName},
			HomeGoals: homeGoals,
			AwayGoals: awayGoals,
			IsPlayed:  isPlayed,
			IsFixed:   isFixed,
			Week:      week,
		}

		matchesByWeek[week] = append(matchesByWeek[week], match)
		if week > maxWeek {
			maxWeek = week
		}
	}

	// turn the map into ordered weeks and make sure we have all 18 weeks
	var fixtures [][]Match
	for week := 1; week <= 18; week++ {
		if matches, exists := matchesByWeek[week]; exists {
			fixtures = append(fixtures, matches)
		} else {
			// add empty week if no matches (shouldn't happen but just in case)
			fixtures = append(fixtures, []Match{})
		}
	}

	return fixtures, nil
}

// close the database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// helper functions to make life easier
func getShortName(fullName string) string {
	shortNames := map[string]string{
		"Manchester City":   "MCI",
		"Arsenal":           "ARS",
		"Liverpool":         "LIV",
		"Manchester United": "MUN",
		"Tottenham":         "TOT",
		"Newcastle":         "NEW",
		"Chelsea":           "CHE",
		"Aston Villa":       "AVL",
		"Brighton":          "BHA",
		"West Ham":          "WHU",
	}

	if short, exists := shortNames[fullName]; exists {
		return short
	}
	return fullName[:3] // just use first 3 characters if we don't have a mapping
}

func parseFormString(formStr string) []string {
	if formStr == "" {
		return make([]string, 5)
	}

	form := make([]string, 5)
	for i, char := range formStr {
		if i < 5 && i < len(formStr) {
			form[i] = string(char)
		}
	}
	return form
}
