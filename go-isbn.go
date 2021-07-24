package goisbn

import (
	"fmt"
	"net/url"
	"os"
)

var DEFAULT_PROVIDERS = []string{
	PROVIDER_GOOGLE,
	PROVIDER_OPENLIBRARY,
	PROVIDER_GOODREADS,
	PROVIDER_ISBNDB,
}

var PROVIDER_RESOLVERS = map[string]func(string) *Book{
	PROVIDER_GOOGLE:      resolveGoogle,
	PROVIDER_OPENLIBRARY: resolveOpenLibrary,
	PROVIDER_GOODREADS:   resolveGoodreads,
	PROVIDER_ISBNDB:      resolveISBNDB,
}

func resolveGoogle(isbn string) *Book {
	url := fmt.Sprintf("%s%s%s", GOOGLE_BOOKS_API_BASE, GOOGLE_BOOKS_API_BOOK, url.Values{"q": {isbn}}.Encode())
	fmt.Println("resolveGoogle isbn:", isbn)
	fmt.Println("url:", url)
	// constant.GOODREADS_API_BASE
	return &Book{}
}

func resolveOpenLibrary(isbn string) *Book {
	// https://openlibrary.org/api/books?bibkeys=ISBN:9780099588986&format=json&jscmd=data
	url := fmt.Sprintf("%s%s%s", OPENLIBRARY_API_BASE, OPENLIBRARY_API_BOOK, url.Values{"bibkeys": {"ISBN:" + isbn}, "format": {"json"}, "jscmd": {"data"}}.Encode())
	// https://openlibrary.org/api/books?bibkeys=ISBN:0451526538
	// `${OPENLIBRARY_API_BASE + OPENLIBRARY_API_BOOK}?bibkeys=ISBN:${isbn}&format=json&jscmd=details`
	fmt.Println("resolveOpenLibrary isbn:", isbn)
	fmt.Println("url:", url)
	return &Book{}
}

func resolveGoodreads(isbn string) *Book {
	envGoodread := os.Getenv("GOODREAD_APIKEY")
	url := fmt.Sprintf("%s%s%s", GOODREADS_API_BASE, GOODREADS_API_BOOK, url.Values{"q": {isbn}, "key": {envGoodread}}.Encode())
	// https://www.goodreads.com/search/index.xml?q=9780099588986&key=6qVbqOjnzhHws97M5gYYA
	fmt.Println("resolveGoodreads isbn:", isbn)
	fmt.Println("url:", url)
	return &Book{}
}

func resolveISBNDB(isbn string) *Book {
	fmt.Println("resolveISBNDB isbn:", isbn)
	return &Book{}
}

func GetBookInfo(isbn string, providers []string) ([]*Book, error) {
	if !Validate(isbn) {
		return nil, errInvalidISBN
	}
	books := []*Book{}
	resolvedProviders := resolveProviders(providers)
	for _, v := range resolvedProviders {
		book := PROVIDER_RESOLVERS[v](isbn)
		if book != nil {
			books = append(books, book)
		}
	}
	if len(books) == 0 {
		return nil, errBookNotFound
	}
	return books, nil
}

func resolveProviders(providers []string) []string {
	if len(providers) == 0 {
		return DEFAULT_PROVIDERS
	}
	uniqueProviders := map[string]int{}
	res := []string{}
	// remove duplicates
	for _, v := range providers {
		uniqueProviders[v]++
	}
	// check if provider is valid
	for k, _ := range uniqueProviders {
		if _, ok := PROVIDER_RESOLVERS[k]; ok {
			res = append(res, k)
		}
	}
	return res
}
