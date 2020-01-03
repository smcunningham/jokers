CREATE TABLE users(
	user_id serial PRIMARY KEY,
	username VARCHAR (50) UNIQUE NOT NULL,
	password text NOT NULL,
	email VARCHAR (355) UNIQUE NOT NULL,
	firstname VARCHAR (50) NOT NULL,
	lastname VARCHAR (250) NOT NULL,
	created_on TIMESTAMP NOT NULL,
	fav_jokes integer ARRAY
);