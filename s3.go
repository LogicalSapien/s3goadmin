package main

import (
	"io"
	"net/http"
	"os"
	"strings"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var sess *session.Session

func createSession() {
	// Initialize a session in provided region that the SDK will use to load
	// get credentials
	c := getAwsCred()
	// credentials can also be in ~/.aws/credentials.
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(c.Region),
		Credentials: credentials.NewStaticCredentials(c.Akey, c.Skey, "")},
	)

	if err != nil {
		exitErrorf("Unable to connect to Server, %v", err)
	}

	// assign session to global variable
	sess = s
}

// ListBuckets ishandler for / renders the bucketlist.html
func ListBuckets(w http.ResponseWriter, r *http.Request) {

	svc := s3.New(sess)

	pageVars := PageVars{}
	addPageVars(r, &pageVars)

	// get bucket list fomr s3 api
	result, err := svc.ListBuckets(nil)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			pageVars.ErrorM = awsErr.Message()
		} else {
			pageVars.ErrorM = "Failed to load buckets list"
		}
	} else {
		pageVars.BList = result.Buckets
	}

	render(w, "bucketlist", pageVars)
}

// ListObjects for listing objects
func ListObjects(w http.ResponseWriter, r *http.Request) {

	svc := s3.New(sess)

	pageVars := PageVars{}
	addPageVars(r, &pageVars)

	if len(pageVars.BName) <= 0 {
		if len(pageVars.ErrorM) <= 0 {
			pageVars.ErrorM = "Invalid bucket name"
		}
		render(w, "objectlist", pageVars)
	} else {

		if len(pageVars.Prefix) <= 0 {
			pageVars.Prefix = ""
		}

		if len(pageVars.Delimiter) <= 0 {
			pageVars.Delimiter = ""
		}

		// Get the list of items
		resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket: aws.String(pageVars.BName),
			Prefix: aws.String(pageVars.Prefix),
			Delimiter: aws.String(pageVars.Delimiter),
		})

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				pageVars.ErrorM = awsErr.Message()
			} else {
				pageVars.ErrorM = "Failed to get objects"
			}
		} else {
			for _, o := range resp.Contents {
				var shouldAdd = true
				// check how many / in prefix
				ns := strings.Count(pageVars.Prefix, "/")
				// check if its folder
				if strings.HasSuffix(*o.Key, "/") {					
					// check if its top level folder
					if strings.Count(*o.Key, "/") == (ns + 1) {
						shouldAdd = true
					} else {
						shouldAdd = false
					}
				} else {
					// check if its top level folder
					if strings.Count(*o.Key, "/") == ns {
						shouldAdd = true
					} else {
						shouldAdd = false
					}
				}				

				if shouldAdd {
					name := *o.Key					
					if len(pageVars.Prefix) > 0 {
						name = strings.Replace(name, pageVars.Prefix, "", -1)
					}
					if strings.HasSuffix(*o.Key, "/") {
						od := ObjectDetails{*o.Key, name, *o.LastModified, *o.Size, *o.StorageClass, "Folder"}
						pageVars.OList = append(pageVars.OList, od)
					} else {
						od := ObjectDetails{*o.Key, name, *o.LastModified, *o.Size, *o.StorageClass, "File"}
						pageVars.OList = append(pageVars.OList, od)
					}														
				}				
			}
			// add folder names f prefix ends with /
			sl := strings.Split(pageVars.Prefix, "/")
			pp := ""
			// remove last element as its empy due to trailing /
			if len(sl) > 0 {
				sl = sl[:len(sl)-1]
				for _, fld := range sl {
					pp = pp+fld+"/"
					pageVars.FList = append(pageVars.FList, FolderDetails{fld, pp})					
				}
				
				pageVars.FCount = len(pageVars.FList)
			}
			
		}

		render(w, "objectlist", pageVars)
	}

}

