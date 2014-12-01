package main

// Work in progress!

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/xyproto/permissions2"
	"net/http"
	"strings"
)

const (
	instaTokenName = "insta_user_access_token"
	instaIDName    = "insta_user_id"
)

func setupInsta(m *martini.ClassicMartini, r martini.Handler, userstate *permissions.UserState) {

	// Store access token for a given user (token includes userID)
	m.Any(API+"insta/reg/:username/:token", func(params martini.Params, r render.Render) {
		//m.Post(API+"insta/reg/:username/:token", func(params martini.Params, r render.Render) {
		username := params["username"]
		token := params["token"]

		if !userstate.HasUser(username) {
			r.JSON(http.StatusNotFound, map[string]interface{}{"error": "no such user " + username})
			return
		}

		users := userstate.Users()
		users.Set(username, instaTokenName, token)

		r.JSON(http.StatusOK, map[string]interface{}{"user access token set": true})
	})

	// Number of friends on Instagram
	m.Get(API+"insta/friends/:username", func(params martini.Params, r render.Render) {
		username := params["username"]

		if !userstate.HasUser(username) {
			r.JSON(http.StatusNotFound, map[string]interface{}{"error": "no such user " + username})
			return
		}

		users := userstate.Users()

		userAccessToken, err := users.Get(username, instaTokenName)
		if err != nil {
			r.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "could not get insta user access token for " + username})
			return
		}

		if !strings.Contains(userAccessToken, ".") {
			r.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "user access token must contain a . for separating user ID and access token"})
			return
		}
		fields := strings.Split(userAccessToken, ".")
		userID := fields[0]

		friends, err := instaFriends(userID, userAccessToken)
		if err != nil {
			r.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "could not fetch friends from instagram: " + err.Error()})
			return
		} else {
			r.JSON(http.StatusOK, map[string]interface{}{"friends": friends})
		}
	})

}

func instaFriends(userID, accessToken string) (string, error) {
	if !strings.HasPrefix(accessToken, userID) {
		// accessToken must start with userId
		return "", errors.New("accessToken must start with userId (internal safety)")
	}

	// Ok, fetch friends from instagraom
	url := fmt.Sprintf("https://api.instagram.com/v1/users/%s/?access_token=%s", userID, accessToken)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fmt.Printf("%#v\n", resp)

	dec := json.NewDecoder(resp.Body)
	if dec == nil {
		return "", errors.New("Could not decode JSON data from instagram")
	}

	json_map := make(map[string]interface{})
	err = dec.Decode(&json_map)
	if err != nil {
		return "", err
	}

	// Get the number of people following the given userID
	data := json_map["data"].(map[string]interface{})
	counts := data["counts"].(map[string]interface{})
	friends := counts["followed_by"]

	// Return the number of friends, as a string
	return fmt.Sprintf("%v", friends), nil
}
