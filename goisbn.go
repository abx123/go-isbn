package goisbn

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// DEFAULT_PROVIDERS contains all available providers, ie: Google Books, Open
// Library, Goodreads, & ISBNDB
var DEFAULT_PROVIDERS = []string{
	ProviderGoogle,
	ProviderOpenLibrary,
	ProviderGoodreads,
	ProviderIsbndb,
}

// Queryer is the main interface for GoISBN
type Queryer interface {
	Get(string) (*Book, error)
	ValidateISBN(string) bool
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// GoISBN contains the providers and their respective resolver function with API
// Key for Goodreads and ISBNDB provider
type GoISBN struct {
	providers      []string
	goodreadAPIKey string
	isbndbAPIKey   string
	resolvers      map[string]func(string, chan *Book)
	client         httpClient
}

// NewGoISBN generates a new instance of GoISBN
func NewGoISBN(providers []string) *GoISBN {
	gi := &GoISBN{
		goodreadAPIKey: os.Getenv(goodreadsAPIKey),
		isbndbAPIKey:   os.Getenv(isbndbAPIKey),
		providers:      providers,
		client:         &http.Client{Timeout: timeout},
	}
	gi.resolvers = map[string]func(string, chan *Book){
		ProviderGoogle:      (gi.resolveGoogle),
		ProviderOpenLibrary: (gi.resolveOpenLibrary),
		ProviderGoodreads:   (gi.resolveGoodreads),
		ProviderIsbndb:      (gi.resolveISBNDB),
	}
	gi.providers = gi.resolveProviders()
	return gi
}

// Get retreives the details of a book with the ISBN provided from previously
// initialized providers
func (gi *GoISBN) Get(isbn string) (*Book, error) {

	if !gi.ValidateISBN(isbn) {
		log.Printf("isbn %s provided is not valid\n", isbn)
		return nil, errInvalidISBN
	}

	book := &Book{}
	ch := make(chan *Book, len(gi.providers))
	respCount := 0
	for _, v := range gi.providers {
		go gi.resolvers[v](isbn, ch)
	}

	for book.Title == "" {
		if respCount == len(gi.providers) {
			log.Printf("book with isbn %s not found from %s\n", isbn, strings.Join(gi.providers, ", "))
			return nil, errBookNotFound
		}
		tempBook := <-ch
		respCount++
		if tempBook != nil {
			book = tempBook
		}
	}
	return book, nil
}

// ValidateISBN checks if the input isbn is in a valid ISBN 10 or ISBN 13 format
func (gi *GoISBN) ValidateISBN(isbn string) bool {
	isbn = strings.ReplaceAll(strings.ReplaceAll(isbn, " ", ""), "-", "")
	switch len(isbn) {
	case 10:
		return validate10(isbn)
	case 13:
		return validate13(isbn)
	}
	return false
}

func (gi *GoISBN) resolveGoogle(isbn string, ch chan *Book) {
	url := fmt.Sprintf("%s%s%s", googleBooksAPIBase, googleBooksAPIBook, url.Values{"q": {isbn}}.Encode())

	req, _ := http.NewRequest(get, url, nil)
	resp, err := gi.client.Do(req)
	if err != nil {
		log.Printf("Error retrieving book details from Google Books API: %s\n", err)
		ch <- nil
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Printf("Google Books API returns non 200 status. Status: %s\n", resp.Status)
		ch <- nil
		return

	}
	val := &googleBooksResponse{}
	err = json.NewDecoder(resp.Body).Decode(&val)
	if err != nil {
		log.Printf("Error decoding response from Google Books API: %s\n", err)
		ch <- nil
		return
	}

	if val.TotalItems == 0 {
		log.Printf("Google Books API returns 0 item\n")
		ch <- nil
		return
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
		log.Printf("Google Books API returns incorrect item, isbn10: %s, isbn13:%s\n", isbn10, isbn13)
		ch <- nil
		return
	}
	b := val.Items[0].VolumeInfo
	book := &Book{
		IndustryIdentifiers: &Identifier{
			ISBN:   isbn10,
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
		Source:        ProviderGoogle,
	}

	ch <- book
}

func (gi *GoISBN) resolveOpenLibrary(isbn string, ch chan *Book) {
	url := fmt.Sprintf("%s%s%s", openLibraryAPIBase, openLibraryAPIBook, url.Values{"bibkeys": {"ISBN:" + isbn}, "format": {"json"}, "jscmd": {"data"}}.Encode())

	req, _ := http.NewRequest(get, url, nil)
	resp, err := gi.client.Do(req)
	if err != nil {
		log.Printf("Error retrieving book details from Open Library API: %s\n", err)
		ch <- nil
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Printf("Open Library API returns non 200 status. Status: %s\n", err)
		ch <- nil
		return
	}
	key := fmt.Sprintf("ISBN:%s", isbn)
	data := map[string]openLibraryresponse{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Printf("Error unmarshaling response from Open Library API: %s\n", err)
		ch <- nil
		return
	}
	if _, ok := data[key]; !ok {
		ch <- nil
		log.Printf("Open Library API returns incorrect item: %s\n", err)
		return
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
	identifiers := &Identifier{
		ISBN:   isbn10,
		ISBN13: isbn13,
	}
	for _, v := range data[key].Publishers {
		publishers = append(publishers, v.Name)
	}
	ch <- &Book{
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
		Source: ProviderOpenLibrary,
	}

}

func (gi *GoISBN) resolveGoodreads(isbn string, ch chan *Book) {
	url := fmt.Sprintf("%s%s%s", goodreadsAPIBase, goodreadsAPIBook, url.Values{"q": {isbn}, "key": {gi.goodreadAPIKey}}.Encode())

	req, _ := http.NewRequest(get, url, nil)
	resp, err := gi.client.Do(req)
	if err != nil {
		log.Printf("Error retrieving book details from Goodreads API: %s\n", err)
		ch <- nil
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Printf("Goodreads API returns non 200 status. Status: %s\n", resp.Status)
		ch <- nil
		return
	}
	val := &goodreadsResponse{}
	if err := xml.NewDecoder(resp.Body).Decode(val); err != nil {
		log.Printf("Error decoding response from Goodreads API: %s\n", err)
		ch <- nil
		return
	}
	if val.Search.Results.Work.Book.Title == "" {
		log.Printf("Goodreads API returns 0 item\n")
		ch <- nil
		return
	}
	b := val.Search.Results.Work.Book

	identifiers := &Identifier{}
	if validate10(isbn) {
		identifiers.ISBN = isbn
	}
	if validate13(isbn) {
		identifiers.ISBN13 = isbn
	}

	ch <- &Book{
		Title:         b.Title,
		PublishedYear: fmt.Sprintf("%d", val.Search.Results.Work.PublicationYear),
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
		Source: ProviderGoodreads,
	}
}

func (gi *GoISBN) resolveISBNDB(isbn string, ch chan *Book) {
	url := fmt.Sprintf("%s%s%s", isbndbAPIBase, isbndbAPIBook, isbn)

	req, _ := http.NewRequest(get, url, nil)
	req.Header.Add(authorizationHeaderKey, gi.isbndbAPIKey)
	resp, err := gi.client.Do(req)
	if err != nil {
		log.Printf("Error retrieving book details from ISBNDB API: %s\n", err)
		ch <- nil
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Printf("ISBNDB API returns non 200 status. Status: %s\n", resp.Status)
		ch <- nil
		return
	}
	val := &isbndbResponse{}
	err = json.NewDecoder(resp.Body).Decode(&val)
	if err != nil {
		log.Printf("Error decoding response from ISBNDB API: %s\n", err)
		ch <- nil
		return
	}
	if val.Book.ISBN != isbn && val.Book.ISBN13 != isbn {
		log.Printf("ISBNDB API returns incorrect item, isbn10: %s, isbn13:%s\n", val.Book.ISBN, val.Book.ISBN13)
		ch <- nil
		return
	}

	ch <- &Book{
		Title:         val.Book.Title,
		PublishedYear: val.Book.PublishedDate,
		Authors:       val.Book.Authors,
		// Description: ,
		IndustryIdentifiers: &Identifier{
			ISBN:   val.Book.ISBN,
			ISBN13: val.Book.ISBN13,
		},
		// PageCount: ,
		// Categories: ,
		ImageLinks: &ImageLinks{
			SmallImageURL: val.Book.Image,
		},
		Publisher: val.Book.Publisher,
		Language:  val.Book.Language,
		Source:    ProviderIsbndb,
	}
}

func (gi *GoISBN) resolveProviders() []string {
	if len(gi.providers) == 0 {
		return DEFAULT_PROVIDERS
	}
	uniqueProviders := map[string]int{}
	res := []string{}
	// remove duplicates
	for _, v := range gi.providers {
		uniqueProviders[v]++
	}
	// check if provider is valid
	for k := range uniqueProviders {
		if _, ok := gi.resolvers[k]; ok {
			if k == ProviderGoodreads && gi.goodreadAPIKey == "" {
				log.Printf("Goodreads API Key not set, removing Goodreads from provider list\n")
				continue
			}
			if k == ProviderIsbndb && gi.isbndbAPIKey == "" {
				log.Printf("ISBNDB API Key not set, removing Isbndb from provider list\n")
				continue
			}
			res = append(res, k)
		}
	}
	return res
}
