package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"
	"github.com/timpwbaker/go_vouchers/pkg/env"
	"github.com/timpwbaker/go_vouchers/server/handlers"
)

var db *dynamodb.DynamoDB

var redisPool *redis.Pool

func main() {
	appEnv := env.GetAppEnv()
	logger := log.New(os.Stderr, "[boot] ", log.LstdFlags)

	err := env.LoadEnvFileIfNeeded(appEnv)
	if err != nil {
		logger.Fatalf("dotenv error: %v\n", err)
	}

	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	// "Signin" and "Signup" are handler that we will implement
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		handlers.Signup(w, r, db, redisPool)
	})
	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
		handlers.Signin(w, r, db, redisPool)
	})
	http.HandleFunc("/welcome", func(w http.ResponseWriter, r *http.Request) {
		handlers.Welcome(w, r, redisPool)
	})
	// initialize our database connection
	initDB()
	initCache()
	// start the server on the port specificed in the ENV
	log.Fatal(http.ListenAndServe(":"+port, nil))
	fmt.Printf("running server")
}

func initCache() {
	redisURL := os.Getenv("REDIS_URL")
	// Assign the connection to the package level `redisPool` variable
	redisPool = newPool(redisURL)
}

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.DialURL(addr) },
	}
}

func initDB() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	db = dynamodb.New(sess)
}
