package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"snippetbox.jamespaul.com/internal/models"
)

// a) Improved Logger // Inject dependencies
type application struct { 
	errorLog *log.Logger 
	infoLog *log.Logger
	snippets *models.SnippetModel
	templateCache map[string]*template.Template
	formDecoder *form.Decoder
	sessionManager *scs.SessionManager
}


func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")
	
	flag.Parse()
	
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	
	db, err := openDB(*dsn) 

	if err != nil {
		errorLog.Fatal(err) 
	}

	defer db.Close()

	// Initialize a new template cache...
	templateCache, err := newTemplateCache() 
	if err != nil {
		errorLog.Fatal(err) 
	}

	// Initialize a decoder instance...
	formDecoder := form.NewDecoder()

	// Use the scs.New() function to initialize a new session manager. Then we 
	// configure it to use our MySQL database as the session store, and set a 
	// lifetime of 12 hours (so that sessions automatically expire 12 hours
	// after first being created).
	sessionManager := scs.New() 
	sessionManager.Store = mysqlstore.New(db) 
	sessionManager.Lifetime = 12 * time.Hour

	// Initialize a models.SnippetModel instance and add it to the application 
	// dependencies.
	app := &application{
		errorLog: errorLog,
		infoLog: infoLog,
		snippets: &models.SnippetModel{DB: db},
		templateCache: templateCache,
		formDecoder: formDecoder,
		sessionManager: sessionManager,
	}

	srv := &http.Server{ 
		Addr: *addr,
		ErrorLog: errorLog,
		Handler: app.routes(), 
	}

	infoLog.Printf("Starting server on http://localhost%s", *addr) 
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

// for a given DSN.
func openDB(dsn string) (*sql.DB, error) { 
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err 
	}

	if err = db.Ping(); err != nil { 
		return nil, err
	}

	return db, nil 
}