// functions to handle sessions and authentication
package main

import (
	"net/http"
	"time"
	"strings"

	"github.com/google/uuid"
)

// Login is handler for /login renders the login.html
func Login(w http.ResponseWriter, r *http.Request) {

	pageVars := PageVars{}
	addPageVars(r, &pageVars)
	render(w, "login", pageVars)
}


// LoginAction is handler for /loginaction
func LoginAction(w http.ResponseWriter, r *http.Request) {

	pageVars := PageVars{}
	addPageVars(r, &pageVars)

	userName := r.FormValue("userName")
	// validate username
	if len(userName) <= 0 {
		http.Redirect(w, r, "/login?errorM=Username specified", http.StatusSeeOther)
		return
	}
	password := r.FormValue("password")
	// validate password
	if len(password) <= 0 {
		http.Redirect(w, r, "/login?errorM=Password specified", http.StatusSeeOther)
		return
	}

	// get expected password
	expectedPassword := getStringFromDB("Users", userName)

	// verify the password
	if expectedPassword != password {
		http.Redirect(w, r, "/login?errorM=Invalid credentials", http.StatusSeeOther)
		return
	}
	
	// Create a new random session token
	sessionToken, err := uuid.NewUUID()
	if err != nil {
		http.Redirect(w, r, "/login?errorM=Unable to create token", http.StatusSeeOther)
	}
	// Set the token in the db, along with the user whom it represents
	updateDBString("Sessions", userName, sessionToken.String())

	// set the expiration time
	expires := time.Now().Add(300 * time.Second)
	ck := http.Cookie{
        Name: "JSESSION_ID",
        Domain: "localhost",
        Path: "/",
		Expires: expires,
		Value: userName+"_"+sessionToken.String(),
	}

    // write the cookie to response
    http.SetCookie(w, &ck)
	http.Redirect(w, r, "/listbuckets", http.StatusSeeOther)

}

// A closure function to handle all sessions
func validateSession(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {  
	return func(w http.ResponseWriter, r *http.Request) {
		// get cookie from request
	 	c, err := r.Cookie("JSESSION_ID")
		if err != nil {
			if err == http.ErrNoCookie {		
				http.Redirect(w, r, "/login?errorM=No session present in request", http.StatusSeeOther)
				return	
			}
			http.Redirect(w, r, "/login?errorM=Not authorised", http.StatusSeeOther)
			return
		}
		// if no errors, validate the cookie
		sessToken := c.Value
		sv := strings.Split(sessToken, "_")
		if len(sv) != 2 {		
			http.Redirect(w, r, "/login?errorM=Invalid cookie format", http.StatusSeeOther)
			return	
		}
		expSessToken := getStringFromDB("Sessions", sv[0])
		if sv[1] != expSessToken {
			http.Redirect(w, r, "/login?Invalid session", http.StatusSeeOther)		
			return
		}
		// if sucess process the handler		
		f(w, r)
	}
}
  