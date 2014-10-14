package main

import (
	fb "github.com/huandu/facebook"
)

const (
	appName        = "PonyBumpCommander"
	appAccessToken = "1516006388643143|iQlhYN9N80RQnAOSXWNndB_oVos"
)

func facebookFriends(userAccessToken string) (string, error) {
	// create a global App var to hold your app id and secret.
	var globalApp = fb.New(appName, appAccessToken)

	session := globalApp.Session(userAccessToken)

	// validate access token. err is nil if token is valid.
	err := session.Validate()
	if err != nil {
		return err, -1
	}

	// use session to send api request with your access token.
	res, _ := session.Get("/me/friends", nil)

	friends := res.Get("summary")

	return friends
}
