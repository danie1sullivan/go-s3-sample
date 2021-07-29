package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3ListObjectsAPI interface {
	ListObjectsV2(ctx context.Context,
		params *s3.ListObjectsV2Input,
		optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

type S3PutObjectAPI interface {
	PutObject(ctx context.Context,
		params *s3.PutObjectInput,
		optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type S3DeleteObjectAPI interface {
	DeleteObject(ctx context.Context,
		params *s3.DeleteObjectInput,
		optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

func GetObjects(c context.Context, api S3ListObjectsAPI, input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	return api.ListObjectsV2(c, input)
}

func PutFile(c context.Context, api S3PutObjectAPI, input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return api.PutObject(c, input)
}

func DeleteItem(c context.Context, api S3DeleteObjectAPI, input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return api.DeleteObject(c, input)
}

func main() {
	bucket := flag.String("b", "", "The name of the bucket")
	flag.Parse()

	if *bucket == "" {
		fmt.Println("You must supply the name of a bucket (-b BUCKET)")
		return
	}

	http.HandleFunc("/", handleHealth)
	http.HandleFunc("/s3", handleS3(bucket))
	http.HandleFunc("/s3/add", handleS3Add(bucket))
	http.HandleFunc("/s3/delete", handleS3Delete(bucket))
	http.HandleFunc("/health", handleHealth)
	log.Fatal(http.ListenAndServe(":80", nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleS3Delete(b *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			panic("configuration error, " + err.Error())
		}

		client := s3.NewFromConfig(cfg)

		objectName := "Hello-World"
		input := &s3.DeleteObjectInput{
			Bucket: b,
			Key:    &objectName,
		}

		_, err = DeleteItem(context.TODO(), client, input)
		if err != nil {
			fmt.Println("Got an error deleting item:")
			fmt.Println(err)
			return
		}

		fmt.Println("Deleted " + objectName + " from " + *b)
		http.Redirect(w, r, "/s3", http.StatusSeeOther)
	}
}

func handleS3Add(b *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			panic("configuration error, " + err.Error())
		}

		client := s3.NewFromConfig(cfg)

		filename := "Hello-World"
		input := &s3.PutObjectInput{
			Bucket: b,
			Key:    &filename,
			Body:   strings.NewReader("Hello World"),
		}

		_, err = PutFile(context.TODO(), client, input)
		if err != nil {
			fmt.Println("Got error uploading file:")
			fmt.Println(err)
			return
		}
		http.Redirect(w, r, "/s3", http.StatusSeeOther)
	}
}

func handleS3(b *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			panic("configuration error, " + err.Error())
		}

		client := s3.NewFromConfig(cfg)

		input := &s3.ListObjectsV2Input{
			Bucket: b,
		}

		resp, err := GetObjects(context.TODO(), client, input)
		if err != nil {
			fmt.Println("Got error retrieving list of objects:")
			fmt.Println(err)
			return
		}

		fmt.Fprintln(w, "<html>")
		fmt.Fprintln(w, "Objects in "+*b+":<br>")

		for _, item := range resp.Contents {
			fmt.Fprintln(w, "Name:          ", *item.Key, "<br>")
			fmt.Fprintln(w, "Last modified: ", *item.LastModified, "<br>")
			fmt.Fprintln(w, "Size:          ", item.Size, "<br>")
			fmt.Fprintln(w, "Storage class: ", item.StorageClass, "<br>")
			fmt.Fprintln(w, "")
		}

		fmt.Fprintln(w, "Found", len(resp.Contents), "items in bucket", *b, "<br>")
		fmt.Fprintln(w, "<br>")
		fmt.Fprintln(w, "<a href='/s3/add'>add</a><br>")
		fmt.Fprintln(w, "<a href='/s3/delete'>/elete</a><br>")
		fmt.Fprintln(w, "<br>")
		fmt.Fprintln(w, "</html>")
	}
}
