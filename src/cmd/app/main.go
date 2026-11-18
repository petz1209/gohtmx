package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"bikes/internal/database"
	"bikes/internal/features/home"
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

	mux := http.NewServeMux()

	scriptRegistry := ui.NewRegistry(ui.UiFS)
	if err := home.NewHomeRouter(mux, ui.UiFS, scriptRegistry); err != nil {

		log.Fatal(err)
	}

	fileServer := scriptRegistry.Handler()

	//mux.Handle("/static/", http.FileServer(http.FS(ui.StaticFS)))
	mux.Handle("/static/", fileServer)

	svr := &http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}

	fmt.Println("Listening on http://localhost:8080")
	if err := svr.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
