package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"bikes/internal/database"
	"bikes/internal/features/home"
	"bikes/internal/features/riders"
	"bikes/internal/ui"
)

func main() {
	dbPath := flag.String("db", "db.sql", "path to the SQLite database file")
	flag.Parse()

	db, err := database.Open(*dbPath)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()
	log.Printf("database ready: %s", *dbPath)

	app, err := ui.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET "+ui.StaticPrefix, ui.StaticHandler())

	// features
	if err := home.Register(mux, app, db); err != nil {
		log.Fatal(err)
	}

	if err := riders.Register(mux, app, db); err != nil {
		log.Fatal(err)
	}

	svr := &http.Server{
		Addr:    "localhost:8080",
		Handler: app.Recover(mux),
	}

	fmt.Println("Listening on http://localhost:8080")
	if err := svr.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
