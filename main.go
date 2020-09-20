package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
	"os"

	"github.com/aws/aws-sdk-go/service/s3"
)

// PageVars structs
type PageVars struct {
	BList    []*s3.Bucket
	OList    []ObjectDetails
	// Bucket name
	BName    string
	// File name
	FName    string
	Prefix string
	Delimiter string
	// Folder list
	FList []FolderDetails
	// Folder count
	FCount int
	ErrorM   string
	SuccessM string
}

// ObjectDetails struct
type ObjectDetails struct {
	Key string
	Name string
	LastModified time.Time
	Size int64
	StorageClass string
	// Folder, File
	Type string
}

// FolderDetails struct
type FolderDetails struct {
	Name string
	// Prefix until the current folder
	PrevPrefix string
}

// AwsCred for connection
type AwsCred struct {
	Akey   string
	Skey   string
	Region string
}

var tpl *template.Template

func init() {
	//parse the template file held in the templates folder
	//add the custom add function to template
	funcs := template.FuncMap{"add": add}
	tpl = template.Must(template.New("*").Funcs(funcs).ParseGlob("templates/*"))

	// create in memory dba nd intialize with admin user
	createDb()
	// create aws session
	createSession()
}

func main() {	

	// serve everything in the css folder, the img folder and mp3 folder as a file
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("img"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))

	// when navigating to /home it should serve the home page
	http.HandleFunc("/", Login)
	http.HandleFunc("/loginaction", LoginAction)
	http.HandleFunc("/listbuckets", validateSession(ListBuckets))
	http.HandleFunc("/objectlist", validateSession(ListObjects))
	http.HandleFunc("/uploadfile", validateSession(UploadFile))
	http.HandleFunc("/uploadaction", validateSession(UploadAction))
	http.HandleFunc("/downloadfileaction", validateSession(DownloadFileAction))
	http.HandleFunc("/deleteobjectaction", validateSession(DeleteObjectAction))
	http.HandleFunc("/createbucket", validateSession(CreateBucket))
	http.HandleFunc("/createbucketaction", validateSession(CreateBucketAction))
	http.HandleFunc("/deletebucketaction", validateSession(DeleteBucketAction))	
	http.HandleFunc("/createfolder", validateSession(CreateFolder))
	http.HandleFunc("/createfolderaction", validateSession(CreateFolderAction))
	http.HandleFunc("/logoutaction", LogoutAction)	
	
	http.ListenAndServe(getPort(), nil)

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

func addPageVars(r *http.Request, p *PageVars) {
	bucketName := r.URL.Query().Get("bucketName")
	if len(bucketName) > 0 {
		p.BName = bucketName
	}
	fileName := r.URL.Query().Get("fileName")
	if len(fileName) > 0 {
		p.FName = fileName
	}
	prefix := r.URL.Query().Get("prefix")
	if len(prefix) > 0 {
		p.Prefix = prefix
	}
	delimiter := r.URL.Query().Get("delimiter")
	if len(delimiter) > 0 {
		p.Delimiter = delimiter
	}
	errorM := r.URL.Query().Get("errorM")
	if len(errorM) > 0 {
		p.ErrorM = errorM
	}
	successM := r.URL.Query().Get("successM")
	if len(successM) > 0 {
		p.SuccessM = successM
	}
}
