package main

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gen2brain/go-fitz"
)

// ImageMetadata contains image metadata
type ImageMetadata struct {
	PublicURL   string `json:"public_url"`
	Created     string `json:"created"`
	BarLocation string `json:"bar_location"`
}

// downloadFile downloads the contents of a file and saves it to a specified location.
// Returns the full path of the file.
func downloadFile(url string, dest string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	f, err := ioutil.TempFile(os.TempDir(), "beer.pdf")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	io.Copy(f, resp.Body)

	return f.Name()
}

// getImage returns a buffer containing the png encoded contents of a give pdf file
func getImage(pdf *fitz.Document) (*bytes.Buffer, error) {
	img, err := pdf.Image(0)
	if err != nil {
		return nil, err
	}

	buff := new(bytes.Buffer)

	err = png.Encode(buff, img)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

// uploadToS3 uploads a file to an S3 bucket and returns its location
func uploadToS3(sess *session.Session, r io.Reader, bucket string, key string) (string, error) {
	svc := s3manager.NewUploader(sess)

	// Upload the file to S3 with public read ACL
	result, err := svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   r,
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return "", err
	}

	return result.Location, nil
}

func saveImageMetadata(sess *session.Session, table string, record ImageMetadata) error {
	svc := dynamodb.New(sess)

	av, err := dynamodbattribute.MarshalMap(record)
	if err != nil {
		return err
	}

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(table),
	})
	if err != nil {
		return err
	}

	return nil
}

func main() {
	const pdfEndpoint = "https://motherkellys.co.uk/wp-content/menu/Menu_N16.pdf"
	const downloadDest = "/tmp/beer.pdf"
	const awsRegion = "eu-west-1"
	const awsProfile = "personal"
	const bucket = "mother-kellys"
	const dynamoDBTable = "MotherKellysMenus"
	const barLocation = "stokenewington"
	var date = time.Now().Format("06-01-02")
	var name = fmt.Sprintf("images/%s/pdf-%s.png", barLocation, date)

	f := downloadFile(pdfEndpoint, downloadDest)

	pdf, err := fitz.New(f)
	if err != nil {
		panic(err)
	}

	png, err := getImage(pdf)
	if err != nil {
		panic(err)
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewSharedCredentials("", awsProfile),
	})
	if err != nil {
		panic(err)
	}

	location, err := uploadToS3(sess, png, bucket, name)
	if err != nil {
		panic(err)
	}

	imageMetadata := ImageMetadata{
		PublicURL:   location,
		Created:     date,
		BarLocation: barLocation,
	}

	err = saveImageMetadata(sess, dynamoDBTable, imageMetadata)
	if err != nil {
		panic(err)
	}
}
