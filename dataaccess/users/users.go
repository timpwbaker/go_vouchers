package users

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// Repository contains all the repository methods relating to DeviceAccount.
type Repository interface {
	GetUserItem(email string) (*dynamodb.GetItemOutput, error)
}

// RepoCtx holds an interface to the DB and implements the above Repository
// interface
type RepoCtx struct {
	db *dynamodb.DynamoDB
}

var _ Repository = &RepoCtx{}

// NewRepo instantiates a new instance of the DeviceAccount repository.
func NewRepo(db *dynamodb.DynamoDB) *RepoCtx {
	return &RepoCtx{db: db}
}

func (r *RepoCtx) GetUserItem(email string) (*dynamodb.GetItemOutput, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String("go_vouchers_users"),
	}

	result, err := r.db.GetItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
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
	}

	return result, err
}
