package main

import (
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/prtka/go-serverless-api/pkg/handlers"
)

var (
	dynamoServer dynamodbiface.DynamoDBAPI
)

func main() {
	region := os.Getenv("AWS_REGION")
	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(region)})
	if err != nil {
		return
	}
	dynamoServer = dynamodb.New(awsSession)
	lambda.Start(handler)
}

const tableName = "go-serverless-test-table"

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		return handlers.GetUser(req, tableName, dynamoServer)
	case "POST":
		return handlers.CreateUser(req, tableName, dynamoServer)
	case "PUT":
		return handlers.UpdateUser(req, tableName, dynamoServer)
	case "DELETE":
		return handlers.DeleteUser(req, tableName, dynamoServer)
	default:
		return handlers.UnhandledMethodResponse(req, tableName, dynamoServer)
	}

}
