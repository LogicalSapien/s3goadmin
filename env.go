// fucntions to manage env/flag variables
package main

import (
	"os"
)

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
	}
	// Get secret key
	skey := os.Getenv("SECRET_ACCESS")
	if akey != "" {
		c.Skey = skey
	}
	// region
	reg := os.Getenv("REGION")

	if reg != "" {
		c.Region = reg
	} else {
		c.Region = "us-east-1"
	}
	return c
}

func getAdminPass() string {
	// Get admin pass
	apass := os.Getenv("ADMIN_PASS")
	if apass != "" {
		return apass
	}
	return "password"
}
