package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3ListObjectsAPI interface {
	ListObjectsV2(ctx context.Context,
		params *s3.ListObjectsV2Input,
		optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

func GetObjects(c context.Context, api S3ListObjectsAPI, input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	return api.ListObjectsV2(c, input)
}

func main() {
	bucket := flag.String("b", "", "The name of the bucket")
	flag.Parse()

	if *bucket == "" {
		fmt.Println("You must supply the name of a bucket (-b BUCKET)")
		return
	}

	http.HandleFunc("/", handleHome(bucket))
	log.Fatal(http.ListenAndServe(":80", nil))
}

func handleHome(b *string) http.HandlerFunc {
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

		fmt.Println("Objects in " + *b + ":")

		for _, item := range resp.Contents {
			fmt.Fprintln(w, "Name:          ", *item.Key)
			fmt.Fprintln(w, "Last modified: ", *item.LastModified)
			fmt.Fprintln(w, "Size:          ", item.Size)
			fmt.Fprintln(w, "Storage class: ", item.StorageClass)
			fmt.Fprintln(w, "")
		}

		fmt.Fprintln(w, "Found", len(resp.Contents), "items in bucket", *b)
		fmt.Fprintln(w, "")
	}
}
