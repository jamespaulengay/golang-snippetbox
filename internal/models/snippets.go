package models

import (
	"database/sql"
	"time"
)

// Define a Snippet type to hold the data for an individual snippet. Notice how
// the fields of the struct correspond to the fields in our MySQL snippets
// table?
type Snippet struct {
	ID int
	Title string
	Content string
	Created time.Time
	Expires time.Time
}


// Define a SnippetModel type which wraps a sql.DB connection pool
type SnippetModel struct { 
	DB *sql.DB
}


// This will return a specific snippet based on its id.
func (m *SnippetModel) Get(id int) (*Snippet, error) { 
	return nil, nil
}

// This will return the 10 most recently created snippets.
func (m *SnippetModel) Latest() ([]*Snippet, error) { 
	return nil, nil
}

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) { 
	// Write the SQL statement we want to execute. I've split it over two lines
	// for readability (which is why it's surrounded with backquotes instead
	// of normal double quotes).
	stmt := `INSERT INTO snippets (title, content, created, expires) 
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`
	
	// Use the Exec() method on the embedded connection pool to execute the 
	// statement. The first parameter is the SQL statement, followed by the 
	// title, content and expiry values for the placeholder parameters. This 
	// method returns a sql.Result type, which contains some basic
	// information about what happened when the statement was executed.
	result, err := m.DB.Exec(stmt, title, content, expires) 
	if err != nil {
		return 0, err 
	}

	// Use the LastInsertId() method on the result to get the ID of our 
	// newly inserted record in the snippets table.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// The ID returned has the type int64, so we convert it to an int type 
	// before returning.
	return int(id), nil
}