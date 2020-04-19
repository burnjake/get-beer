package main

import (
	"bytes"
	"context"
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

// Image is a simple struct to contain public url string
type Image struct {
	URL string
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

// func getMenu(w http.ResponseWriter, req *http.Request) {
// 	image := Image{URL: ""}
// }

func main() {
	// const beerEndpoint = "https://motherkellys.co.uk/wp-content/menu/Menu_SE1.pdf"
	const downloadDest = "/tmp/beer.pdf"
	const beerPdf = "/Users/jakeburn/Documents/Repos/github_personal/get-beer/resources/beer_190721.pdf"
	const projectID = "beer-274619"
	const bucket = "mother-kellys-beer"
	var name = "images/beer-" + time.Now().Format("060102")
	ctx := context.Background()

	// f := downloadFile(beerEndpoint, downloadDest)

	pdf, err := fitz.New(beerPdf)
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

	fmt.Println(objectURL(attr))

	// for n := 0; n < doc.NumPage(); n++ {
	// 	text, err := doc.Text(n)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println(text)
	// }

}
