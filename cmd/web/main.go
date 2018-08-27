package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
	"github.com/timpwbaker/go_vouchers/pkg/env"
	"github.com/timpwbaker/go_vouchers/server/handlers"
)

var db *dynamodb.DynamoDB

// Store the redis connection as a package level variable
var cache redis.Conn

func main() {
	appEnv := env.GetAppEnv()
	logger := log.New(os.Stderr, "[boot] ", log.LstdFlags)

	err := env.LoadEnvFileIfNeeded(appEnv)
	if err != nil {
		logger.Fatalf("dotenv error: %v\n", err)
	}

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
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	db = dynamodb.New(sess)
}
