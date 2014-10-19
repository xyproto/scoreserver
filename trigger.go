package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"net/http"
)

func setupTrigger(m *martini.ClassicMartini, r martini.Handler) {
	var trigger = false

	m.Get("/trigger/get", func(r render.Render) {
		r.JSON(http.StatusOK, map[string]interface{}{"trigger": trigger})
	})

	m.Get("/trigger/set", func(r render.Render) {
		trigger = true
		r.JSON(http.StatusOK, map[string]interface{}{"trigger": trigger})
	})

	m.Get("/trigger/clear", func(r render.Render) {
		trigger = false
		r.JSON(http.StatusOK, map[string]interface{}{"trigger": trigger})
	})

	m.Get("/trigger", func(r render.Render) {
		data := map[string]string{
			"title": "trigger",
			"msg":   b2yn(trigger),
		}

		// Uses templates/index.tmpl
		r.HTML(http.StatusOK, "index", data)
	})

}
