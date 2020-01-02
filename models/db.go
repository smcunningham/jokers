package models

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Datastore defines DB functionality
type Datastore interface {
	UserLogin(User) (User, bool)
	InsertUser(User) error
}

// DB contains a sql db
type DB struct {
	*sql.DB
}

// DbConfig holds our application's .env variables
type DbConfig struct {
	Host, Port, User, Password, DbName string
}

// OpenDB opens a connection to the postgres database
func OpenDB(cfg DbConfig) (*DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("INFO:OpenDB:Successful connection to database")
	return &DB{db}, nil
}

// User holds the user table data
type User struct {
	UserID    int       `json:"userId" ,db:"user_id"`
	Username  string    `json:"username" ,db:"username"`
	Password  string    `json:"password" ,db:"password"`
	FirstName string    `json:"firstname" ,db:"firstname"`
	LastName  string    `json:"lastname" ,db:"lastname"`
	Email     string    `json:"email" ,db:"email"`
	CreatedOn time.Time `json:"createdOn" ,db:"created_on"`
	FavJokes  []int     `json:"favJokes" ,db:"fav_jokes"`
}

// UserLogin is called from login template
func (db *DB) UserLogin(u User) (User, bool) {
	// Try to find email in user table
	loginStmt := `SELECT username, password, email, firstname, lastname, created_on FROM users WHERE email=$1`
	row := db.QueryRow(loginStmt, u.Email)

	storedCreds := User{}

	switch err := row.Scan(&storedCreds.Username,
		&storedCreds.Password,
		&storedCreds.Email,
		&storedCreds.FirstName,
		&storedCreds.LastName,
		&storedCreds.CreatedOn); err {
	case sql.ErrNoRows:
		fmt.Printf("ERROR:Scan:No rows returned, username not found!")
		return User{}, false
	case nil:
		err := checkPasswordHash(u.Password, storedCreds.Password)
		if err != nil {
			// No match, probably `hashedPassword is not the hash of the given password`
			fmt.Printf("ERROR:CheckPasswordHash:%s\n", err.Error())
			return User{}, false
		}
		// Passwords match, but don't return password with User{} because this data will be passed to HTML templates
		return User{
			Username:  storedCreds.Username,
			Email:     storedCreds.Email,
			FirstName: storedCreds.FirstName,
			LastName:  storedCreds.LastName,
			CreatedOn: storedCreds.CreatedOn}, true
	default:
		// All other errors
		fmt.Printf("ERROR:UserLogin:%s\n", err.Error())
		return User{}, false
	}
}

// InsertUser inserts a new user into the user table
func (db *DB) InsertUser(u User) error {
	hashedPassword, err := hashPassword(u.Password)
	if err != nil {
		return err
	}

	if _, err := db.Query(`INSERT INTO users(username, password, email, firstname, lastname, created_on) values($1, $2, $3, $4, $5, $6)`,
		u.Username,
		hashedPassword,
		u.Email,
		u.FirstName,
		u.LastName,
		time.Now()); err != nil {
		return err
	}
	return nil
}

// Helper funcs
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}
