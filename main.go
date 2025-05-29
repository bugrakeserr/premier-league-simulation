package main

import (
	"fmt"
	"log"
)

var db *Database

func main() {
	fmt.Printf("Premier League Simulator\n")
	fmt.Printf("========================\n\n")

	// Initialize database
	fmt.Println("Initializing database...")
	var err error
	db, err = InitDatabase("premier_league.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	fmt.Println("Starting GUI mode...")
	gui := NewGUI()
	gui.window.ShowAndRun()
}
