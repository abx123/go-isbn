package goisbn

const (
	GOOGLE_BOOKS_API_BASE = "https://www.googleapis.com"
	GOOGLE_BOOKS_API_BOOK = "/books/v1/volumes"

	OPENLIBRARY_API_BASE = "https://openlibrary.org"
	OPENLIBRARY_API_BOOK = "/api/books"

	ISBNDB_API_BASE = "https://api2.isbndb.com"
	ISBNDB_API_BOOK = "/book"

	GOODREADS_API_BASE = "https://www.goodreads.com"
	GOODREADS_API_BOOK = "/search/index.xml"

	PROVIDER_GOOGLE      = "google"
	PROVIDER_OPENLIBRARY = "openlibrary"
	PROVIDER_GOODREADS   = "goodreads"
	PROVIDER_WORLDCAT    = "worldcat"
	PROVIDER_ISBNDB      = "isbndb"
)
