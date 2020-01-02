package main

import (
	"fmt"
	"jokers/api"
	"jokers/models"
	"log"
	"net/http"
	"os"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var session *scs.Session

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

	client.Session = scs.NewSession()
	client.Session.Store = memstore.New()

	// Create router, serve static assets
	router := mux.NewRouter().StrictSlash(true)
	router.
		PathPrefix("/web/static").
		Handler(http.StripPrefix("/web/static", http.FileServer(http.Dir("."+"/web/static"))))

	// Handle templates
	router.HandleFunc("/", client.LoginHandler)
	router.HandleFunc("/home", client.HomeHandler)
	router.HandleFunc("/signup", client.SignupHandler)
	router.HandleFunc("/signupact", client.SignupActionHandler)

	router.HandleFunc("/jokes/random", client.RandomJokeHandler)
	router.HandleFunc("/jokes/personal", client.PersonalJokeHandler)

	log.Println("listening...")
	http.ListenAndServe(":3000", client.Session.LoadAndSave(router))
}
