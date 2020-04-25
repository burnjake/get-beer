package main

import (
	"bytes"
	"context"
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

// upload uploads a file to a GCS bucket
func upload(ctx context.Context, r io.Reader, projectID, bucket, name string, public bool) (*storage.ObjectHandle, *storage.ObjectAttrs, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	bh := client.Bucket(bucket)
	// Next check if the bucket exists
	if _, err = bh.Attrs(ctx); err != nil {
		return nil, nil, err
	}

	obj := bh.Object(name)
	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, r); err != nil {
		return nil, nil, err
	}
	if err := w.Close(); err != nil {
		return nil, nil, err
	}

	if public {
		if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
			return nil, nil, err
		}
	}

	attrs, err := obj.Attrs(ctx)
	return obj, attrs, err
}

func objectURL(objAttrs *storage.ObjectAttrs) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", objAttrs.Bucket, objAttrs.Name)
}

func (imageHandler ImageHandler) getMenu(w http.ResponseWriter, req *http.Request) {
	block := Block{
		Type:     "image",
		ImageURL: "https://storage.googleapis.com/mother-kellys-beer/images/pdf200425",
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
	const projectID = "beer-274619"
	const bucket = "mother-kellys-beer"
	var name = "images/pdf" + time.Now().Format("060102")
	ctx := context.Background()

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

	_, attr, err := upload(ctx, png, projectID, bucket, name, true)
	if err != nil {
		panic(err)
	}

	imageHandler := ImageHandler{ImageURL: objectURL(attr)}
	http.HandleFunc("/beer", imageHandler.getMenu)
	http.ListenAndServe(":8090", nil)
}
