package main

import (
	"fmt"
	"net/http"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/binding"
	"github.com/xyproto/fizz"
	"github.com/xyproto/instapage"

)

const (
	Version       = "1.0"
	API           = "/api/" + Version + "/"
	Title         = "Highscore Server " + Version
	Auth_Username = "admin"
	Auth_Password = "testfest"
)

type UsernamePassword struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
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
	perm := fizz.Perm()

	// --- Admin panel ---

	// This admin panel performs its own checks
	perm.SetAdminPath([]string{})
	m.Get("/admin", func(w http.ResponseWriter, req *http.Request, r render.Render) {
		// TODO:
		// If no admin user, ask for username and password
		// If an admin user, ask for login
		// If logged in as admin user, list users (admin panel from ftls2)
		if !userstate.HasUser("admin") {
			w.Header().Add("Content-Type", "text/html")
			fmt.Fprint(w, instapage.RegisterForm())
			return
		}
		if !userstate.AdminRights(req) {
			w.Header().Add("Content-Type", "text/html")
			fmt.Fprint(w, instapage.LoginForm())
			return
		}
		r.HTML(http.StatusOK, "admin", userstate)
	})

	// Register the admin username and password
	m.Post("/login", binding.Bind(UsernamePassword{}), func (up UsernamePassword) string {
		return "logging in " + up.Username + ":" + up.Password
	})

	// --- REST methods ---

	// Public page
	m.Get("/", func() string {
		return Title
	})

	// The API uses HTTP Basic Auth instead of cookies
	perm.AddPublicPath(API)

	// For testing the API
	m.Any(API, func(r render.Render) {
		r.JSON(200, map[string]interface{}{"hello": "fjaselus"})
	})

	// For adding users
	m.Post(API + "create/:username/:password", func(params martini.Params, r render.Render) {
		username := params["username"]
		password := params["password"]
		if userstate.HasUser(username) {
			r.JSON(200, map[string]interface{}{"error": "user " + username + " already exists"})
			return
		}
		userstate.AddUser(username, password, "")
		if userstate.HasUser(username) {
			r.JSON(200, map[string]interface{}{"create": true})
		} else {
			r.JSON(200, map[string]interface{}{"error": "user " + username + " was not created"})
		}
	})

	// For logging in
	m.Post(API + "login/:username/:password", func(w http.ResponseWriter, params martini.Params, r render.Render) {
		username := params["username"]
		password := params["password"]
		if userstate.CorrectPassword(username, password) {
			userstate.SetLoggedIn(username)
		}
		if !userstate.IsLoggedIn(username) {
			r.JSON(200, map[string]interface{}{"error": "could not log in " + username})
			return
		}
		r.JSON(200, map[string]interface{}{"login": true})
	})

	// For logging out
	m.Any(API + "logout/:username", func(params martini.Params, r render.Render) {
		username := params["username"]
		userstate.Logout(username)
		if userstate.IsLoggedIn(username) {
			r.JSON(200, map[string]interface{}{"error": "user " + username + " is still logged in!"})
			return
		}
		r.JSON(200, map[string]interface{}{"logout": true})
	})

	// For login status
	m.Post(API + "status/:username", func(params martini.Params, r render.Render) {
		username := params["username"]
		if userstate.IsLoggedIn(username) {
			r.JSON(200, map[string]interface{}{"login": true})
		} else {
			r.JSON(200, map[string]interface{}{"login": "false"})
		}
	})

	// Score POST og GET + timestamp
	m.Post(API + "score/:username/:score", func(params martini.Params, r render.Render) {
		username := params["username"]
		score := params["score"]

		if !userstate.HasUser(username) {
			r.JSON(200, map[string]interface{}{"error": "no such user " + username})
			return
		}

		users := userstate.GetUsers()
		users.Set(username, "score", score)

		r.JSON(200, map[string]interface{}{"score set": true})
	})
	m.Get(API + "score/:username", func(params martini.Params, r render.Render) {
		username := params["username"]

		if !userstate.HasUser(username) {
			r.JSON(200, map[string]interface{}{"error": "no such user " + username})
			return
		}

		users := userstate.GetUsers()
		score, err := users.Get(username, "score")
		if err != nil {
			r.JSON(200, map[string]interface{}{"error": "could not get score for " + username})
			return
		}

		r.JSON(200, map[string]interface{}{"score": score})
	})

	// Activate the permission middleware
	m.Use(fizz.All())

	// Share the files in static
	m.Use(martini.Static("static"))

	// HTTP Basic Auth
	m.Use(auth.BasicFunc(func(username, password string) bool {
		// Using "admin" and the password set in the admin panel
		return auth.SecureCompare(username, "admin") && auth.SecureCompare(password, "SECRET_PASSWORD")
	}))

	m.Run() // port 3000 by default
}
