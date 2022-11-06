package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"snippetbox.jamespaul.com/internal/models"
	"snippetbox.jamespaul.com/internal/validator"
)

// Define a snippetCreateForm struct to represent the form data and validation
// errors for the form fields. Note that all the struct fields are deliberately
// exported (i.e. start with a capital letter). This is because struct fields
// must be exported in order to be read by the html/template package when
// rendering the template.

// Update our snippetCreateForm struct to include struct tags which tell the
// decoder how to map HTML form values into the different struct fields. So, for // example, here we're telling the decoder to store the value from the HTML form // input with the name "title" in the Title field. The struct tag `form:"-"`
// tells the decoder to completely ignore a field during decoding.
type snippetCreateForm struct {
	Title string `form:"title"` 
	Content string `form:"content"` 
	Expires int `form:"expires"` 
	validator.Validator `form:"-"`
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

	// Use the PopString() method to retrieve the value for the "flash" key. 
	// PopString() also deletes the key and value from the session data, so it 
	// acts like a one-time fetch. If there is no matching key in the session 
	// data this will return the empty string.
	flash := app.sessionManager.PopString(r.Context(), "flash")	

	// And do the same thing again here...
	data := app.newTemplateData(r) 
	data.Snippet = snippet

	// Pass the flash message to the template.
	data.Flash = flash
	
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

	// Declare a new empty instance of the snippetCreateForm struct.
	var form snippetCreateForm

	// Call the Decode() method of the form decoder, passing in the current
	// request and *a pointer* to our snippetCreateForm struct. This will
	// essentially fill our struct with the relevant values from the HTML form. 
	// If there is a problem, we return a 400 Bad Request response to the client. 
	err = app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Then validate and use the data as normal...
	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank") 
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long") 
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank") 
	form.CheckField(validator.PermittedInt(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")	


	// Use the Valid() method to see if any of the checks failed. If they did, 
	// then re-render the template passing in the form in the same way as
	// before.
	if !form.Valid() {
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


	// Use the Put() method to add a string value ("Snippet successfully
	// created!") and the corresponding key ("flash") to the session data. 
	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}


