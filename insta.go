package main

// Work in progress!

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"net/http"
	"strings"
)

func setupInsta(m *martini.ClassicMartini, r martini.Handler) {

	// Number of friends on Instagram
	m.Any(API+"insta/friends/:userID/:accessToken", func(params martini.Params, r render.Render) {
		userID := params["userID"]
		accessToken := params["accessToken"]
		friends, err := instaFriends(userID, accessToken)
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
