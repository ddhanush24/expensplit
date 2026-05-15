package data

import (
	"database/sql"
	"log"
)

func CreateTables(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users(
	    id SERIAL PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		password TEXT,
		name TEXT,
		email TEXT,
		ph_no VARCHAR(15),
		cr_date DATE,
		grps INT DEFAULT 0,
		friends INT[]
	);`) //friends can be depreciated once all checks are made
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// friends TABE
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS friends(
	    id SERIAL PRIMARY KEY,
		user_id1 int,
		user_id2 int,
		accepted BOOL DEFAULT false,
		requestedby int,
		UNIQUE(user_id1,user_id2),
		FOREIGN KEY (user_id1) REFERENCES users(id),
		FOREIGN KEY (user_id2) REFERENCES users(id),
		FOREIGN KEY (requestedby) REFERENCES users(id)
	);`)
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// grousp TABLE
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS groups(
	    id SERIAL PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		created_by int,
		cr_date DATE,
		members INT[] DEFAULT '{}',
		FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
	);`) //Memebers next change to sep tble
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS expenses(
	    id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		amount FLOAT DEFAULT 0.00,
		paidby INT NOT NULL,
		cr_date DATE,
		group_id INT NOT NULL,
		FOREIGN KEY (paidby) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (group_id) REFERENCES groups(id)
	);`) //friends can be depreciated once all checks are made
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Table transactions
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS transactions(
	    id SERIAL PRIMARY KEY,
		id1 INT NOT NULL,
		id2 INT NOT NULL,
		amt1 FLOAT DEFAULT 0.00,
		remainingamt FLOAT DEFAULT 0.00,
		expid INT NOT NULL,
		FOREIGN KEY (id1) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (id2) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (expid) REFERENCES expenses(id)
	);`) //complete_date DATE can be added for tracking Ig+ also useON DELETE CASCADE expense
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}
	log.Println("Created all Tables")

}
