package main

import (
	"errors"
	"fmt"
	"github.com/go-martini/martini"
	fb "github.com/huandu/facebook"
	"github.com/martini-contrib/render"
	"net/http"
)

const (
	appName        = "PonyBumpCommander"
	appAccessToken = "1516006388643143|iQlhYN9N80RQnAOSXWNndB_oVos"
)

func setupFB(m *martini.ClassicMartini, r martini.Handler) {

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
