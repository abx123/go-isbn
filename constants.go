package goisbn

import "time"

const (
	googleBooksAPIBase = "https://www.googleapis.com"
	googleBooksAPIBook = "/books/v1/volumes?"

	openLibraryAPIBase = "https://openlibrary.org"
	openLibraryAPIBook = "/api/books?"

	isbndbAPIBase = "https://api2.isbndb.com"
	isbndbAPIBook = "/book/"
	isbndbAPIKey  = "ISBNDB_APIKEY"

	goodreadsAPIBase = "https://www.goodreads.com"
	goodreadsAPIBook = "/search/index.xml?"
	goodreadsAPIKey  = "GOODREAD_APIKEY"

	Provider_Google      = "google"
	Provider_OpenLibrary = "openlibrary"
	Provider_Goodreads   = "goodreads"
	provider_Isbndb      = "isbndb"

	timeout = 3 * time.Second

	get = "GET"

	authorizationHeaderKey = "Authorization"
)
