package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

// a) Improved Logger
type application struct { 
	errorLog *log.Logger 
	infoLog *log.Logger
}

func main() {
	// Command-line flags
	addr := flag.String("addr", ":4000", "HTTP network address") 

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialize a new instance of our application struct, containing the 
	// dependencies.
	app := &application{
		errorLog: errorLog,
		infoLog: infoLog, 
	}

	srv := &http.Server{ 
		Addr: *addr,
		ErrorLog: errorLog,
		// Call the new app.routes() method to get the servemux containing our routes.
		Handler: app.routes(),
	}

	infoLog.Printf("Starting server on %s", *addr) 
	err := srv.ListenAndServe() 
	errorLog.Fatal(err)
}