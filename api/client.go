package api

import (
	"fmt"
	"html/template"
	"jokers/models"
	"log"
	"net/http"
)

const (
	templateFileDir = "./web/templates/"
)

// Client holds the database and will be used to interact w the API
type Client struct {
	DB models.Datastore
}

var loginTmpl, homeTmpl, registrationTmpl *template.Template

func init() {
	// Create and cache templates
	homeTmpl, loginTmpl, registrationTmpl =
		template.Must(template.ParseFiles(templateFileDir+"home.html")),
		template.Must(template.ParseFiles(templateFileDir+"login.html")),
		template.Must(template.ParseFiles(templateFileDir+"registration.html"))
}

// LoginHandler routes the user to the login page
func (c *Client) LoginHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("login", r)

	if err := loginTmpl.ExecuteTemplate(w, "login", nil); err != nil {
		log.Printf("ERROR:LoginHandler:Failed to execute template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// HomeHandler handles login requests and routes to the home page
func (c *Client) HomeHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("home", r)

	creds := models.User{
		Email:    r.FormValue("email"),
		Password: r.FormValue("pass"),
	}
	fmt.Printf("Email: %s\nPassword: %s\n", creds.Email, creds.Password)

	if c.DB.UserLogin(creds) {
		if err := homeTmpl.ExecuteTemplate(w, "home", nil); err != nil {
			log.Printf("ERROR:HomeHandler:Failed to execute template: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// Error rendoring home page, return to login
	if err := loginTmpl.ExecuteTemplate(w, "login", nil); err != nil {
		log.Printf("ERROR:HomeHandler:Failed to execute template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// SignupHandler routes a user to the new user registration
func (c *Client) SignupHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("signupHandler", r)

	if err := registrationTmpl.ExecuteTemplate(w, "registration", nil); err != nil {
		log.Printf("ERROR:SignupHandler:Failed to execute template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// SignupActionHandler handles the actual registration of a new user using the html form
func (c *Client) SignupActionHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("signupActionHandler", r)

	// Parse and decode request into 'creds' struct
	creds := models.User{
		Username: r.FormValue("username"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	err := c.DB.InsertUser(creds)
	if err != nil {
		fmt.Println("ERROR:SignupActionHandler:Error inserting into database:", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("Inserted user into database: %+v\n", creds)
	if err := loginTmpl.ExecuteTemplate(w, "login", nil); err != nil {
		log.Printf("ERROR:SignupActionHandler:Failed to execute template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Helper functions
func consoleLog(name string, r *http.Request) {
	fmt.Printf("%s - Request Method: %s Request URL: %s\n", name, r.Method, r.URL.EscapedPath())
}
