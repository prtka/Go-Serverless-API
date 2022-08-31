package user

import (
	"encoding/json"
	"errors"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var (
	ErrorMsgCustomer             = "Error  Failed to fetch response"
	ErrorFailedtoUnmarshalRecord = "Error failed to unmarshal record"
	ErrorFetchingData            = "Error fetching data"
	ErrorInvalidUserData         = "ErrorInvalidUserData"
	ErrorInvalidEmail            = "ErrorInvalidEmail"
	ErrorCouldNotMarshalItem     = "ErrorCouldNotMarshalItem"
	ErrorCouldNotPutItem         = "ErrorCouldNotPutItem"
	EmailAlreadyExists           = "EmailAlreadyExists"
	EmailDoesNotExists           = "EmailDoesNotExists"
	Error                        = "ErrorDeleting"
)

type User struct {
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

func isEmailValid(email string) bool {
	var rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]{1,64}@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if len(email) < 3 || len(email) > 256 || !rxEmail.MatchString(email) {
		return false
	}
	return true
}

func FetchUser(email string, tableName string, dynamoServer dynamodbiface.DynamoDBAPI) (*User, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(tableName),
	}
	result, err := dynamoServer.GetItem(input)
	if err != nil {
		return nil, errors.New(ErrorMsgCustomer)
	}

	item := new(User)
	err = dynamodbattribute.UnmarshalMap(result.Item, item)
	if err != nil {
		return nil, errors.New(ErrorFailedtoUnmarshalRecord)
	}

	return item, nil
}

func FetchUsers(email string, tableName string, dynamoServer dynamodbiface.DynamoDBAPI) (*[]User, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := dynamoServer.Scan(input)
	if err != nil {
		return nil, errors.New(ErrorFetchingData)
	}

	item := new([]User)
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, item)
	return item, nil
}

func CreateUser(req events.APIGatewayProxyRequest, tableName string, dynamoServer dynamodbiface.DynamoDBAPI) (*User, error) {
	var u User
	if err := json.Unmarshal([]byte(req.Body), &u); err != nil {
		return nil, errors.New(ErrorInvalidUserData)
	}

	if !isEmailValid(u.Email) {
		return nil, errors.New(ErrorInvalidEmail)
	}

	currentUser, _ := FetchUser(u.Email, tableName, dynamoServer)
	if currentUser != nil && len(currentUser.Email) != 0 {
		return nil, errors.New(EmailAlreadyExists)
	}

	av, err := dynamodbattribute.MarshalMap(u)
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = dynamoServer.PutItem(input)
	if err != nil {
		return nil, errors.New(ErrorCouldNotPutItem)
	}
	return &u, nil
}

func UpdateUser(req events.APIGatewayProxyRequest, tableName string, dynamoServer dynamodbiface.DynamoDBAPI) (*User, error) {
	var u User
	if err := json.Unmarshal([]byte(req.Body), &u); err != nil {
		return nil, errors.New(ErrorInvalidUserData)
	}

	currentUser, _ := FetchUser(u.Email, tableName, dynamoServer)
	if currentUser != nil && len(currentUser.Email) == 0 {
		return nil, errors.New(EmailDoesNotExists)
	}

	av, err := dynamodbattribute.MarshalMap(u)
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = dynamoServer.PutItem(input)
	if err != nil {
		return nil, errors.New(ErrorCouldNotPutItem)
	}

	return &u, nil
}

func DeleteUser(req events.APIGatewayProxyRequest, tableName string, dynamoServer dynamodbiface.DynamoDBAPI) error {
	email := req.QueryStringParameters["email"]
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(tableName),
	}
	_, err := dynamoServer.DeleteItem(input)
	if err != nil {
		return errors.New(Error)
	}
	return nil
}
