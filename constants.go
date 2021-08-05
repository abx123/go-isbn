package goisbn

const (
	googleBooks_Api_Base = "https://www.googleapis.com"
	googleBooks_Api_Book = "/books/v1/volumes?"

	OpenLibrary_Api_Base = "https://openlibrary.org"
	OpenLibrary_Api_Book = "/api/books?"

	isbndb_Api_Base = "https://api2.isbndb.com"
	isbndb_Api_Book = "/book/"

	goodreads_Api_Base = "https://www.goodreads.com"
	goodreads_Api_Book = "/search/index.xml?"

	provider_Google      = "google"
	provider_OpenLibrary = "openlibrary"
	provider_Goodreads   = "goodreads"
	Provider_Isbndb      = "isbndb"
)
