package main

import (
	"errors"
	"fmt"
	"github.com/go-martini/martini"
	fb "github.com/huandu/facebook"
	"github.com/martini-contrib/render"
	"github.com/xyproto/permissions2"
	"net/http"
)

const (
	// Your values goes here
	appName        = "fb_app_name"
	appAccessToken = "fb_app_access_token"
	fbTokenName    = "fb_user_access_token"
	fbIDName       = "fb_user_id"
)

func setupFB(m *martini.ClassicMartini, r martini.Handler, userstate *permissions.UserState) {

	// Store access token for a given user
	//m.Post(API+"fb/reg/:username/:token", func(params martini.Params, r render.Render) {
	m.Any(API+"fb/reg/:username/:token", func(params martini.Params, r render.Render) {
		username := params["username"]
		token := params["token"]

		if !userstate.HasUser(username) {
			r.JSON(http.StatusNotFound, map[string]interface{}{"error": "no such user " + username})
			return
		}

		users := userstate.Users()
		users.Set(username, fbTokenName, token)

		r.JSON(http.StatusOK, map[string]interface{}{"user id and access token set": true})
	})

	// Number of friends on Instagram
	m.Get(API+"fb/friends/:username", func(params martini.Params, r render.Render) {
		username := params["username"]

		if !userstate.HasUser(username) {
			r.JSON(http.StatusNotFound, map[string]interface{}{"error": "no such user " + username})
			return
		}

		users := userstate.Users()

		userAccessToken, err := users.Get(username, fbTokenName)
		if err != nil {
			r.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "could not get fb user access token for " + username})
			return
		}

		friends, err := facebookFriends(userAccessToken)
		if err != nil {
			r.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "could not fetch friends from facebook: " + err.Error()})
			return
		} else {
			r.JSON(http.StatusOK, map[string]interface{}{"friends": friends})
		}
	})

}

func facebookFriends(userAccessToken string) (string, error) {
	// create a global App var to hold your app id and secret.
	var globalApp = fb.New(appName, appAccessToken)

	session := globalApp.Session(userAccessToken)

	// validate access token. err is nil if token is valid.
	err := session.Validate()
	if err != nil {
		return "", err
	}

	// use session to send api request with your access token.
	res, _ := session.Get("/me/friends", nil)

	friends := res.Get("summary")

	friendCountMap := friends.(map[string]interface{})

	friendCount, ok := friendCountMap["total_count"]
	if !ok {
		return "", errors.New("could not find total_count in result from fb")
	}

	return fmt.Sprintf("%v", friendCount), nil
}