// UploadAction to upload file to aws s3
func UploadAction(w http.ResponseWriter, r *http.Request) {

	pageVars := PageVars{}
	addPageVars(r, &pageVars)

	if len(pageVars.BName) <= 0 {
		http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&prefix="+pageVars.Prefix+"&errorM=Invalid bucket name", http.StatusSeeOther)
	} else {
		bucket := aws.String(pageVars.BName)
		// Maximum upload of 1024 MB files
		r.ParseMultipartForm(1024 << 20)

		// Get handler for filename, size and headers
		file, handler, err := r.FormFile("uploadfile")

		// close file after func
		defer file.Close()

		if err != nil {
			http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&prefix="+pageVars.Prefix+"&errorM=Error uploading the file", http.StatusSeeOther)
		} else {
			fn := handler.Filename
			if len(pageVars.Prefix) > 0 {
				fn = pageVars.Prefix + "/" + fn
			}
			filename := aws.String(fn)

			uploader := s3manager.NewUploader(sess)

			_, err = uploader.Upload(&s3manager.UploadInput{
				Bucket: bucket,
				Key:    filename,
				Body:   file,
			})

			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&prefix="+pageVars.Prefix+"&errorM="+awsErr.Message(), http.StatusSeeOther)
				} else {
					http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&prefix="+pageVars.Prefix+"&errorM=Error in uploading to S3", http.StatusSeeOther)
				}
			} else {
				http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&prefix="+pageVars.Prefix+"&successM=Successfully uploaded", http.StatusSeeOther)
			}
		}
	}

}

// UploadFile is handler for /uploadfile renders the uploadfile.html
func UploadFile(w http.ResponseWriter, r *http.Request) {

	pageVars := PageVars{}
	addPageVars(r, &pageVars)

	// add folder names f prefix ends with /
	sl := strings.Split(pageVars.Prefix, "/")
	pp := ""
	// remove last element as its empy due to trailing /
	if len(sl) > 0 {
		sl = sl[:len(sl)-1]
		for _, fld := range sl {
			pp = pp+fld+"/"
			pageVars.FList = append(pageVars.FList, FolderDetails{fld, pp})					
		}
		
		pageVars.FCount = len(pageVars.FList)
	}

	render(w, "uploadfile", pageVars)
}

// DownloadFileAction is handler for /downloadfile
func DownloadFileAction(w http.ResponseWriter, r *http.Request) {

	pageVars := PageVars{}
	addPageVars(r, &pageVars)

	bucket := aws.String(pageVars.BName)
	filename := aws.String(pageVars.FName)
	filenameReplaced := aws.String(strings.Replace(pageVars.FName, "/", "_", -1))

	if len(pageVars.BName) <= 0 {
		http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&prefix="+pageVars.Prefix+"&errorM=Invalid bucket name", http.StatusSeeOther)
	} else if len(pageVars.FName) <= 0 {
		http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&prefix="+pageVars.Prefix+"&errorM=Invalid file name", http.StatusSeeOther)
	} else {

		file, err := os.Create(*filenameReplaced)

		if err != nil {
			http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&prefix="+pageVars.Prefix+"&errorM=Error in downloading", http.StatusSeeOther)
		} else {
			downloader := s3manager.NewDownloader(sess)

			_, err = downloader.Download(file,
				&s3.GetObjectInput{
					Bucket: bucket,
					Key:    filename,
				})

			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&prefix="+pageVars.Prefix+"&errorM="+awsErr.Message(), http.StatusSeeOther)
				} else {
					http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&prefix="+pageVars.Prefix+"&errorM=Failed to get object", http.StatusSeeOther)
				}
			} else {
				//copy the relevant headers. If you want to preserve the downloaded file name, extract it with go's url parser.
				w.Header().Set("Content-Disposition", "attachment; filename="+pageVars.FName)
				w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
				w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

				//stream the body to the client without fully loading it into memory
				io.Copy(w, file)
			}
			file.Close()
			err = os.Remove(*filenameReplaced)
		}
	}

}

// DeleteObjectAction handler to delete items in s3
func DeleteObjectAction(w http.ResponseWriter, r *http.Request) {

	svc := s3.New(sess)

	pageVars := PageVars{}
	addPageVars(r, &pageVars)

	if len(pageVars.BName) <= 0 {
		http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&errorM=Invalid bucket name", http.StatusSeeOther)
	} else if len(pageVars.FName) <= 0 {
		http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&errorM=Invalid file name", http.StatusSeeOther)
	} else {
		bucket := aws.String(pageVars.BName)
		item := aws.String(pageVars.FName)

		_, err := svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: bucket,
			Key:    item,
		})

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&errorM="+awsErr.Message(), http.StatusSeeOther)
			} else {
				http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&errorM=Failed to delete", http.StatusSeeOther)
			}
		} else {
			err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
				Bucket: bucket,
				Key:    item,
			})
			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&errorM="+awsErr.Message(), http.StatusSeeOther)
				} else {
					http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&errorM=Failed to delete", http.StatusSeeOther)
				}
			} else {
				http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&successM=Successfully deleted", http.StatusSeeOther)
			}
		}

	}
}

