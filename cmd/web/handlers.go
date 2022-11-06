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

// Create a new userSignupForm struct.
type userSignupForm struct {
	Name string `form:"name"`
	Email string `form:"email"`
	Password string `form:"password"`
	validator.Validator `form:"-"`
}

// Create a new userLoginForm struct.
type userLoginForm struct {
	Email string `form:"email"` 
	Password string `form:"password"` 
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





func (app *application) userSignup(w http.ResponseWriter, r *http.Request) { 
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, http.StatusOK, "signup.tmpl", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) { 
	// Declare an zero-valued instance of our userSignupForm struct.
	var form userSignupForm

	// Parse the form data into the userSignupForm struct.
	err := app.decodePostForm(r, &form) 
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form contents using our helper functions.
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank") 
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address") 
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank") 
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")
	
	// If there are any errors, redisplay the signup form along with a 422 
	// status code.
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data) 
		return
	}

	// Try to create a new user record in the database. If the email already 
	// exists then add an error message to the form and re-display it.
	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) { 
			form.AddFieldError("email", "Email address is already in use")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			app.serverError(w, err)
		}

		return
	}
	// Otherwise add a confirmation flash message to the session confirming that
	// their signup worked.
	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")
	
	// And redirect the user to the login page.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) { 
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, http.StatusOK, "login.tmpl", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) { 
	// Decode the form data into the userLoginForm struct.
	var form userLoginForm

	err := app.decodePostForm(r, &form) 
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Do some validation checks on the form. We check that both email and
	// password are provided, and also check the format of the email address as
	// a UX-nicety (in case the user makes a typo).
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank") 
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address") 
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data) 
		return
	}

	// Check whether the credentials are valid. If they're not, add a generic 
	// non-field error message and re-display the login page.
	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) { 
			form.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Use the RenewToken() method on the current session to change the session 
	// ID. It's good practice to generate a new session ID when the
	// authentication state or privilege levels changes for the user (e.g. login 
	// and logout operations).
	err = app.sessionManager.RenewToken(r.Context()) 
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Add the ID of the current user to the session, so that they are now 
	// 'logged in'.
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	
	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) { 
	// Use the RenewToken() method on the current session to change the session 
	// ID again.
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil { 
		app.serverError(w, err) 
		return
	}
	// Remove the authenticatedUserID from the session data so that the user is 
	// 'logged out'.
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	// Add a flash message to the session to confirm to the user that they've been
	// logged out.
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
	
	// Redirect the user to the application home page.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}