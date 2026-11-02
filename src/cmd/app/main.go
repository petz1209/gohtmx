package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"

	"bikes/internal/database"
)

//go:embed static
var staticFS embed.FS

func main() {
	dbPath := flag.String("db", "db.sql", "path to the SQLite database file")
	flag.Parse()

	db, err := database.Open(*dbPath)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()
	log.Printf("database ready: %s", *dbPath)

	mux := http.NewServeMux()

	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))

	// wire up your routing.
	// probably think of an html renderer

	svr := &http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}

	fmt.Println("Listening on http://localhost:8080")
	if err := svr.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