// CreateBucket is handler for /createbucket renders the createbucket.html
func CreateBucket(w http.ResponseWriter, r *http.Request) {

	pageVars := PageVars{}
	addPageVars(r, &pageVars)

	render(w, "createbucket", pageVars)
}

// CreateBucketAction is handler for crreating a bucket
func CreateBucketAction(w http.ResponseWriter, r *http.Request) {

	// Create S3 service client
	svc := s3.New(sess)

	bucket := r.FormValue("bucketName")

	if len(bucket) <= 0 {
		http.Redirect(w, r, "/createbucket?errorM=No bucket name specified", http.StatusSeeOther)
	} else {
		// Create the S3 Bucket
		_, err := svc.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				// process SDK error
				http.Redirect(w, r, "/createbucket?errorM="+awsErr.Message(), http.StatusSeeOther)
			} else {
				http.Redirect(w, r, "/createbucket?errorM=Failed to create bucket", http.StatusSeeOther)
			}
		} else {
			http.Redirect(w, r, "/bucketlist?successM=Bucket created succcesfully", http.StatusSeeOther)
		}
	}

}

// DeleteBucketAction handler to delete bucket in s3
func DeleteBucketAction(w http.ResponseWriter, r *http.Request) {

	svc := s3.New(sess)

	pageVars := PageVars{}
	addPageVars(r, &pageVars)

	if len(pageVars.BName) <= 0 {
		http.Redirect(w, r, "/bucketlist?bucketName="+pageVars.BName+"&errorM=Invalid bucket name", http.StatusSeeOther)
	} else {
		_, err := svc.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(pageVars.BName),
		})

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				http.Redirect(w, r, "/bucketlist?errorM="+awsErr.Message(), http.StatusSeeOther)
			} else {
				http.Redirect(w, r, "/bucketlist?errorM=Failed to delete bucket", http.StatusSeeOther)
			}
		} else {
			err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
				Bucket: aws.String(pageVars.BName),
			})
			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					// process SDK error
					http.Redirect(w, r, "/bucketlist?errorM="+awsErr.Message(), http.StatusSeeOther)
				} else {
					http.Redirect(w, r, "/bucketlist?errorM=Failed to delete bucket", http.StatusSeeOther)
				}
			} else {
				http.Redirect(w, r, "/bucketlist?successM=Successfully deleted", http.StatusSeeOther)
			}
		}

	}
}

// CreateFolder is handler for /createfolder renders the createfolder.html
func CreateFolder(w http.ResponseWriter, r *http.Request) {

	pageVars := PageVars{}
	addPageVars(r, &pageVars)

	// add folder names f prefix ends with /
	sl := strings.Split(pageVars.Prefix, "/")
	pp := ""
	// remove last element as its empy due to trailing /
	if len(sl) > 0 {
		sl = sl[:len(sl)-1]
		for _, fld := range sl {
			pp = pp+fld+"/"
			pageVars.FList = append(pageVars.FList, FolderDetails{fld, pp})					
		}
		
		pageVars.FCount = len(pageVars.FList)
	}

	render(w, "createfolder", pageVars)
}

// CreateFolderAction to create folders in aws s3
func CreateFolderAction(w http.ResponseWriter, r *http.Request) {

	pageVars := PageVars{}
	addPageVars(r, &pageVars)

	folderName := r.FormValue("folderName")

	if len(pageVars.BName) <= 0 {
		http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&prefix="+pageVars.Prefix+"&errorM=Invalid bucket name", http.StatusSeeOther)
	} else if len(folderName) <= 0 {
		http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&prefix="+pageVars.Prefix+"&errorM=Invalid folder name", http.StatusSeeOther)
	} else {
		bucket := aws.String(pageVars.BName)

		if (len(pageVars.Prefix) > 0) && strings.HasSuffix(pageVars.Prefix, "/") {
			folderName = pageVars.Prefix + folderName
		}

		svc := s3.New(sess)

		_, err := svc.PutObject(&s3.PutObjectInput{
			Bucket: bucket,
			Key:    aws.String(folderName + "/"),	
			ContentLength: aws.Int64(0),
		})

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&prefix="+pageVars.Prefix+"&errorM="+awsErr.Message(), http.StatusSeeOther)
			} else {
				http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&prefix="+pageVars.Prefix+"&errorM=Error in creating folder", http.StatusSeeOther)
			}
		} else {
			http.Redirect(w, r, "/objectlist?bucketName="+pageVars.BName+"&prefix="+pageVars.Prefix+"&successM=Successfully created", http.StatusSeeOther)
		}
	}

}