package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"snippetbox.jamespaul.com/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) { 
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}


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
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
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
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost) 
		app.clientError(w, http.StatusMethodNotAllowed) 
		return
	}

	// Create some variables holding dummy data. We'll remove these later on
	// during the build.
	title := "O snail"
	content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n– Kobayashi Issa" 
	expires := 7

	// Pass the data to the SnippetModel.Insert() method, receiving the // ID of the new record back.
	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
		return
	}
	
	// Redirect the user to the relevant page for the snippet.
	http.Redirect(w, r, fmt.Sprintf("/snippet/view?id=%d", id), http.StatusSeeOther)
}


