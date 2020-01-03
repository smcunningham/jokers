JOKERS is a fun ,full-stack RESTful application that uses a modern technology stack to generate funny jokes from the Internet Chuck Norris Database (ICND). It uses `Bootstrap HTML/CSS` templates, along with some `Javascript` and `JQuery` to give it a simple yet elegant feel. `Golang` is used for the server-side code underneath the hood, and data is persisted using a `PostgreSQL` database.

To use the application, you need to create the database using the provided SQL schema file. Update the .env file so that the desired port and database login credentials are also correct.

To run the program, use the command `go run main.go` in a terminal. From here, point a browser to `http://localhost:{{port}}/`. This will take the user to a login page which includes the ability to create a new account. Create an account and login.. From the home page, the user has the ability to generate jokes in several different ways.

Going from left to right, the three joke generators work as follows:

1. Generate a random joke. Every time that `Generate` is clicked, a random Chuck Norris joke will be displayed in this div.
2. Generate a personalized joke. Every time that a user clicks `Generate`, a joke will be created with the first and last name from the current logged-in account.
3. Generate a custom joke. Clicking `Generate` will open a modal that allows the user to enter a custom name. This name will then be used to generate a random joke. 

Each of these generators work asynchronously, and can be clicked as many times as desired. Data about the current session is saved using the SCS library, and BCrypt is used to hash and decrypt passwords that are stored to increase security. I hope you enjoy playing around with this silly application.

Have fun!