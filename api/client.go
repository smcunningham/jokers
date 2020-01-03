package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"jokers/models"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

const (
	templateFileDir = "./web/templates/"
	apiURL          = "http://api.icndb.com/jokes/random"
)

// Client holds the database and Session info for retrieving current user data
type Client struct {
	DB      models.Datastore
	Session *scs.Session
}

// TemplateData contains all data to be passed into the home page template
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

// JokeData contains actual joke data inside JokesResponse `shell`
type JokeData struct {
	ID         int      `json:"id"`
	Joke       string   `json:"joke"`
	Categories []string `json:"categories"`
}

// CustomResponse has data for custom jokes from XHR requests
type CustomResponse struct {
	FirstName string `json:"first"`
	LastName  string `json:"last"`
}

var loginTmpl, homeTmpl, registrationTmpl *template.Template
var td TemplateData

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

	// Open login template
	if err := loginTmpl.ExecuteTemplate(w, "login", nil); err != nil {
		fmt.Printf("ERROR:LoginHandler:Failed to execute template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// HomeHandler handles login requests and routes to the home page if successful. It also
// contains logic for generating jokes
func (c *Client) HomeHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("home", r)

	creds := models.User{
		Email:    r.FormValue("email"),
		Password: r.FormValue("pass"),
	}

	user, err := c.DB.UserLogin(creds)
	if err != nil {
		// Error rendoring home page, return to login
		if err := loginTmpl.ExecuteTemplate(w, "login", nil); err != nil {
			fmt.Printf("ERROR:HomeHandler:Failed to execute template: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Save name for personalized jokes
	c.Session.Put(r.Context(), "firstname", user.FirstName)
	c.Session.Put(r.Context(), "lastname", user.LastName)

	// Create a random joke
	random, err := doReq(apiURL)
	if err != nil {
		fmt.Printf("ERROR:HomeHandler:%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Add random joke to template data
	td.RandomJoke = random

	// Create a personalized joke
	personal, err := personalJoke(user)
	if err != nil {
		fmt.Printf("ERROR:HomeHandler:%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Add personal joke and user to template data
	td.PersonalJoke = personal
	td.User = user

	// Open home page template
	if err := homeTmpl.ExecuteTemplate(w, "home", td); err != nil {
		fmt.Printf("ERROR:HomeHandler:Failed to execute template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// SignupHandler routes a user to the new user registration page
func (c *Client) SignupHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("signupHandler", r)

	// Open registration template page
	if err := registrationTmpl.ExecuteTemplate(w, "registration", nil); err != nil {
		fmt.Printf("ERROR:SignupHandler:Failed to execute template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// SignupActionHandler handles the registration of a new user using an html form
func (c *Client) SignupActionHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("signupActionHandler", r)

	pwd := r.FormValue("password")
	pwdConfirm := r.FormValue("password_confirm")

	// Hacky way to check that the passwords match so that a user isnt created with incorrect pwd
	if !(pwd == pwdConfirm) {
		fmt.Println("Passwords do not match, redirecting to registration page")

		// Redirect to registration page because passwords dont match
		if err := registrationTmpl.ExecuteTemplate(w, "registration", nil); err != nil {
			fmt.Printf("ERROR:SignupHandler:Failed to execute template: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	// Parse and decode request into 'creds' struct to be inserted into db
	creds := models.User{
		Username:  r.FormValue("username"),
		Email:     r.FormValue("email"),
		FirstName: r.FormValue("firstname"),
		LastName:  r.FormValue("lastname"),
		Password:  r.FormValue("password"),
	}

	// Insert new user into db
	err := c.DB.InsertUser(creds)
	if err != nil {
		fmt.Println("ERROR:SignupActionHandler:Error inserting into database:", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Redirect to login page after successful registration
	fmt.Printf("INFO:SignupActionHandler:Inserted user into database: %+v\n", creds)
	if err := loginTmpl.ExecuteTemplate(w, "login", nil); err != nil {
		fmt.Printf("ERROR:SignupActionHandler:Failed to execute template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// RandomJokeHandler handles random jokes generated on the home page
func (c *Client) RandomJokeHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("randomJokeHandler", r)

	// API call to generate random joke
	joke, err := doReq(apiURL)
	if err != nil {
		fmt.Printf("ERROR:RandomJokeHandler:%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Add random joke to template data
	td.RandomJoke = joke

	// Load random joke on home template
	if err := homeTmpl.ExecuteTemplate(w, "random", td); err != nil {
		fmt.Printf("ERROR:RandomJokeHandler:Failed to execute template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// PersonalJokeHandler handles personalized jokes generated from the home page
func (c *Client) PersonalJokeHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("personalJokeHandler", r)

	// Use session data to get user's first and last name, will be inserted into joke
	user := models.User{FirstName: c.Session.GetString(r.Context(), "firstname"),
		LastName: c.Session.GetString(r.Context(), "lastname")}

	// API call to generate joke
	joke, err := personalJoke(user)
	if err != nil {
		fmt.Printf("ERROR:PersonalJokeHandler:%s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Add personal joke to template data
	td.PersonalJoke = joke

	// Load personal joke on home template
	if err := homeTmpl.ExecuteTemplate(w, "personal", td); err != nil {
		fmt.Printf("ERROR:PersonalJokeHandler:Failed to execute template: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// CustomJokeHandler handles custom jokes generated from the modal on the home page
func (c *Client) CustomJokeHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("CustomJokeHandler", r)

	// Get custom name from response to be used in joke
	res := CustomResponse{}
	err := json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		fmt.Printf("ERROR:CustomJokeHandler:%s", err.Error())
		return
	}

	// Create user from custom name
	user := models.User{FirstName: res.FirstName, LastName: res.LastName}

	// API call to generate personal joke using custom name
	joke, err := personalJoke(user)
	if err != nil {
		fmt.Printf("ERROR:CustomJokeHandler:%s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Add custom joke to template data
	td.CustomJoke = joke

	// Load custom joke on home template
	if err := homeTmpl.ExecuteTemplate(w, "custom", td); err != nil {
		fmt.Printf("ERROR:CustomJokeHandler:%s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// This func is used to generate a personal or custom joke using data passed in as a models.User
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

// Helper functions //

// Perform actual HTTP request and return JokeData that holds joke information
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

// Log helper func that specifies name of the func, request method and request path
func consoleLog(name string, r *http.Request) {
	fmt.Printf("%s - Request Method: %s Request URL: %s\n", name, r.Method, r.URL.EscapedPath())
}
