package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/julienschmidt/httprouter"
	"snippetbox.jamespaul.com/internal/models"
)

// Define a snippetCreateForm struct to represent the form data and validation
// errors for the form fields. Note that all the struct fields are deliberately
// exported (i.e. start with a capital letter). This is because struct fields
// must be exported in order to be read by the html/template package when
// rendering the template.
type snippetCreateForm struct {
	Title string
	Content string
	Expires int
	FieldErrors map[string]string
}
	

func (app *application) home(w http.ResponseWriter, r *http.Request) { 
	snippets, err := app.snippets.Latest() 
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Call the newTemplateData() helper to get a templateData struct containing 
	// the 'default' data (which for now is just the current year), and add the 
	// snippets slice to it.
	data := app.newTemplateData(r)
	data.Snippets = snippets

	// Use the new render helper.
	app.render(w, http.StatusOK, "home.tmpl", data)
}


func (app *application) snippetView(w http.ResponseWriter, r *http.Request) { 
	// When httprouter is parsing a request, the values of any named parameters 
	// will be stored in the request context. We'll talk about request context 
	// in detail later in the book, but for now it's enough to know that you can 
	// use the ParamsFromContext() function to retrieve a slice containing these 
	// parameter names and values like so:
	params := httprouter.ParamsFromContext(r.Context())

	// We can then use the ByName() method to get the value of the "id" named 
	// parameter from the slice and validate it as normal.
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}


	snippet, err := app.snippets.Get(id) 

	if err != nil {
		if errors.Is(err, models.ErrNoRecord) { 
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// And do the same thing again here...
	data := app.newTemplateData(r) 
	data.Snippet = snippet
	
	app.render(w, http.StatusOK, "view.tmpl", data)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) { 
	data := app.newTemplateData(r)
	// Initialize a new createSnippetForm instance and pass it to the template. 
	// Notice how this is also a great opportunity to set any default or
	// 'initial' values for the form --- here we set the initial value for the 
	// snippet expiry to 365 days.
	data.Form = snippetCreateForm{ Expires: 365,}
	app.render(w, http.StatusOK, "create.tmpl", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) { 
	// First we call r.ParseForm() which adds any data in POST request bodies
	// to the r.PostForm map. This also works in the same way for PUT and PATCH
	// requests. If there are any errors, we use our app.ClientError() helper to
	// send a 400 Bad Request response to the user.
	err := r.ParseForm() 
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Use the r.PostForm.Get() method to retrieve the title and content 
	// from the r.PostForm map.
	title := r.PostForm.Get("title")
	content := r.PostForm.Get("content")

	// The r.PostForm.Get() method always returns the form data as a *string*. 
	// However, we're expecting our expires value to be a number, and want to 
	// represent it in our Go code as an integer. So we need to manually covert 
	// the form data to an integer using strconv.Atoi(), and we send a 400 Bad 
	// Request response if the conversion fails.
	expires, err := strconv.Atoi(r.PostForm.Get("expires")) 
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Create an instance of the snippetCreateForm struct containing the values 
	// from the form and an empty map for any validation errors.
	form := snippetCreateForm{
		Title: r.PostForm.Get("title"),
		Content: r.PostForm.Get("content"),
		Expires: expires,
		FieldErrors: map[string]string{},
	}


	// Initialize a map to hold any validation errors for the form fields.
	fieldErrors := make(map[string]string)

	// Check that the title value is not blank and is not more than 100
	// characters long. If it fails either of those checks, add a message to the // errors map using the field name as the key.
	if strings.TrimSpace(title) == "" {
		form.FieldErrors["title"] = "This field cannot be blank" 
	} else if utf8.RuneCountInString(title) > 100 {
		form.FieldErrors["title"] = "This field cannot be more than 100 characters long" 
	}

	// Check that the Content value isn't blank.
	if strings.TrimSpace(content) == "" { 
		fieldErrors["content"] = "This field cannot be blank"
	}	
	// Check the expires value matches one of the permitted values (1, 7 or 
	// 365).
	if expires != 1 && expires != 7 && expires != 365 {
		fieldErrors["expires"] = "This field must equal 1, 7 or 365" 
	}

	// If there are any errors, dump them in a plain text HTTP response and 
	// return from the handler.
	if len(fieldErrors) > 0 {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.tmpl", data) 
		return
	}

	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)

	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}


