package goisbn

import "time"

const (
	googleBooks_Api_Base = "https://www.googleapis.com"
	googleBooks_Api_Book = "/books/v1/volumes?"

	openLibrary_Api_Base = "https://openlibrary.org"
	openLibrary_Api_Book = "/api/books?"

	isbndb_Api_Base = "https://api2.isbndb.com"
	isbndb_Api_Book = "/book/"

	goodreads_Api_Base = "https://www.goodreads.com"
	goodreads_Api_Book = "/search/index.xml?"

	provider_Google      = "google"
	provider_OpenLibrary = "openlibrary"
	provider_Goodreads   = "goodreads"
	provider_Isbndb      = "isbndb"

	timeout = 3 * time.Second

	get = "GET"

	authorizationHeaderKey = "Authorization"
)
