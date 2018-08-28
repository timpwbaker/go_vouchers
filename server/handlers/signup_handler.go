package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gomodule/redigo/redis"
	uuid "github.com/satori/go.uuid"
	"github.com/timpwbaker/go_vouchers/dataaccess/users"
	"golang.org/x/crypto/bcrypt"
)

// Create a struct that models the structure of a user, both in the request body, and in the DB
type Credentials struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type Session struct {
	Token string `json:"session_token"`
}

func Signin(w http.ResponseWriter, r *http.Request, db *dynamodb.DynamoDB, cache redis.Conn) {
	// Parse and decode the request body into a new `Credentials` instance
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	repo := users.NewRepo(db)
	result, err := repo.GetUserItem(creds.Email)

	storedCreds := Credentials{
		Email:    *result.Item["Email"].S,
		Password: *result.Item["Password"].S,
	}
	// Compare the stored hashed password, with the hashed version of the password that was received
	if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
		// If the two passwords don't match, return a 401 status
		w.WriteHeader(http.StatusUnauthorized)
	}

	// Create a new random session token
	sessionToken, err := uuid.NewV4()
	if err != nil {
		// If there is an error in setting the cache, return an internal server error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the token in the cache, along with the user whom it represents
	// The token has an expiry time of 120 seconds
	_, err = cache.Do("SETEX", sessionToken.String(), "120", creds.Email)
	if err != nil {
		// If there is an error in setting the cache, return an internal server error
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to set session token in redis: %v", err)
		return
	}

	session := Session{Token: sessionToken.String()}
	body, err := json.Marshal(session)

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)

	// If we reach this point, that means the users password was correct, and that they are authorized
	// The default 200 status is sent
}

func Signup(w http.ResponseWriter, r *http.Request, db *dynamodb.DynamoDB, cache redis.Conn) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("go_vouchers_users"),
		ExpressionAttributeNames: map[string]*string{
			"#P": aws.String("Password"),
		},
		Key: map[string]*dynamodb.AttributeValue{
			"Email": {
				S: aws.String(creds.Email),
			},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":p": {
				S: aws.String(string(hashedPassword)),
			},
		},
		UpdateExpression: aws.String("SET #P = :p"),
	}

	result, err := db.UpdateItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				fmt.Println(dynamodb.ErrCodeConditionalCheckFailedException, aerr.Error())
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
				fmt.Println(dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}
	fmt.Println(result)
}

func Welcome(w http.ResponseWriter, r *http.Request, cache redis.Conn) {

	sessionToken := r.Header.Get("Session_token")
	if sessionToken == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// We then get the name of the user from our cache, where we set the session token
	response, err := cache.Do("GET", sessionToken)
	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if response == nil {
		// If the session token is not present in cache, return an unauthorized error
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Finally, return the welcome message to the user
	w.Write([]byte(fmt.Sprintf("Welcome %s!", response)))

	log.Printf("logged in %s", response)
}
