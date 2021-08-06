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

	// ProviderGoogle is the constant representation for Google Books
	ProviderGoogle      = "google"
	// ProviderOpenLibrary is the constant representation for Open Library
	ProviderOpenLibrary = "openlibrary"
	// ProviderGoodreads is the constant representation for Goodreads
	ProviderGoodreads   = "goodreads"
	// ProviderIsbndb is the constant representation for ISBNDB
	ProviderIsbndb      = "isbndb"

	timeout = 3 * time.Second

	get = "GET"

	authorizationHeaderKey = "Authorization"
)
