package main

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// ListBuckets ishandler for / renders the bucketlist.html
func ListBuckets(w http.ResponseWriter, r *http.Request) {

	// get bucket list fomr s3 api
	result, err := s3Svc.ListBuckets(nil)
	if err != nil {
		exitErrorf("Unable to list buckets, %v", err)
	}

	pageVars := PageVars{
		BList: result.Buckets,
	}

	render(w, "bucketlist", pageVars)
}

// GetObjects for listing objects
func GetObjects(w http.ResponseWriter, r *http.Request) {

	bp := r.URL.Query().Get("bucketName")

	bucket := aws.String(bp)

	// Get the list of items
	resp, err := s3Svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: bucket})
	if err != nil {
		exitErrorf("Unable to list objects, %v", err)
	}

	pageVars := PageVars{
		OList: resp.Contents,
		BName: bp,
	}

	render(w, "objectlist", pageVars)
}
