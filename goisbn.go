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

	log "github.com/sirupsen/logrus"
)

// DEFAULT_PROVIDERS contains all available providers, ie: Google Books, Open
// Library, Goodreads, & ISBNDB
var DEFAULT_PROVIDERS = []string{
	Provider_Google,
	Provider_OpenLibrary,
	Provider_Goodreads,
	provider_Isbndb,
}

// GoISBN contains the providers and their respective resolver function with API
// Key for Goodreads and ISBNDB provider
type GoISBN struct {
	providers      []string
	goodreadAPIKey string
	isbndbAPIKey   string
	resolvers      map[string]func(string, chan *Book)
}

// NewGoISBN generates a new instance of GoISBN
func NewGoISBN(providers []string) *GoISBN {
	gi := &GoISBN{
		goodreadAPIKey: os.Getenv(goodreadsAPIKey),
		isbndbAPIKey:   os.Getenv(isbndbAPIKey),
	}
	gi.resolvers = map[string]func(string, chan *Book){
		Provider_Google:      (gi.resolveGoogle),
		Provider_OpenLibrary: (gi.resolveOpenLibrary),
		Provider_Goodreads:   (gi.resolveGoodreads),
		provider_Isbndb:      (gi.resolveISBNDB),
	}
	gi.providers = gi.resolveProviders()
	return gi
}

