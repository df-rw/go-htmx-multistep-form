# Poor man's multi-step forms with Go

A multi-step (also "wizard") form breaks up a single form into multiple pages
for easier consumption by a user. Here we describe a couple of approaches of
implementing multi-step forms in a Go web application.

## tl;dr

```
go run ./...
```

## The boilerplate

Some boilerplate to make the application easier to digest.

Our application loads templates from `./templates/` into an application
context, and has a single route for showing a home page.

### `main.go`

```go
// main.go
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

func main() {
	app := &Application{
		tpl: template.Must(template.ParseGlob("templates/*.tmpl")),
	}

	http.HandleFunc("GET /", app.home)

	log.Println("listening on port 4005")
	log.Fatal(http.ListenAndServe(":4005", nil))
}
```

### `./templates/home.tmpl`

```gohtml
{{/* templates/home.tmpl */}}
{{block "top" .}}
    <!doctype html>
    <html>
        <head>
        </head>

        <body>
            <ul>
                <li><a href="/">home</a></li>
            </ul>
{{end}}

{{block "bottom" .}}
        </body>
    </html>
{{end}}

{{block "home" .}}
    {{template "top" .}}

    <p>this is the home page</p>

    {{template "bottom" .}}
{{end}}
```

### `go.mod`

Setup `go.mod` and run the application:

```shell
go mod init ms
go run ./...
```

## Setup the multi-step form

- Our form will consist of four pages:
  - Three pages, one for each step in the multi-step form;
  - One page for the submission message.
- Each page of the form is a complete HTML page.
- Each page POSTs to a processing endpoint, named for itself.

### Form templates

- All pages are represented in a single template file.

```gohtml
{{/* ./templates/form.tmpl */}}
{{block "form-one" .}}
    {{template "top" .}}
    <form method="POST" action="/form/one">
        <input type="text" name="name" placeholder="type your name" />
        <button name="next" value="next">next</button>
        <button name="cancel" value="cancel">cancel</button>
    </form>
    {{template "bottom" .}}
{{end}}

{{block "form-two" .}}
    {{template "top" .}}
    <form method="POST" action="/form/two">
        <input type="text" name="email" placeholder="email address" />
        <button name="next" value="next">next</button>
        <button name="prev" value="prev">prev</button>
    </form>
    {{template "bottom" .}}
{{end}}

{{block "form-three" .}}
    {{template "top" .}}
    <form method="POST" action="/form/three">
        <input type="text" name="phone" placeholder="phone number" />
        <button name="submit" value="submit">submit</button>
        <button name="prev" value="prev">prev</button>
    </form>
    {{template "bottom" .}}
{{end}}

{{block "form-submitted" .}}
    {{template "top" .}}
    <p>form submitted</p>
    {{template "bottom" .}}
{{end}}
```

### Routing for the form steps

- We will have a single endpoint for GET requests of each step.
- We will have a single endpoint for POST processing of each step.


```go
    // main.go:main()
	http.HandleFunc("GET /form/{section}", app.form)
	http.HandleFunc("POST /form/{section}", app.formSection)
```

### Processing the form steps

The GET handler:

```go
// main.go
func (app *Application) form(w http.ResponseWriter, r *http.Request) {
	section := r.PathValue("section")

	app.render(w, "form-"+section, nil, http.StatusOK)
}
```

The general form of the POST handler is:

```go
    ...
    // Figure out which step we are on from the URL.
    section := r.PathValue("section")

    // Each section handles it's own validation and subsequent routing.
    switch section {
    case "first-step-in-the-multi-step-form":
        // Figure out what navigation the client performed.
        next := r.FormValue("next") == "next"
        prev := r.FormValue("prev") == "prev"
        cancel := r.FormValue("cancel") == "cancel"
        ...

        // Do valiation.
        if r.FormValue("emailaddress") == "" {
            errs["EmailAddress"] = "You didn't specify an email address."
        }
        ...

        // Tell the client where to go next, based on their input,
        // results of the validation etc.
        switch {
        case next:
            http.Redirect(w, r, "second-step-in-the-multi-step-form", http.StatusSeeOther)
        case cancel:
            http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
        }
        return

    // other cases
    }
    ...
```

Our demonstration code to add to `main.go` looks like:

```go
// main.go
func (app *Application) formSection(w http.ResponseWriter, r *http.Request) {
	section := r.PathValue("section")

	_ = r.ParseForm()

	switch section {
	case "one":
		next := r.FormValue("next") == "next"
		cancel := r.FormValue("cancel") == "cancel"

		switch {
		case next:
			http.Redirect(w, r, "/form/two", http.StatusSeeOther)
		case cancel:
			http.Redirect(w, r, "/", http.StatusSeeOther)
		default:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
		return

	case "two":
		next := r.FormValue("next") == "next"
		prev := r.FormValue("prev") == "prev"

		switch {
		case next:
			http.Redirect(w, r, "/form/three", http.StatusSeeOther)
		case prev:
			http.Redirect(w, r, "/form/one", http.StatusSeeOther)
		default:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
		return

	case "three":
		submit := r.FormValue("submit") == "submit"
		prev := r.FormValue("prev") == "prev"

		switch {
		case submit:
			http.Redirect(w, r, "/form/submitted", http.StatusSeeOther)
		case prev:
			http.Redirect(w, r, "/form/two", http.StatusSeeOther)
		default:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
	}
}
```

