package main

import (
	"database/sql"
	"net/url"
	"os"
	"strings"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// newRemoteStore opens the pure-Go, CGo-free Turso/libSQL remote backend over
// hrana (libsql://, ws(s)://, http(s)://). The auth token is read from the
// TURSO_AUTH_TOKEN environment variable and appended to the DSN as an authToken
// query parameter (unless the caller already supplied one in the URL), so no
// token has to be threaded through the command signatures.
func newRemoteStore(rawURL string) (*sqlStore, error) {
	dsn := rawURL

	token := os.Getenv("TURSO_AUTH_TOKEN")
	if token != "" && !strings.Contains(rawURL, "authToken=") {
		sep := "?"
		if strings.Contains(rawURL, "?") {
			sep = "&"
		}
		dsn = rawURL + sep + "authToken=" + url.QueryEscape(token)
	}

	db, err := sql.Open("libsql", dsn)
	if err != nil {
		return nil, err
	}
	return &sqlStore{db: db}, nil
}
