package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// PageVars structs
type PageVars struct {
	BList []*s3.Bucket
	OList []*s3.Object
	BName string
}

// AwsCred for connection
type AwsCred struct {
	Akey   string
	Skey   string
	Region string
}

var s3Svc *s3.S3
var tpl *template.Template

func init() {
	//parse the template file held in the templates folder

	//add the custom add function to template
	funcs := template.FuncMap{"add": add}
	tpl = template.Must(template.New("*").Funcs(funcs).ParseGlob("templates/*"))
}

func main() {

	// initialize s3 client
	s3Svc = getS3Svc()

	// serve everything in the css folder, the img folder and mp3 folder as a file
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("img"))))

	// when navigating to /home it should serve the home page
	http.HandleFunc("/", ListBuckets)
	http.HandleFunc("/objectlist", GetObjects)
	http.ListenAndServe(getPort(), nil)

}

// Detect $PORT and if present uses it for listen and serve else defaults to :8080
// This is so that app can run on Heroku
func getPort() string {
	p := os.Getenv("PORT")
	if p != "" {
		return ":" + p
	}
	return ":8080"
}

func getAwsCred() AwsCred {
	c := AwsCred{}
	// Get access Key
	akey := os.Getenv("ACCESS_KEY")
	if akey != "" {
		c.Akey = akey
	} else {
		c.Akey = ""
	}
	// Get secret key
	skey := os.Getenv("SECRET_ACCESS")
	if akey != "" {
		c.Skey = skey
	} else {
		c.Skey = ""
	}
	// region
	reg := os.Getenv("REGION")

	if akey != "" {
		c.Region = reg
	} else {
		c.Region = "us-east-1"
	}
	return c
}

func getS3Svc() *s3.S3 {
	// Initialize a session in provided region that the SDK will use to load
	// get credentials
	c := getAwsCred()
	// credentials can also be in ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(c.Region),
		Credentials: credentials.NewStaticCredentials(c.Akey, c.Skey, "")},
	)

	if err != nil {
		exitErrorf("Unable to connect to Server, %v", err)
	}

	// Create S3 service client
	svc := s3.New(sess)

	return svc
}

// for rendeing templates
func render(w http.ResponseWriter, tmpl string, pageVars PageVars) {

	err := tpl.ExecuteTemplate(w, tmpl, pageVars) //execute the template and pass in the variables to fill the gaps

	if err != nil { // if there is an error
		log.Fatalln(err) //log it
	}
}

func add(x, y int) int {
	return x + y
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
