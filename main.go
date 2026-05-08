package main

import (
	"expspl/data"
	"expspl/handle"
	"expspl/routes"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Env vr not loading")
		return
	}
	db := data.Conndb()
	data.CreateTables(db)
	h := &handle.Handler{DB: db}
	mux := routes.SetupRoutes(h)
	fmt.Println("Running on 8080")
	log.Fatal(http.ListenAndServe(":8080", mux))

}
