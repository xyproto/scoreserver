package main

import (
	"errors"
	"fmt"
	"github.com/go-martini/martini"
	fb "github.com/huandu/facebook"
	"github.com/martini-contrib/render"
	"github.com/xyproto/permissions"
	"net/http"
)

const (
	appName        = "PonyBumpCommander"
	appAccessToken = "1516006388643143|iQlhYN9N80RQnAOSXWNndB_oVos"
)

func setupFB(m *martini.ClassicMartini, r martini.Handler, userstate *permissions.UserState) {

	// Store access token for a given user
	m.Post(API+"fb/reg/:username/:token", func(params martini.Params, r render.Render) {
		username := params["username"]
		token := params["token"]

		if !userstate.HasUser(username) {
			r.JSON(http.StatusNotFound, map[string]interface{}{"error": "no such user " + username})
			return
		}

		users := userstate.GetUsers()
		users.Set(username, "fb_user_access_token", token)

		r.JSON(http.StatusOK, map[string]interface{}{"user access token set": true})
	})

	// Number of friends on Facebook
	m.Any(API+"fb/friends/:userAccessToken", func(params martini.Params, r render.Render) {
		userAccessToken := params["userAccessToken"]
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