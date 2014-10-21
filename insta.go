package main

// Work in progress!

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"net/http"
)

const (
	instaAccessToken = "-"
)

func setupInsta(m *martini.ClassicMartini, r martini.Handler) {

	// Number of friends on Instagram
	m.Any(API+"insta/friends/:userAccessToken", func(params martini.Params, r render.Render) {
		userAccessToken := params["userAccessToken"]
		friends, err := instaFriends(userAccessToken)
		if err != nil {
			r.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "could not fetch friends from instagram: " + err.Error()})
			return
		} else {
			r.JSON(http.StatusOK, map[string]interface{}{"friends": friends})
		}
	})
}

func instaFriends(userAccessToken string) (string, error) {
	// todo, use instaAccessToken, fetch friends from instagram
	return "123", nil
}
