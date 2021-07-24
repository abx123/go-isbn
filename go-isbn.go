package goisbn

import "fmt"

const (
	GOOGLE_BOOKS_API_BASE = "https://www.googleapis.com"
	GOOGLE_BOOKS_API_BOOK = "/books/v1/volumes"

	OPENLIBRARY_API_BASE = "https://openlibrary.org"
	OPENLIBRARY_API_BOOK = "/api/books"

	WORLDCAT_API_BASE = "http://xisbn.worldcat.org"
	WORLDCAT_API_BOOK = "/webservices/xid/isbn"

	ISBNDB_API_BASE = "https://api2.isbndb.com"
	ISBNDB_API_BOOK = "/book"

	GOODREADS_API_BASE = ""
	GOODREADS_API_BOOK = ""

	PROVIDER_GOOGLE      = "google"
	PROVIDER_OPENLIBRARY = "openlibrary"
	PROVIDER_GOODREADS   = "goodreads"
	PROVIDER_WORLDCAT    = "worldcat"
	PROVIDER_ISBNDB      = "isbndb"
)

var DEFAULT_PROVIDERS = []string{
	PROVIDER_GOOGLE,
	PROVIDER_OPENLIBRARY,
	PROVIDER_GOODREADS,
	PROVIDER_WORLDCAT,
	PROVIDER_ISBNDB,
}

var PROVIDER_RESOLVERS = map[string]func(string) *Book{
	PROVIDER_GOOGLE:      resolveGoogle,
	PROVIDER_OPENLIBRARY: resolveOpenLibrary,
	PROVIDER_GOODREADS:   resolveGoodreads,
	PROVIDER_WORLDCAT:    resolveWorldcat,
	PROVIDER_ISBNDB:      resolveISBNDB,
}

type Book struct {
	Title               string `json:"title"`
	PublishedDate       string `json:"published_date"`
	Authors             string `json:"authors"`
	Description         string `json:"description"`
	IndustryIdentifiers string `json:"industry_identifiers"`
	PageCount           string `json:"page_count"`
	PrintType           string `json:"print_type"`
	Categories          string `json:"categories"`
	ImageLinks          string `json:"image_links"`
	PreviewLink         string `json:"preview_link"`
	InfoLink            string `json:"info_link"`
	Publisher           string `json:"publisher"`
	Language            string `json:"language"`
}

func resolveGoogle(isbn string) *Book {
	fmt.Println("resolveGoogle isbn:", isbn)
	return &Book{}
}

func resolveOpenLibrary(isbn string) *Book {
	fmt.Println("resolveOpenLibrary isbn:", isbn)
	return &Book{}
}

func resolveGoodreads(isbn string) *Book {
	fmt.Println("resolveGoodreads isbn:", isbn)
	return &Book{}
}

func resolveWorldcat(isbn string) *Book {
	fmt.Println("resolveWorldcat isbn:", isbn)
	return &Book{}
}

func resolveISBNDB(isbn string) *Book {
	fmt.Println("resolveISBNDB isbn:", isbn)
	return &Book{}
}

// func main() {

// 	for _, v := range DEFAULT_PROVIDERS {
// 		// PROVIDER_RESOLVERS[v]("testing")
// 		// switch v {
// 		// case PROVIDER_GOOGLE:
// 		PROVIDER_RESOLVERS[v]("testing")
// 		// case PROVIDER_OPENLIBRARY:
// 		//     PROVIDER_RESOLVERS[v]("testing")
// 		// case PROVIDER_GOODREADS:
// 		//     PROVIDER_RESOLVERS[v]("testing")
// 		// case PROVIDER_WORLDCAT:
// 		//     PROVIDER_RESOLVERS[v]("testing")
// 		// case PROVIDER_ISBNDB:
// 		//     PROVIDER_RESOLVERS[v]("testing")
// 		// case "g":
// 		//     v.(func(string, int))("astring", 42)
// 	}

// }

func GetBookInfo(isbn string, providers []string) []*Book {
	res := []*Book{}
	providers = resolveProviders(providers)
	for _, v := range providers {
		if f, ok := PROVIDER_RESOLVERS[v]; ok {
			book := f(isbn)
			res = append(res, book)
		}
	}
	return res
}


func resolveProviders(providers []string) []string{
	if len(providers) == 0{
		return DEFAULT_PROVIDERS
	}
	uniqueProviders := map[string]int{}
	res := []string{}
	// remove duplicates
	for _, v := range providers{
		uniqueProviders[v]++
	}
	// check if provider is valid
	for k, _ := range uniqueProviders{
		if _, ok := PROVIDER_RESOLVERS[k]; ok{
			res = append(res, k)
		}
	}
	return res
}
