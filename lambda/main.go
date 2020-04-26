package main

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Block is a struct representation of a Slack message block
type Block struct {
	Type     string `json:"type"`
	ImageURL string `json:"image_url"`
	AltText  string `json:"alt_text"`
}

// Message is a struct representation of the Slack message payload
type Message struct {
	Blocks []Block `json:"blocks"`
}

// PublicURL contains the public URL of an object
type PublicURL struct {
	PublicURL string `json:"public_url" dynamodbav:"public_url"`
}

func (publicURL PublicURL) getMenu(w http.ResponseWriter, req *http.Request) {
	block := Block{
		Type:     "image",
		ImageURL: publicURL.PublicURL,
		AltText:  "Mother Kelly's Menu",
	}
	response := Message{
		Blocks: []Block{block},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func getLatestImage(sess *session.Session, table string) (PublicURL, error) {
	var publicURL []PublicURL
	svc := dynamodb.New(sess)
	result, err := svc.Query(&dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v1": {
				S: aws.String("stokenewington"),
			},
		},
		KeyConditionExpression: aws.String("bar_location = :v1"),
		ProjectionExpression:   aws.String("public_url"),
		ScanIndexForward:       aws.Bool(false),
		Limit:                  aws.Int64(1),
		TableName:              aws.String(table),
	})
	if err != nil {
		return publicURL[0], err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &publicURL)
	if err != nil {
		return publicURL[0], err
	}

	return publicURL[0], nil
}

func lambdaHandler() (PublicURL, error) {
	const awsRegion = "eu-west-1"
	const awsProfile = "personal"
	const dynamoDBTable = "MotherKellysMenus"
	const barLocation = "stokenewington"
	var latestImage PublicURL

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewSharedCredentials("", awsProfile),
	})
	if err != nil {
		return latestImage, err
	}

	latestImage, err = getLatestImage(sess, dynamoDBTable)
	if err != nil {
		return latestImage, err
	}

	return latestImage, nil

	// publicURL := PublicURL{PublicURL: latestImage.PublicURL}
	// http.HandleFunc("/beer", publicURL.getMenu)
	// http.ListenAndServe(":8090", nil)
}

func main() {
	lambda.Start(lambdaHandler)
}
