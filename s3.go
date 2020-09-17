package main

import (
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// ListBuckets ishandler for / renders the bucketlist.html
func ListBuckets(w http.ResponseWriter, r *http.Request) {

	svc := s3.New(sess)

	// get bucket list fomr s3 api
	result, err := svc.ListBuckets(nil)
	pageVars := PageVars{}
	if err != nil {
		pageVars.ErrorM = "Failed to load buckets list"
	} else {
		pageVars.BList = result.Buckets
	}

	render(w, "bucketlist", pageVars)
}

// GetObjects for listing objects
func GetObjects(w http.ResponseWriter, r *http.Request) {

	svc := s3.New(sess)

	pageVars := PageVars{}
	addPageVars(r, &pageVars)

	if len(pageVars.BName) <= 0 {
		pageVars.ErrorM = "Invalid bucket name"
		render(w, "objectlist", pageVars)
	} else {
		bucket := aws.String(pageVars.BName)

		// Get the list of items
		resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: bucket})

		if err != nil {
			pageVars.ErrorM = "Failed to get objects"
		} else {
			pageVars.OList = resp.Contents
		}

		render(w, "objectlist", pageVars)
	}

}

// PutFile to upload file tot aws s3
func PutFile(w http.ResponseWriter, r *http.Request) {

	pageVars := PageVars{}
	addPageVars(r, &pageVars)

	if len(pageVars.BName) <= 0 {
		http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&errorM=Invalid bucket namee", http.StatusSeeOther)
	} else {
		bucket := aws.String(pageVars.BName)
		// Maximum upload of 1024 MB files
		r.ParseMultipartForm(1024 << 20)

		// Get handler for filename, size and headers
		file, handler, err := r.FormFile("uploadfile")

		// close file after func
		defer file.Close()

		if err != nil {
			http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&errorM=Error uploading the file", http.StatusSeeOther)
		} else {
			filename := aws.String(handler.Filename)

			uploader := s3manager.NewUploader(sess)

			_, err = uploader.Upload(&s3manager.UploadInput{
				Bucket: bucket,
				Key:    filename,
				Body:   file,
			})

			if err != nil {
				http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&errorM=Error in uploading to S3", http.StatusSeeOther)
			} else {
				http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&successM=Successfully uploaded", http.StatusSeeOther)
			}
		}
	}

}

// UploadFile is handler for /uploadfile renders the uploadfile.html
func UploadFile(w http.ResponseWriter, r *http.Request) {

	bp := r.URL.Query().Get("bucketName")

	pageVars := PageVars{
		BName: bp,
	}

	render(w, "uploadfile", pageVars)
}

// DownloadFile is handler for /downloadfile
func DownloadFile(w http.ResponseWriter, r *http.Request) {

	bp := r.URL.Query().Get("bucketName")
	fp := r.URL.Query().Get("fileName")

	bucket := aws.String(bp)
	filename := aws.String(fp)

	file, err := os.Create(*filename)

	pageVars := PageVars{}
	if err != nil {
		pageVars.ErrorM = "Error in downloading"
		render(w, "objectlist", pageVars)
	} else {
		defer file.Close()

		downloader := s3manager.NewDownloader(sess)

		_, err = downloader.Download(file,
			&s3.GetObjectInput{
				Bucket: bucket,
				Key:    filename,
			})

		if err != nil {
			pageVars.ErrorM = "Failed to get object"
			render(w, "objectlist", pageVars)
		} else {
			//copy the relevant headers. If you want to preserve the downloaded file name, extract it with go's url parser.
			w.Header().Set("Content-Disposition", "attachment; filename="+fp)
			w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
			w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

			//stream the body to the client without fully loading it into memory
			io.Copy(w, file)
		}
	}

}
