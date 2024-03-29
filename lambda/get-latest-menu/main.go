package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var errorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)
var awsRegion string
var dynamoDBTable string
var hashKey string

// Block is a struct representation of a Slack message block
type Block struct {
	Type     string `json:"type"`
	ImageURL string `json:"image_url"`
	AltText  string `json:"alt_text"`
}

// Message is a struct representation of the Slack message payload
type Message struct {
	ResonseType string  `json:"response_type"`
	Blocks      []Block `json:"blocks"`
}

// PublicURL contains the public URL of an object
type PublicURL struct {
	PublicURL string `json:"public_url" dynamodbav:"public_url"`
}

func getLatestImage(table string, region string, hashKeyVal string) (PublicURL, error) {
	var queryResponse []PublicURL
	// Empty struct required so that there is always a valid variable to return when error handling
	var publicURL PublicURL

	client := dynamodb.New(session.Must(session.NewSession()))
	result, err := client.Query(&dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v1": {
				S: aws.String(hashKeyVal),
			},
		},
		KeyConditionExpression: aws.String("bar_location = :v1"),
		ProjectionExpression:   aws.String("public_url"),
		ScanIndexForward:       aws.Bool(false),
		Limit:                  aws.Int64(1),
		TableName:              aws.String(table),
	})
	if err != nil {
		return publicURL, err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &queryResponse)
	if err != nil {
		return publicURL, err
	}

	return queryResponse[0], nil
}

func httpError(status int) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}
}

func constructResponse(url string) (events.APIGatewayProxyResponse, error) {
	block := Block{
		Type:     "image",
		ImageURL: url,
		AltText:  "Mother Kelly's Menu",
	}

	message := Message{
		ResonseType: "in_channel",
		Blocks:      []Block{block},
	}

	responseBody, err := json.Marshal(message)
	if err != nil {
		errorLogger.Println(err.Error())
		return httpError(http.StatusInternalServerError), nil
	}

	// All good if we got to here!
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(responseBody),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func lambdaHandler(events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	latestImageURL, err := getLatestImage(dynamoDBTable, awsRegion, hashKey)
	if err != nil {
		errorLogger.Println(err.Error())
		return httpError(http.StatusInternalServerError), nil
	}
	if latestImageURL.PublicURL == "" {
		return httpError(http.StatusNotFound), nil
	}

	response, err := constructResponse(latestImageURL.PublicURL)
	if err != nil {
		errorLogger.Println(err.Error())
		return httpError(http.StatusInternalServerError), nil
	}

	return response, nil
}

func main() {
	lambda.Start(lambdaHandler)
}
