package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"jokers/models"
	"log"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

const (
	templateFileDir = "./web/templates/"
	apiURL          = "http://api.icndb.com/jokes/random"
)

// Client holds the database and will be used to interact w the API
type Client struct {
	DB      models.Datastore
	Session *scs.Session
}

// TemplateData contains all data that will be passed into the home page template
type TemplateData struct {
	User         models.User
	RandomJoke   JokeData
	PersonalJoke JokeData
	CustomJoke   JokeData
}

// JokeResponse contains data returned from joke api call
type JokeResponse struct {
	Type  string   `json:"type"`
	Value JokeData `json:"value"`
}

// JokeData contains actual joke data inside JokesResponse shell
type JokeData struct {
	ID         int      `json:"id"`
	Joke       string   `json:"joke"`
	Categories []string `json:"categories"`
}

// CustomResponse is used to name data from XHR requests
type CustomResponse struct {
	FirstName string `json:"first"`
	LastName  string `json:"last"`
}

var loginTmpl, homeTmpl, registrationTmpl *template.Template
var td TemplateData

// User is exported to hold the data of current logged-in user
var User models.User

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

	user, ok := c.DB.UserLogin(creds)
	if ok {
		// Save name for personalized jokes
		c.Session.Put(r.Context(), "firstname", user.FirstName)
		c.Session.Put(r.Context(), "lastname", user.LastName)

		// Create a random joke
		random, err := doReq(apiURL)
		if err != nil {
			log.Printf("ERROR:HomeHandler:%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		td.RandomJoke = random

		// Create a personalized joke
		personal, err := personalJoke(user)
		if err != nil {
			log.Printf("ERROR:HomeHandler:%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		td.PersonalJoke = personal

		td.User = user

		if err := homeTmpl.ExecuteTemplate(w, "home", td); err != nil {
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
		Username:  r.FormValue("username"),
		Email:     r.FormValue("email"),
		FirstName: r.FormValue("firstname"),
		LastName:  r.FormValue("lastname"),
		Password:  r.FormValue("password"),
	}

	err := c.DB.InsertUser(creds)
	if err != nil {
		fmt.Println("ERROR:SignupActionHandler:Error inserting into database:", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("INFO:SignupActionHandler:Inserted user into database: %+v\n", creds)
	if err := loginTmpl.ExecuteTemplate(w, "login", nil); err != nil {
		log.Printf("ERROR:SignupActionHandler:Failed to execute template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// RandomJokeHandler handles random jokes being generated on the home page
func (c *Client) RandomJokeHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("randomJokeHandler", r)

	joke, err := doReq(apiURL)
	if err != nil {
		fmt.Printf("ERROR:RandomJokeHandler:%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	td.RandomJoke = joke

	if err := homeTmpl.ExecuteTemplate(w, "random", td); err != nil {
		log.Printf("ERROR:RandomJokeHandler:Failed to execute template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// PersonalJokeHandler handles personalized jokes being generated from the home page
func (c *Client) PersonalJokeHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("personalJokeHandler", r)

	user := models.User{FirstName: c.Session.GetString(r.Context(), "firstname"),
		LastName: c.Session.GetString(r.Context(), "lastname")}

	joke, err := personalJoke(user)
	if err != nil {
		fmt.Printf("ERROR:PersonalJokeHandler:%s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	td.PersonalJoke = joke

	if err := homeTmpl.ExecuteTemplate(w, "personal", td); err != nil {
		log.Printf("ERROR:PersonalJokeHandler:Failed to execute template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// CustomJokeHandler handles custom jokes generated from the custom modal on the home page
func (c *Client) CustomJokeHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("CustomJokeHandler", r)

	res := CustomResponse{}
	err := json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		fmt.Printf("ERROR:CustomJokeHandler:%s", err.Error())
	}

	user := models.User{FirstName: res.FirstName, LastName: res.LastName}

	joke, err := personalJoke(user)
	if err != nil {
		fmt.Printf("ERROR:CustomJokeHandler:%s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
	td.CustomJoke = joke

	if err := homeTmpl.ExecuteTemplate(w, "custom", td); err != nil {
		log.Printf("ERROR:CustomJokeHandler:%s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func personalJoke(u models.User) (JokeData, error) {
	fname := u.FirstName
	lname := u.LastName
	url := apiURL + `?firstName=` + fname + `&lastName=` + lname

	joke, err := doReq(url)
	if err != nil {
		return JokeData{}, err
	}
	return joke, nil
}

// Helper functions
func doReq(url string) (JokeData, error) {
	var jr JokeResponse

	req, _ := http.NewRequest("GET", url, nil)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return JokeData{}, err
	}
	defer res.Body.Close()

	response, _ := ioutil.ReadAll(res.Body)

	err = json.Unmarshal(response, &jr)
	if err != nil {
		return JokeData{}, err
	}

	return jr.Value, nil
}

func consoleLog(name string, r *http.Request) {
	fmt.Printf("%s - Request Method: %s Request URL: %s\n", name, r.Method, r.URL.EscapedPath())
}
