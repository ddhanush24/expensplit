package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Conndb() *sql.DB {
	connstr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_URL"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	db, err := sql.Open("pgx", connstr)
	if err != nil {
		log.Fatal("Error ocnnecting to database")

	}
	log.Println("Connected to the postgrSQL server ")

	return db
}
