// REST/JSON server for managing users and scores.
package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/xyproto/instapage"
	"github.com/xyproto/permissions2"
)

const (
	Version = "1.0"
	API     = "/api/" + Version + "/"
	Title   = "Score Server " + Version
)

var (
	AdminUsername = "admin"
)

type RegisterAdmin struct {
	Password1 string `form:"password1" binding:"required"`
	Password2 string `form:"password2" binding:"required"`
	Email     string `form:"email" binding:"required"`
}

type LoginAdmin struct {
	Password string `form:"password" binding:"required"`
}

var yesnomap = map[bool]string{true: "yes", false: "no"}

// Helper function for converting a bool to "yes" or "no"
func b2yn(b bool) string {
	return yesnomap[b]
}

// Retrieve the username and password given in the HTTP Authorization header
func HTTPBasicAuthUsernamePassword(r *http.Request) (string, string, error) {
	auth := r.Header.Get("Authorization")
	if len(auth) < 6 || auth[:6] != "Basic " {
		return "", "", errors.New("HTTP Basic Auth: Invalid header: " + auth)
	}
	b, err := base64.StdEncoding.DecodeString(auth[6:])
	if err != nil {
		return "", "", err
	}
	tokens := strings.SplitN(string(b), ":", 2)
	if len(tokens) != 2 {
		return "", "", errors.New("HTTP Basic Auth: Invalid number of tokens: " + strconv.Itoa(len(tokens)))
	}
	// Return the given username and password
	return tokens[0], tokens[1], nil
}

// Return the HTTP Basic Auth username or an empty string.
func HTTPBasicAuthUsername(r *http.Request) string {
	// Return the username if there are no errors
	if username, _, err := HTTPBasicAuthUsernamePassword(r); err == nil {
		return username
	}
	// For all other cases, return an empty string
	return ""
}

// Ask the user for a username and password, by using HTTP Basic Auth, setting a header and
// rejecting the current HTTP request. "Authorization Required" will be used as the realm.
func HTTPBasicAuthRejectPrompt(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Authorization Required"`)
	http.Error(w, "Not Authorized", http.StatusUnauthorized)
}

// Ask the user for a username and password, by using HTTP Basic Auth, setting a header and
// rejecting the current HTTP request. The given realm will be used as the realm. The
// realm is often shown in the dialog box that asks for a username and password, and can be
// used to identify a website or a collection of websites.
func HTTPBasicAuthRejectPromptWithRealm(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	http.Error(w, "Not Authorized", http.StatusUnauthorized)
}

// SecureCompare performs a constant time compare of two strings to limit timing attacks.
// From https://github.com/martini-contrib/auth/blob/master/util.go
func SecureCompare(given string, actual string) bool {
	givenSha := sha256.Sum256([]byte(given))
	actualSha := sha256.Sum256([]byte(actual))
	return subtle.ConstantTimeCompare(givenSha[:], actualSha[:]) == 1
}

// For some URL path prefixes, check if the user has the right username on password with HTTP Basic Auth (when already registered on the server).
// This is completely unsecure over unencrypted HTTP, but is secure over HTTPS.
// This function can be used as Martini middleware. If onlyAdmin is enabled, only the administrator user will be allowed for the given prefixes.
func MartiniBasicAuthWithPathPrefixes(userstate *permissions.UserState, pathPrefixes []string, onlyAdmin bool) martini.Handler {
	return func(w http.ResponseWriter, r *http.Request, c martini.Context) {
		for _, pathPrefix := range pathPrefixes {
			if strings.HasPrefix(r.URL.Path, pathPrefix) {
				// Protected by HTTP Basic Auth
				username, password, err := HTTPBasicAuthUsernamePassword(r)
				if err != nil {
					// There was an error retrieving the username and password, reject and return
					HTTPBasicAuthRejectPrompt(w)
					return
				}
				// Check if the username is empty
				if username == "" {
					// Empty username
					HTTPBasicAuthRejectPrompt(w)
					return
				}
				// Check if the user is the administrator user, if onlyAdmin is toggled on
				if onlyAdmin && !SecureCompare(AdminUsername, username) {
					// Reject and return
					HTTPBasicAuthRejectPrompt(w)
					return
				}
				// Check if the username exists
				if !userstate.HasUser(username) {
					// Reject and return
					HTTPBasicAuthRejectPrompt(w)
					return
				}
				// Check if the password is correct
				if !userstate.CorrectPassword(username, password) {
					// Reject and return
					HTTPBasicAuthRejectPrompt(w)
					return
				}
				// Ok
			}
		}
		// Go on
		c.Next()
	}
}