// Get retreives the details of a book with the ISBN provided from previously
// initialized providers
func (gi *GoISBN) Get(isbn string) (*Book, error) {

	if !gi.ValidateISBN(isbn) {
		log.Info("isbn %s provided is not valid", isbn)
		return nil, errInvalidISBN
	}

	book := &Book{}
	resolvedProviders := gi.resolveProviders()
	ch := make(chan *Book, len(resolvedProviders))
	respCount := 0
	for _, v := range resolvedProviders {
		go gi.resolvers[v](isbn, ch)
	}

	for book.Title == "" {
		if respCount == len(resolvedProviders) {
			log.Info("book with isbn %s not found from %s", isbn, strings.Join(resolvedProviders, ", "))
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

	client := http.Client{Timeout: timeout}
	req, err := http.NewRequest(get, url, nil)
	if err != nil {
		log.Warn("Error generating request for Google Books API: %s", err)
		ch <- nil
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Warn("Error retrieving book details from Google Books API: %s", err)
		ch <- nil
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Warn("Google Books API returns non 200 status. Status: %s", resp.Status)
		ch <- nil
		return

	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn("Error reading book details from Google Books API: %s", err)
		ch <- nil
		return
	}

	val := &googleBooksResponse{}
	r := bytes.NewReader([]byte(string(body)))
	decoder := json.NewDecoder(r)
	err = decoder.Decode(val)

	if err != nil {
		log.Warn("Error decoding response from Google Books API: %s", err)
		ch <- nil
		return
	}

	if val.TotalItems == 0 {
		log.Warn("Google Books API returns 0 item")
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
		log.Warn("Google Books API returns incorrect item, isbn10: %s, isbn13:%s", isbn10, isbn13)
		ch <- nil
		return
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
		Source:        Provider_Google,
	}

	ch <- book
}

func (gi *GoISBN) resolveOpenLibrary(isbn string, ch chan *Book) {
	url := fmt.Sprintf("%s%s%s", openLibraryAPIBase, openLibraryAPIBook, url.Values{"bibkeys": {"ISBN:" + isbn}, "format": {"json"}, "jscmd": {"data"}}.Encode())

	client := http.Client{Timeout: timeout}
	req, err := http.NewRequest(get, url, nil)
	if err != nil {
		log.Warn("Error generating request for Open Library API: %s", err)
		ch <- nil
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Warn("Error retrieving book details from Open Library API: %s", err)
		ch <- nil
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Warn("Open Library API returns non 200 status. Status: %s", err)
		ch <- nil
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn("Error reading book details from Open Library API: %s", err)
		ch <- nil
		return
	}

	key := fmt.Sprintf("ISBN:%s", isbn)
	var data map[string]openLibraryresponse
	err = json.Unmarshal([]byte(string(body)), &data)
	if err != nil {
		log.Warn("Error unmarshaling response from Open Library API: %s", err)
		ch <- nil
		return
	}
	if _, ok := data[key]; !ok {
		ch <- nil
		log.Warn("Open Library API returns incorrect item: %s", err)
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
	if isbn10 != isbn && isbn13 != isbn {
		log.Warn("Open Library API returns incorrect item, isbn10: %s, isbn13:%s", isbn10, isbn13)
		ch <- nil
		return
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
		Source: Provider_OpenLibrary,
	}

}

func (gi *GoISBN) resolveGoodreads(isbn string, ch chan *Book) {
	url := fmt.Sprintf("%s%s%s", goodreadsAPIBase, goodreadsAPIBook, url.Values{"q": {isbn}, "key": {gi.goodreadAPIKey}}.Encode())

	client := http.Client{Timeout: timeout}
	req, err := http.NewRequest(get, url, nil)
	if err != nil {
		log.Warn("Error generating request for Goodreads API: %s", err)
		ch <- nil
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Warn("Error retrieving book details from Goodreads API: %s", err)
		ch <- nil
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Warn("Goodreads API returns non 200 status. Status: %s", resp.Status)
		ch <- nil
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn("Error reading book details from Goodreads API: %s", err)
		ch <- nil
		return
	}
	xmlReader := bytes.NewReader([]byte(string(body)))
	xmlBook := new(goodreadsResponse)
	if err := xml.NewDecoder(xmlReader).Decode(xmlBook); err != nil {
		log.Warn("Error decoding response from Goodreads API: %s", err)
		ch <- nil
		return
	}
	if xmlBook.Search.Results.Work.Book.Title == "" {
		log.Warn("Goodreads API returns 0 item")
		ch <- nil
		return
	}
	b := xmlBook.Search.Results.Work.Book

	identifiers := &Identifier{}
	if validate10(isbn) {
		identifiers.ISBN = isbn
	}
	if validate13(isbn) {
		identifiers.ISBN13 = isbn
	}

	ch <- &Book{
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
		Source: Provider_Goodreads,
	}
}

func (gi *GoISBN) resolveISBNDB(isbn string, ch chan *Book) {
	url := fmt.Sprintf("%s%s%s", isbndbAPIBase, isbndbAPIBook, isbn)

	client := http.Client{Timeout: timeout}
	req, err := http.NewRequest(get, url, nil)
	if err != nil {
		log.Warn("Error generating request for ISBNDB API: %s", err)
		ch <- nil
		return
	}
	req.Header.Add(authorizationHeaderKey, gi.isbndbAPIKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Warn("Error retrieving book details from ISBNDB API: %s", err)
		ch <- nil
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Warn("ISBNDB API returns non 200 status. Status: %s", resp.Status)
		ch <- nil
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn("Error reading book details from ISBNDB API: %s", err)
		ch <- nil
		return
	}
	val := &isbndbResponse{}
	r := bytes.NewReader([]byte(string(body)))
	decoder := json.NewDecoder(r)
	err = decoder.Decode(val)
	if err != nil {
		log.Warn("Error decoding response from ISBNDB API: %s", err)
		ch <- nil
		return
	}
	if val.Book.ISBN != isbn && val.Book.ISBN13 != isbn {
		log.Warn("ISBNDB API returns incorrect item, isbn10: %s, isbn13:%s", val.Book.ISBN, val.Book.ISBN13)
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
		Source:    provider_Isbndb,
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
			if k == Provider_Goodreads && gi.goodreadAPIKey == "" {
				log.Info("Goodreads API Key not set, removing Goodreads from provider list")
				continue
			}
			if k == provider_Isbndb && gi.isbndbAPIKey == "" {
				log.Info("ISBNDB API Key not set, removing Isbndb from provider list")
				continue
			}
			res = append(res, k)
		}
	}
	return res
}
