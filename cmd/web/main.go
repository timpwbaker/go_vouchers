package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
	"github.com/timpwbaker/go_burgers/server/handlers"
)

// The "DB" package level variable will hold the reference to our database instance
var db *sql.DB

// Store the redis connection as a package level variable
var cache redis.Conn

func main() {
	// "Signin" and "Signup" are handler that we will implement
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		handlers.Signup(w, r, db, cache)
	})
	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
		handlers.Signin(w, r, db, cache)
	})
	http.HandleFunc("/welcome", func(w http.ResponseWriter, r *http.Request) {
		handlers.Welcome(w, r, cache)
	})
	// initialize our database connection
	initDB()
	initCache()
	// start the server on port 8000
	log.Fatal(http.ListenAndServe(":8000", nil))
	fmt.Printf("running server")
}

func initCache() {
	// Initialize the redis connection to a redis instance running on your local machine
	conn, err := redis.DialURL("redis://localhost")
	if err != nil {
		panic(err)
	}
	// Assign the connection to the package level `cache` variable
	cache = conn
}

func initDB() {
	var err error
	// Connect to the postgres db
	//you might have to change the connection string to add your database credentials
	db, err = sql.Open("postgres", "dbname=go_burgers sslmode=disable")
	if err != nil {
		panic(err)
	}
}
