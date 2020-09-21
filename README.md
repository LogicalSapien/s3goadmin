# s3goadmin
An S3 Admin utility witten in Goland

The utility has following features:

- Create Bucket
- List Buckets
- Delete Bucket
- Upload Object
- Download Object
- Delete Object
- List Objects
- Search Objects using prefix & delimited
- Folder navigation
- User authenitcation for s3goadmin
- Session management for s3goadmin

To get started, install the aws dependencies to start

 `go get -u github.com/aws/aws-sdk-go`

 `go get github.com/google/uuid`

 `go get -u  go get github.com/boltdb/bolt/`

To build & run locally:

 `go build -o main`
 
 `export ACCESS_KEY=__________`
 
 `export SECRET_ACCESS=__________`

 `./main`

Docker build command

 `docker build -t s3goadmin .`

