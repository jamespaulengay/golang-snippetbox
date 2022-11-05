package main

import (
	"net/http"

	"github.com/justinas/alice"
)

// Update the signature for the routes() method so that it returns a
// http.Handler instead of *http.ServeMux.
func (app *application) routes() http.Handler { 
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/")) 
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", app.home) 
	mux.HandleFunc("/snippet/view", app.snippetView) 
	mux.HandleFunc("/snippet/create", app.snippetCreate)
	

	// Create a middleware chain containing our 'standard' middleware
	// which will be used for every request our application receives. 
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Pass the servemux as the 'next' parameter to the secureHeaders middleware. 
	// Because secureHeaders is just a function, and the function returns a
	// http.Handler we don't need to do anything else.
	//secureHeaders → servemux → application handler → servemux → secureHeaders
	return standard.Then(mux)
}