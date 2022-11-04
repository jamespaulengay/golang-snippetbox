package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) { 
	if r.URL.Path != "/" {
		app.notFound(w) // Use the notFound() helper
		return
	}

	files := []string{ 
		"./ui/html/base.tmpl", 
		"./ui/html/partials/nav.tmpl", 
		"./ui/html/pages/home.tmpl",
	}
		
	ts, err := template.ParseFiles(files...) 

	if err != nil {
		app.serverError(w, err) // Use the serverError() helper.
		return
	}

	err = ts.ExecuteTemplate(w, "base", nil) 

	if err != nil {
		app.serverError(w, err) // Use the serverError() helper. 
	}
}


func (app *application) snippetView(w http.ResponseWriter, r *http.Request) { 
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil || id < 1 {
		app.notFound(w) // Use the notFound() helper.
		return
	}
	
	fmt.Fprintf(w, "Display a specific snippet with ID %d...", id) 
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