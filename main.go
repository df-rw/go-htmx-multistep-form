package main

import (
	"html/template"
	"log"
	"net/http"
)

type Application struct {
	tpl *template.Template
}

func (app *Application) render(w http.ResponseWriter, pageName string, pageData map[string]any, statusCode int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	app.tpl.ExecuteTemplate(w, pageName, pageData)
}

func (app *Application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, "home", nil, http.StatusOK)
}

func (app *Application) form(w http.ResponseWriter, r *http.Request) {
	section := r.PathValue("section")

	app.render(w, "page-form-"+section, nil, http.StatusOK)
}

func (app *Application) formSection(w http.ResponseWriter, r *http.Request) {
	section := r.PathValue("section")

	_ = r.ParseForm()

	hxRequest := r.Header.Get("Hx-Request") == "true"
	hxBoosted := r.Header.Get("Hx-Boosted") == "true"

	switch section {
	case "one":
		next := r.FormValue("next") == "next"
		cancel := r.FormValue("cancel") == "cancel"

		switch {
		case next:
			if hxRequest && hxBoosted {
				app.render(w, "form-two", nil, http.StatusOK)
			} else {
				http.Redirect(w, r, "/form/two", http.StatusSeeOther)
			}

		case cancel:
			if hxRequest && hxBoosted {
				w.Header().Set("Hx-Redirect", "/")
			} else {
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
		default:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
		return

	case "two":
		next := r.FormValue("next") == "next"
		prev := r.FormValue("prev") == "prev"

		switch {
		case next:
			if hxRequest && hxBoosted {
				app.render(w, "form-three", nil, http.StatusOK)
			} else {
				http.Redirect(w, r, "/form/three", http.StatusSeeOther)
			}

		case prev:
			if hxRequest && hxBoosted {
				app.render(w, "form-one", nil, http.StatusOK)
			} else {
				http.Redirect(w, r, "/form/one", http.StatusSeeOther)
			}

		default:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
		return

	case "three":
		submit := r.FormValue("submit") == "submit"
		prev := r.FormValue("prev") == "prev"

		switch {
		case submit:
			if hxRequest && hxBoosted {
				w.Header().Set("Hx-Redirect", "/form/submitted")
			} else {
				http.Redirect(w, r, "/form/submitted", http.StatusSeeOther)
			}

		case prev:
			if hxRequest && hxBoosted {
				app.render(w, "form-two", nil, http.StatusOK)
			} else {
				http.Redirect(w, r, "/form/two", http.StatusSeeOther)
			}

		default:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
	}
}

func (app *Application) formSubmitted(w http.ResponseWriter, r *http.Request) {
	app.render(w, "form-submitted", nil, http.StatusOK)
}

func main() {
	app := &Application{
		tpl: template.Must(template.ParseGlob("templates/*.tmpl")),
	}

	http.HandleFunc("GET /", app.home)
	http.HandleFunc("GET /form/{section}", app.form)
	http.HandleFunc("POST /form/{section}", app.formSection)

	log.Println("listening on port 4005")
	log.Fatal(http.ListenAndServe(":4005", nil))
}
