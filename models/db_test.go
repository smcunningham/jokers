package models_test

import (
	"jokers/models"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUserLogin(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error was not expected when opening stub DB conn: %s", err.Error())
	}

	testDB := models.DB{
		DB: db,
	}

	testPassword, err := bcrypt.GenerateFromPassword([]byte("password"), 10)
	if err != nil {
		t.Fatalf("Failed generating hashed password: %s", err.Error())
	}

	// User supplied to test func
	user := models.User{
		Email:    "tEmail",
		Password: string(testPassword),
	}

	time := time.Now()

	rows := sqlmock.NewRows([]string{"username", "password", "email", "firstname", "lastname", "created_on"}).
		AddRow("tUser", string(testPassword), "tEmail", "tFirst", "tLast", time)

	mockDB.ExpectQuery("^SELECT (.+) FROM users *").
		WithArgs("tEmail").
		WillReturnRows(rows)

	// Test func
	returnUser, err := testDB.UserLogin(user)
	if err != nil {
		t.Fatalf("ERROR:TestUserLogin: %s", err.Error())
	}

	// User to compare to test func
	mockUser := models.User{
		Username:  "tUser",
		Email:     "tEmail",
		FirstName: "tFirst",
		LastName:  "tLast",
		CreatedOn: time,
	}

	assert.Equal(t, mockUser, returnUser)
}

func TestInsertUser(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error was not expected when opening stub DB conn: %s", err.Error())
	}

	testDB := models.DB{
		DB: db,
	}

	// User supplied to test func
	user := models.User{
		Email:     "tEmail",
		Password:  "password",
		Username:  "tUser",
		FirstName: "tFirst",
		LastName:  "tLast",
		CreatedOn: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"username", "password", "email", "firstname", "lastname", "created_on"}).
		AddRow("tUser", "password", "tEmail", "tFirst", "tLast", time.Now())

	mockDB.ExpectQuery(regexp.QuoteMeta("INSERT INTO users(username, password, email, firstname, lastname, created_on) values($1, $2, $3, $4, $5, $6)")).
		WithArgs(user.Username, sqlmock.AnyArg(), user.Email, user.FirstName, user.LastName, sqlmock.AnyArg()).
		WillReturnRows(rows)

	err = testDB.InsertUser(user)
	if err != nil {
		t.Fatalf("ERROR:InsertUser: %s", err.Error())
	}

}
