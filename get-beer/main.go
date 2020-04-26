package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gen2brain/go-fitz"
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

// ImageHandler contains image metadata
type ImageHandler struct {
	ImageURL string
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

// getText returns the contents of a given .pdf file as text
func getText(pdf *fitz.Document) (string, error) {
	text, err := pdf.Text(0)
	if err != nil {
		return "", err
	}

	return text, nil
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

// upload uploads a file to an S3 bucket and returns its location
func upload(r io.Reader, region string, profile string, bucket string, key string) (string, error) {
	// Create session
	// sess, err := session.NewSessionWithOptions(session.Options{
	// 	Profile: profile,
	// })
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewSharedCredentials("", profile),
	})
	if err != nil {
		return "", err
	}

	uploader := s3manager.NewUploader(sess)

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   r,
	})
	if err != nil {
		return "", err
	}

	return result.Location, nil
}

func objectURL(objAttrs *storage.ObjectAttrs) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", objAttrs.Bucket, objAttrs.Name)
}

func (imageHandler ImageHandler) getMenu(w http.ResponseWriter, req *http.Request) {
	block := Block{
		Type:     "image",
		ImageURL: imageHandler.ImageURL,
		AltText:  "Mother Kelly's Menu",
	}
	response := Message{
		Blocks: []Block{block},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func main() {
	const pdfEndpoint = "https://motherkellys.co.uk/wp-content/menu/Menu_N16.pdf"
	const downloadDest = "/tmp/beer.pdf"
	const awsRegion = "eu-west-1"
	const awsProfile = "personal"
	const bucket = "mother-kellys"
	var name = fmt.Sprintf("images/stokenewington/pdf-%s.png", time.Now().Format("060102"))

	f := downloadFile(pdfEndpoint, downloadDest)

	pdf, err := fitz.New(f)
	if err != nil {
		panic(err)
	}

	_, err = getText(pdf)
	if err != nil {
		panic(err)
	}

	png, err := getImage(pdf)
	if err != nil {
		panic(err)
	}

	location, err := upload(png, awsRegion, awsProfile, bucket, name)
	if err != nil {
		panic(err)
	}

	imageHandler := ImageHandler{ImageURL: location}
	http.HandleFunc("/beer", imageHandler.getMenu)
	http.ListenAndServe(":8090", nil)
}
