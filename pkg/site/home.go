package site

import (
	"html/template"
	"net/http"
)

type HomePageRenderer struct {
	tmpl *template.Template
}

func NewHomePageRenderer() *HomePageRenderer {
	return &HomePageRenderer{
		tmpl: template.Must(template.ParseFiles("home.html")),
	}
}

func (h *HomePageRenderer) HomePage(w http.ResponseWriter, r *http.Request) {
	h.tmpl.Execute(w, nil)
}
