package repository

// CreateParticipantsTableQuery creates the participants table
const CreateParticipantsTableQuery = `
CREATE TABLE IF NOT EXISTS participants (
    competition_id INT NOT NULL,
    dossard_number INT NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    PRIMARY KEY (competition_id, dossard_number)
);
`

// DropParticipantsTableQuery drops the participants table
const DropParticipantsTableQuery = `
DROP TABLE IF EXISTS participants;
`

// CreateCompetitionsTableQuery creates the competitions table
const CreateCompetitionsTableQuery = `
CREATE TABLE IF NOT EXISTS competitions (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    date VARCHAR(50) NOT NULL,
    location VARCHAR(255) NOT NULL,
    organizer VARCHAR(255) NOT NULL,
    contact VARCHAR(255) NOT NULL,
    PRIMARY KEY (id)
);
`

// DropCompetitionsTableQuery drops the competitions table
const DropCompetitionsTableQuery = `
DROP TABLE IF EXISTS competitions;
`

// CreateScalesTableQuery creates the scales table
const CreateScalesTableQuery = `
CREATE TABLE IF NOT EXISTS scales (
    competition_id INT NOT NULL,
    category VARCHAR(100) NOT NULL,
    zone VARCHAR(100) NOT NULL,
    points_door1 INT NOT NULL,
    points_door2 INT NOT NULL,
    points_door3 INT NOT NULL,
    points_door4 INT NOT NULL,
    points_door5 INT NOT NULL,
    points_door6 INT NOT NULL,
    PRIMARY KEY (competition_id, category, zone),
    FOREIGN KEY (competition_id) REFERENCES competitions(id) ON DELETE CASCADE
);
`

// DropScalesTableQuery drops the scales table
const DropScalesTableQuery = `
DROP TABLE IF EXISTS scales;
`

// CreateUsersTableQuery creates the users table
const CreateUsersTableQuery = `
CREATE TABLE IF NOT EXISTS users (
    id INT NOT NULL AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    roles VARCHAR(500) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY (email)
);
`

// DropUsersTableQuery drops the users table
const DropUsersTableQuery = `
DROP TABLE IF EXISTS users;
`

// CreateRunsTableQuery creates the runs table
const CreateRunsTableQuery = `
CREATE TABLE IF NOT EXISTS runs (
    competition_id INT NOT NULL,
    dossard INT NOT NULL,
    run_number INT NOT NULL,
    zone VARCHAR(100) NOT NULL,
    door1 BOOLEAN NOT NULL DEFAULT false,
    door2 BOOLEAN NOT NULL DEFAULT false,
    door3 BOOLEAN NOT NULL DEFAULT false,
    door4 BOOLEAN NOT NULL DEFAULT false,
    door5 BOOLEAN NOT NULL DEFAULT false,
    door6 BOOLEAN NOT NULL DEFAULT false,
    penality INT NOT NULL DEFAULT 0,
    chrono_sec INT NOT NULL DEFAULT 0,
    referee_id INT NOT NULL DEFAULT 0,
    PRIMARY KEY (competition_id, run_number, dossard),
    FOREIGN KEY (competition_id, dossard) REFERENCES participants(competition_id, dossard_number) ON DELETE CASCADE
);
`

// DropRunsTableQuery drops the runs table
const DropRunsTableQuery = `
DROP TABLE IF EXISTS runs;
`

// CreateLiverankingsTableQuery creates the liverankings table
const CreateLiverankingsTableQuery = `
CREATE TABLE IF NOT EXISTS liverankings (
    competition_id INT NOT NULL,
    dossard_number INT NOT NULL,
    number_of_runs INT NOT NULL DEFAULT 0,
    total_points INT NOT NULL DEFAULT 0,
    penality INT NOT NULL DEFAULT 0,
    chrono_sec INT NOT NULL DEFAULT 0,
    PRIMARY KEY (competition_id, dossard_number),
    FOREIGN KEY (competition_id, dossard_number) REFERENCES participants(competition_id, dossard_number) ON DELETE CASCADE
);
`

// DropLiverankingsTableQuery drops the liverankings table
const DropLiverankingsTableQuery = `
DROP TABLE IF EXISTS liverankings;
`

// SetupDatabase creates necessary tables for the application
func SetupDatabase(db interface{}) error {
	// The actual implementation depends on the database/sql package or ORM being used
	// For a basic implementation with database/sql:
	/*
		_, err := db.(*sql.DB).Exec(CreateUsersTableQuery)
		if err != nil {
			return err
		}
		_, err = db.(*sql.DB).Exec(CreateCompetitionsTableQuery)
		if err != nil {
			return err
		}
		_, err = db.(*sql.DB).Exec(CreateParticipantsTableQuery)
		if err != nil {
			return err
		}
		_, err = db.(*sql.DB).Exec(CreateScalesTableQuery)
		if err != nil {
			return err
		}
		_, err = db.(*sql.DB).Exec(CreateRunsTableQuery)
		return err
	*/

	// This is a placeholder - implement based on the actual DB interface used in the project
	return nil
}