func main() {
	fmt.Println(Title)

	// New Martini
	m := martini.Classic()

	// New Renderer
	r := render.Renderer(render.Options{})
	m.Use(r)

	// Initiate the user system
	perm := permissions.NewWithRedisConf(7, "")
	userstate := perm.UserState()

	// Protect the API url prefix with HTTP Basic Auth.
	// Only the admin user is allowed to access the API.
	m.Use(MartiniBasicAuthWithPathPrefixes(userstate, []string{API}, true))

	// --- Public pages and admin panel ---

	// Public page
	m.Get("/", func(r render.Render) {
		msg := "Everything is fine."
		if !userstate.HasUser(AdminUsername) {
			msg = "No registered administrator. Please visit /register."
		}
		data := map[string]string{
			"title": Title,
			"msg":   msg,
		}

		// Uses templates/index.tmpl
		r.HTML(http.StatusOK, "index", data)
	})

	// AJAX / server state test
	setupTrigger(m, r)

	// Admin status
	m.Any("/status", func(req *http.Request, r render.Render) {
		data := map[string]string{
			"title":          Title,
			"admin":          b2yn(userstate.HasUser(AdminUsername)),
			"serverloggedin": b2yn(userstate.IsLoggedIn(AdminUsername)),
			"basicusername":  HTTPBasicAuthUsername(req),
		}
		// Uses templates/status.tmpl
		r.HTML(http.StatusOK, "status", data)
	})

	// The admin panel
	m.Get("/admin", func(w http.ResponseWriter, req *http.Request, r render.Render) {
		// TODO: Write an admin panel for managing users

		data := map[string]string{
			"title": Title,
			"msg":   "Admin panel, work in progress",
		}

		// Uses templates/index.tmpl
		r.HTML(http.StatusOK, "index", data)
	})

	// Enable temporarily for removing and re-creating the admin user with a new password
	//m.Get("/remove", func() string {
	//	userstate.RemoveUser(AdminUsername)
	//	return "removed admin user"
	//})

	// --- Admin user management ---

	// Register the admin password
	m.Get("/register", func(w http.ResponseWriter, req *http.Request) {
		if userstate.HasUser(AdminUsername) {
			fmt.Fprint(w, "Error: Already has a registered administrator.")
			return
		}
		w.Header().Add("Content-Type", "text/html")
		fmt.Fprint(w, "<!doctype html><html><body>")
		fmt.Fprint(w, instapage.RegisterForm())
		fmt.Fprint(w, "</body></html>")
	})
	m.Post("/register", binding.Bind(RegisterAdmin{}), func(ra RegisterAdmin) string {
		username := AdminUsername
		if !userstate.HasUser(username) {
			userstate.AddUser(username, ra.Password1, ra.Email)
			userstate.SetAdminStatus(username)
		}
		return "Success: Registered administrator: " + username + "."
	})

	// Login admin
	m.Get("/login", func(w http.ResponseWriter, req *http.Request) {
		if userstate.AdminRights(req) || userstate.UserRights(req) {
			fmt.Fprint(w, "Error: Already logged in as a user or as an administrator.")
			return
		}
		w.Header().Add("Content-Type", "text/html")
		fmt.Fprint(w, "<!doctype html><html><body>")
		fmt.Fprint(w, instapage.LoginForm())
		fmt.Fprint(w, "</body></html>")
	})
	m.Post("/login", binding.Bind(LoginAdmin{}), func(la LoginAdmin, w http.ResponseWriter, req *http.Request) string {
		username := AdminUsername
		if !userstate.HasUser(username) {
			return "Error: User " + username + " does not exist."
		}
		if !userstate.CorrectPassword(username, la.Password) {
			return "Error: Incorrect password."
		}
		userstate.Login(w, username)
		if !userstate.AdminRights(req) {
			return "Error: User " + username + " was logged in, but does not have admin rights. Cookie problem?"
		}
		return "Success: Logged in " + username + "."
	})

	// Logout admin
	m.Any("/logout", func(req *http.Request) string {
		username := AdminUsername
		if !userstate.AdminRights(req) {
			return "Error: Need administrator rights to log out the administrator user."
		}
		userstate.Logout(username)
		if userstate.IsLoggedIn(username) {
			// logout failed
			return "Error: Could not log out " + username + "."
		}
		return "Success: Logged out " + username + "."
	})

	// --- REST methods ---

	// For testing the API
	m.Any(API, func(r render.Render) {
		r.JSON(http.StatusOK, map[string]interface{}{"all systems": "go"})
	})

	// For adding users

	m.Post(API+"create/:username", func(params martini.Params, r render.Render) {
		username := params["username"]
		if userstate.HasUser(username) {
			r.JSON(http.StatusConflict, map[string]interface{}{"error": "user " + username + " already exists"})
			return
		}
		userstate.AddUser(username, "", "")
		if userstate.HasUser(username) {
			r.JSON(http.StatusOK, map[string]interface{}{"create": true})
		} else {
			r.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "user " + username + " was not created"})
		}
	})

	m.Post(API+"register/:username/:password/:email", func(params martini.Params, r render.Render) {
		username := params["username"]
		password := params["password"]
		email := params["email"]
		if userstate.HasUser(username) {
			r.JSON(http.StatusConflict, map[string]interface{}{"error": "user " + username + " already exists"})
			return
		}
		userstate.AddUser(username, password, email)
		if userstate.HasUser(username) {
			r.JSON(http.StatusOK, map[string]interface{}{"create": true})
		} else {
			r.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "user " + username + " was not created"})
		}
	})

	// For logging in
	m.Post(API+"login/:username/:password", func(w http.ResponseWriter, params martini.Params, r render.Render) {
		username := params["username"]
		password := params["password"]
		if userstate.CorrectPassword(username, password) {
			userstate.SetLoggedIn(username)
		}
		if !userstate.IsLoggedIn(username) {
			r.JSON(http.StatusUnauthorized, map[string]interface{}{"error": "could not log in " + username})
			return
		}
		r.JSON(http.StatusOK, map[string]interface{}{"login": true})
	})

	// For logging out
	m.Any(API+"logout/:username", func(params martini.Params, r render.Render) {
		username := params["username"]
		userstate.Logout(username)
		if userstate.IsLoggedIn(username) {
			r.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "user " + username + " is still logged in!"})
			return
		}
		r.JSON(http.StatusOK, map[string]interface{}{"logout": true})
	})

	// For login status
	m.Any(API+"status/:username", func(params martini.Params, r render.Render) {
		username := params["username"]
		if userstate.IsLoggedIn(username) {
			r.JSON(http.StatusOK, map[string]interface{}{"login": true})
		} else {
			r.JSON(http.StatusOK, map[string]interface{}{"login": "false"})
		}
	})

	// Score POST og GET + timestamp
	m.Post(API+"score/:username/:score", func(params martini.Params, r render.Render) {
		username := params["username"]
		score := params["score"]

		if !userstate.HasUser(username) {
			r.JSON(http.StatusNotFound, map[string]interface{}{"error": "no such user " + username})
			return
		}

		users := userstate.Users()
		users.Set(username, "score", score)

		r.JSON(http.StatusOK, map[string]interface{}{"score set": true})
	})
	m.Get(API+"score/:username", func(params martini.Params, r render.Render) {
		username := params["username"]

		if !userstate.HasUser(username) {
			r.JSON(http.StatusNotFound, map[string]interface{}{"error": "no such user " + username})
			return
		}

		users := userstate.Users()
		score, err := users.Get(username, "score")
		if err != nil {
			r.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "could not get score for " + username})
			return
		}

		r.JSON(http.StatusOK, map[string]interface{}{"score": score})
	})

	// Share the files in static
	m.Use(martini.Static("static"))

	// --- Social network function ---

	// Facebook friends
	setupFB(m, r, userstate)

	// Instagram friends
	setupInsta(m, r, userstate)

	// port 3000 by default, uses PORT and HOST environment variables
	m.Run()
}
