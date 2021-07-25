package goisbn

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
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

	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	val := &googleBooksResponse{}
	r := bytes.NewReader([]byte(string(body)))
	decoder := json.NewDecoder(r)
	err = decoder.Decode(val)

	if err != nil || val.TotalItems == 0 {
		return nil
	}

	isbn10, isbn13 := "", ""
	for _, v := range val.Items[0].VolumeInfo.Identifier {
		if v.Type == "ISBN_10" {
			isbn10 = v.Identifier
		}
		if v.Type == "ISBN_13" {
			isbn13 = v.Identifier
		}
	}
	if isbn != isbn10 && isbn != isbn13 {
		return nil
	}
	b := val.Items[0].VolumeInfo
	book := &Book{
		IndustryIdentifiers: &Identifier{
			ISBN:   isbn,
			ISBN13: isbn13,
		},
		Title:   b.Title,
		Authors: b.Authors,
		ImageLinks: &ImageLinks{
			SmallImageURL: b.Image.SmallImageURL,
			ImageURL:      b.Image.ImageURL,
		},
		PublishedYear: b.PublicationYear,
		Description:   b.Description,
		PageCount:     b.PageCount,
		Categories:    b.Categories,
		Publisher:     b.Publisher,
		Language:      b.Language,
		Source:        PROVIDER_GOOGLE,
	}

	return book
}

func resolveOpenLibrary(isbn string) *Book {
	url := fmt.Sprintf("%s%s%s", OPENLIBRARY_API_BASE, OPENLIBRARY_API_BOOK, url.Values{"bibkeys": {"ISBN:" + isbn}, "format": {"json"}, "jscmd": {"data"}}.Encode())

	resp, err := http.Get(url)
	if err != nil {
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	key := fmt.Sprintf("ISBN:%s", isbn)
	var data map[string]openLibraryresponse
	err = json.Unmarshal([]byte(string(body)), &data)
	if err != nil {
		return nil
	}
	if _, ok := data[key]; !ok {
		return nil
	}
	authors := []string{}
	for _, v := range data[key].Authors {
		authors = append(authors, v.Name)
	}
	isbn10, isbn13 := "", ""
	publishers := []string{}
	if len(data[key].Identifiers.ISBN) > 0 {
		isbn10 = data[key].Identifiers.ISBN[0]
	}
	if len(data[key].Identifiers.ISBN13) > 0 {
		isbn13 = data[key].Identifiers.ISBN13[0]
	}
	if isbn10 != isbn && isbn13 != isbn {
		return nil
	}
	identifiers := &Identifier{
		ISBN:   isbn10,
		ISBN13: isbn13,
	}
	for _, v := range data[key].Publishers {
		publishers = append(publishers, v.Name)
	}
	book := &Book{
		Title:         data[key].Title,
		PublishedYear: data[key].PublishedYear,
		Authors:       authors,
		// Description: ,
		IndustryIdentifiers: identifiers,
		PageCount:           data[key].PageCount,
		// Categories: ,
		ImageLinks: &ImageLinks{
			SmallImageURL: data[key].Cover.Small,
			ImageURL:      data[key].Cover.Medium,
			LargeImageURL: data[key].Cover.Large,
		},
		Publisher: strings.Join(publishers, ", "),
		// Language: ,
		Source: PROVIDER_OPENLIBRARY,
	}

	return book

}

func resolveGoodreads(isbn string) *Book {
	envGoodread := os.Getenv("GOODREAD_APIKEY")
	url := fmt.Sprintf("%s%s%s", GOODREADS_API_BASE, GOODREADS_API_BOOK, url.Values{"q": {isbn}, "key": {envGoodread}}.Encode())

	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	xmlReader := bytes.NewReader([]byte(string(body)))
	xmlBook := new(goodreadsResponse)
	if err := xml.NewDecoder(xmlReader).Decode(xmlBook); err != nil {
		return nil
	}
	if xmlBook.Search.Results.Work.Book.Title == "" {
		return nil
	}
	b := xmlBook.Search.Results.Work.Book

	identifiers := &Identifier{}
	if Validate10(isbn) {
		identifiers.ISBN = isbn
	}
	if Validate13(isbn) {
		identifiers.ISBN13 = isbn
	}

	book := &Book{
		Title:         b.Title,
		PublishedYear: fmt.Sprintf("%d", xmlBook.Search.Results.Work.PublicationYear),
		Authors: []string{
			b.Author.Name,
		},
		// Description: ,
		IndustryIdentifiers: identifiers,
		// PageCount: ,
		// Categories: ,
		ImageLinks: &ImageLinks{
			SmallImageURL: b.SmallImageURL,
			ImageURL:      b.ImageURL,
		},
		// Publisher: ,
		// Language: ,
		Source: PROVIDER_GOODREADS,
	}

	return book
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
	for k := range uniqueProviders {
		if _, ok := PROVIDER_RESOLVERS[k]; ok {
			res = append(res, k)
		}
	}
	return res
}