Add a link to access the form for easy navigation from the `top` block:

```gohtml
    {{/* ./templates/home.tmpl */}}
    <li><a href="/form/one">form</a></li>
```

Rerunning the application at this point will let you step forward and backward through the form.

## Upgrades

### Removing full page reloads

Our application at the moment makes a full round-trip to the server when each
step in the form is completed. The entire page contents are re-rendered. It
would be nice if, on each step, we retrieved the next step, and replaced the
current step with that new content without doing a full page reload. We will
need to leverage JavaScript to be able to do this.

### Graceful degradation, or progressive enhancement

In addition to avoiding full page reloads, it would be nice to fall back to
full page reloads if JavaScript wasn't available.

### htmx

[htmx](https://htmx.org) is a JavaScript library that allows us to do AJAX
calls without needing to write JavaScript. Requests are made using HTTP verbs,
and responses are HTML that can be placed directly onto the current page. In
this fashion, we are able to swap out sections of HTML as we desire.

htmx includes [it's own headers in
requests](https://htmx.org/reference/#request_headers). A server can
interrogate a request to see if it is htmx-enabled, and respond with either
HTML or perform a full reload. This enables us to achieve our graceful
degradation / progressive enhancement desire.

## Upgrading

### Include htmx

Add a script tag for `htmx` in the `top` block:

```diff
--- a/multistep-form/templates/home.tmpl
+++ b/multistep-form/templates/home.tmpl
@@ -2,6 +2,7 @@
     <!doctype html>
     <html>
         <head>
+            <script src="https://unpkg.com/htmx.org@1.9.10"></script>
         </head>

         <body>
```

### Templates return full pages or partials

Each template block currently describes a single page, which includes the `top`
and `bottom` templates (these describe the "top" and the "bottom" of a HTML
page):

```gohtml
{{block "form-one" .}}
    {{template "top" .}}
    <form ...>
    ...
     </form>
    {{template "bottom" .}}
{{end}}
```

We need to break these up, so that we have one block each for:
- the partial we wish to return if JavaScript is enabled;
- the entire page if JavaScript is not enabled.

So the above example would be broken into two blocks: the first for the
partial:

```gohtml
{{/* partial */}}
{{block "form-one" .}}
    <form ...>
    ...
    </form>
{{end}}
```

And the second for the page that includes the partial:

```gohtml
{{/* entire page */}}
{{bock "page-form-one" .}}
    {{template "top" .}}
    {{template "form-one" .}}
    {{template "bottom" .}}
{{end}}
```

### Add htmx `hx-boost`, `hx-target` and `hx-swap` attributes to each form

Each form must have some way of communicating to htmx that we wish to send an
AJAX request and swap in the response.

The [`hx-boost` attribute](https://htmx.org/attributes/hx-boost/)  tells htmx
to issue this request as an AJAX request. The [`hx-target`
attribute](https://htmx.org/attributes/hx-target/) tells htmx where to swap the
returned HTML into the current page. The [`hx-swap`
attribute](https://htmx.org/attributes/hx-swap/) tells htmx how to swap the new
element in:

```diff
-    <form method="POST" action="/form/one">
+    <form hx-boost="true" hx-target="this" hx-swap="outerHTML" method="POST" action="/form/one">
```

Here, we are turning on AJAX reqeusts for this form, and telling htmx to replace
this entire form, with the HTML that is returned from `/form/one`.

### Update server to check for htmx requests

The server must know what to return on each request. It will check for a htmx
request first:

```diff
    _ = r.ParseForm()

+   hxRequest := r.Header.Get("Hx-Request") == "true"
+   hxBoosted := r.Header.Get("Hx-Boosted") == "true"
```

And based on it's value, respond appropriately:

```diff
    switch {
        case next:
-           http.Redirect(w, r, "/form/two", http.StatusSeeOther)
+           if hxRequest && hxBoosted {
+               app.render(w, "form-two", nil, http.StatusOK)
+           } else {
+               http.Redirect(w, r, "/form/two", http.StatusSeeOther)
+           }
```

Here, if we have an htmx request, we only respond with the partial. Otherwise,
we send the entire page via redirect.

### Update the GET handler

Our GET handler needs to return pages, not partials:

```diff
-       app.render(w, "form-"+section, nil, http.StatusOK)
+       app.render(w, "page-form-"+section, nil, http.StatusOK)
```

## Notes

- Single entry point for form POSTs gets pretty messy pretty quickly. One
    endpoint per POST would tidy this up.
- Depending on the logic in the POST handler, you could end up with cyclees in
  your multi-step form.
- No validation in this demo, but you get the idea.
