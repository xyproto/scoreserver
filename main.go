package main

import (
	"fmt"
	"net/http"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/xyproto/auth"
	"github.com/xyproto/fizz"
	"github.com/xyproto/instapage"
)

const (
	Version       = "1.0"
	API           = "/api/" + Version + "/"
	Title         = "Score Server " + Version
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

func main() {
	fmt.Println(Title)

	// New Martini
	m := martini.Classic()

	// New Renderer
	r := render.Renderer(render.Options{})
	m.Use(r)

	// Initiate the user and permission system
	fizz := fizz.NewWithRedisConf(7, "")
	userstate := fizz.UserState()

	// Permission system
	//perm := fizz.Perm()

	// Authorization is proviced by http basic auth.
	//perm.Clear()

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

	// Admin status
	// TODO Consider adding a bool2yesno function and use fmt.Sprintf instead
	m.Any("/status", func(r render.Render) {
		s := "has administrator: "
		if userstate.HasUser(AdminUsername) {
			s += "yes"
		} else {
			s += "no"
		}
		s += ", "
		s += "logged in: "
		if userstate.IsLoggedIn(AdminUsername) {
			s += "yes"
		} else {
			s += "no"
		}

		data := map[string]string{
			"title": Title,
			"msg":   s,
		}

		// Uses templates/index.tmpl
		r.HTML(http.StatusOK, "index", data)
	})

	// The admin panel
	m.Get("/admin", func(w http.ResponseWriter, req *http.Request, r render.Render) {
		// TODO: Write a nice admin panel, for managing users, like in ftls2

		data := map[string]string{
			"title": Title,
			"msg":   "Admin panel, work in progress",
		}

		// Uses templates/index.tmpl
		r.HTML(http.StatusOK, "index", data)
	})

	// Enable temporarily for removing and re-creating the admin user with a new pasword
	//m.Get("/remove", func() string {
	//	userstate.RemoveUser(AdminUsername)
	//	return "removed admin user"
	//})

	// --- Admin user management ---

	// Register the admin password
	m.Get("/register", func(w http.ResponseWriter, req *http.Request) {
		// TODO: Handle things differently if a regular user is logged in
		if !userstate.HasUser(AdminUsername) {
			w.Header().Add("Content-Type", "text/html")
			fmt.Fprint(w, "<!doctype html><html><body>")
			fmt.Fprint(w, instapage.RegisterForm())
			fmt.Fprint(w, "</body></html>")
			return
		}
	})
	m.Post("/register", binding.Bind(RegisterAdmin{}), func(ra RegisterAdmin) string {
		username := AdminUsername
		if !userstate.HasUser(username) {
			userstate.AddUser(username, ra.Password1, ra.Email)
			userstate.SetAdminStatus(username)
		}
		return "registered " + username
	})

	// Login admin
	m.Get("/login", func(w http.ResponseWriter, req *http.Request) {
		// TODO: Handle things differently if a regular user is logged in
		if !userstate.AdminRights(req) {
			w.Header().Add("Content-Type", "text/html")
			fmt.Fprint(w, "<!doctype html><html><body>")
			fmt.Fprint(w, instapage.LoginForm())
			fmt.Fprint(w, "</body></html>")
			return
		}
	})
	m.Post("/login", binding.Bind(LoginAdmin{}), func(la LoginAdmin) string {
		username := AdminUsername
		if !userstate.HasUser(username) {
			return "FAIL must have " + username
		}
		if !userstate.CorrectPassword(username, la.Password) {
			return "FAIL password"
		}
		userstate.SetLoggedIn(username)
		return "logged in " + username
	})

	// Logout admin
	m.Any("/logout", func() string {
		username := AdminUsername
		userstate.Logout(username)
		// TODO errorcheck
		return "logged out " + username
	})

	// --- REST methods ---

	// The API uses HTTP Basic Auth instead of cookies
	//perm.AddPublicPath(API)

	// For testing the API
	m.Any(API, func(r render.Render) {
		r.JSON(http.StatusOK, map[string]interface{}{"hello": "fjaselus"})
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
		password := params["email"]
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

		users := userstate.GetUsers()
		users.Set(username, "score", score)

		r.JSON(http.StatusOK, map[string]interface{}{"score set": true})
	})
	m.Get(API+"score/:username", func(params martini.Params, r render.Render) {
		username := params["username"]

		if !userstate.HasUser(username) {
			r.JSON(http.StatusNotFound, map[string]interface{}{"error": "no such user " + username})
			return
		}

		users := userstate.GetUsers()
		score, err := users.Get(username, "score")
		if err != nil {
			r.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "could not get score for " + username})
			return
		}

		r.JSON(http.StatusOK, map[string]interface{}{"score": score})
	})

	// Activate the permission middleware
	//m.Use(fizz.All())

	// Share the files in static
	m.Use(martini.Static("static"))

	// Only enable HTTP Basic Auth for paths that starts with "/api"
	m.Use(auth.BasicFunc(func(username, password string) bool {
		// Check if the admin user has the correct password, as registered for the admin user
		return auth.SecureCompare(AdminUsername, username) && userstate.CorrectPassword(AdminUsername, password)
	}, "/api"))

	m.Run() // port 3000 by default
}
