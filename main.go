package main

import (
	"fmt"
	"jokers/api"
	"jokers/models"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env vars
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("ERROR:Load:%s\n", err.Error())
	}

	// Create DB config with .env values
	dbConfig := models.DbConfig{
		Host:     os.Getenv("host"),
		Port:     os.Getenv("port"),
		User:     os.Getenv("user"),
		Password: os.Getenv("password"),
		DbName:   os.Getenv("dbname"),
	}
	fmt.Printf(".env Variables:\n%+v\n", dbConfig)

	// Create DB
	db, err := models.OpenDB(dbConfig)
	if err != nil {
		fmt.Printf("ERROR:OpenDB:%s", err.Error())
	}
	// Create a client and assign our DB to it for use with http handlers
	client := &api.Client{DB: db}

	// Serve static files
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/web/static/", http.StripPrefix("/web/static/", fs))

	// Handle templates
	http.HandleFunc("/", client.LoginHandler)
	http.HandleFunc("/home", client.HomeHandler)
	http.HandleFunc("/signup", client.SignupHandler)
	http.HandleFunc("/signupact", client.SignupActionHandler)

	log.Println("listening...")
	http.ListenAndServe(":3000", nil)
}
