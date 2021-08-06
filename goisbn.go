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
	"time"
)

var DEFAULT_PROVIDERS = []string{
	provider_Google,
	provider_OpenLibrary,
	provider_Goodreads,
	Provider_Isbndb,
}

type Queryer interface {
	Get(string) (*Book, error)
	ValidateISBN(string) bool
	resolveProviders() []string
}

type GoISBN struct {
	providers      []string
	goodreadAPIKey string
	isbndbAPIKey   string
	resolvers      map[string]func(string, chan *Book)
}

func NewGoISBN(providers []string) *GoISBN {
	gi := &GoISBN{
		goodreadAPIKey: os.Getenv("GOODREAD_APIKEY"),
		isbndbAPIKey:   os.Getenv("ISBNDB_APIKEY"),
	}
	gi.resolvers = map[string]func(string, chan *Book){
		provider_Google:      (gi.resolveGoogle),
		provider_OpenLibrary: (gi.resolveOpenLibrary),
		provider_Goodreads:   (gi.resolveGoodreads),
		Provider_Isbndb:      (gi.resolveISBNDB),
	}
	gi.providers = gi.resolveProviders()
	return gi
}

func (gi *GoISBN) Get(isbn string) (*Book, error) {

	if !gi.ValidateISBN(isbn) {
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
	url := fmt.Sprintf("%s%s%s", googleBooks_Api_Base, googleBooks_Api_Book, url.Values{"q": {isbn}}.Encode())

	client := http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		ch <- nil
		return
	}
	resp, err := client.Do(req)
	if err != nil || (resp.StatusCode < 200 || resp.StatusCode > 299) {
		ch <- nil
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch <- nil
		return
	}

	val := &googleBooksResponse{}
	r := bytes.NewReader([]byte(string(body)))
	decoder := json.NewDecoder(r)
	err = decoder.Decode(val)

	if err != nil || val.TotalItems == 0 {
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
		Source:        provider_Google,
	}

	ch <- book
}

func (gi *GoISBN) resolveOpenLibrary(isbn string, ch chan *Book) {
	url := fmt.Sprintf("%s%s%s", OpenLibrary_Api_Base, OpenLibrary_Api_Book, url.Values{"bibkeys": {"ISBN:" + isbn}, "format": {"json"}, "jscmd": {"data"}}.Encode())

	client := http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		ch <- nil
		return
	}
	resp, err := client.Do(req)
	if err != nil || (resp.StatusCode < 200 || resp.StatusCode > 299) {
		ch <- nil
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch <- nil
		return
	}

	key := fmt.Sprintf("ISBN:%s", isbn)
	var data map[string]openLibraryresponse
	err = json.Unmarshal([]byte(string(body)), &data)
	if err != nil {
		ch <- nil
		return
	}
	if _, ok := data[key]; !ok {
		ch <- nil
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
		Source: provider_OpenLibrary,
	}

}

func (gi *GoISBN) resolveGoodreads(isbn string, ch chan *Book) {
	url := fmt.Sprintf("%s%s%s", goodreads_Api_Base, goodreads_Api_Book, url.Values{"q": {isbn}, "key": {gi.goodreadAPIKey}}.Encode())

	client := http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		ch <- nil
		return
	}
	resp, err := client.Do(req)
	if err != nil || (resp.StatusCode < 200 || resp.StatusCode > 299) {
		ch <- nil
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch <- nil
		return
	}
	xmlReader := bytes.NewReader([]byte(string(body)))
	xmlBook := new(goodreadsResponse)
	if err := xml.NewDecoder(xmlReader).Decode(xmlBook); err != nil {
		ch <- nil
		return
	}
	if xmlBook.Search.Results.Work.Book.Title == "" {
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
		Source: provider_Goodreads,
	}
}

func (gi *GoISBN) resolveISBNDB(isbn string, ch chan *Book) {
	url := fmt.Sprintf("%s%s%s", isbndb_Api_Base, isbndb_Api_Book, isbn)

	client := http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		ch <- nil
		return
	}
	req.Header.Add("Authorization", gi.isbndbAPIKey)
	resp, err := client.Do(req)
	if err != nil || (resp.StatusCode < 200 || resp.StatusCode > 299) {
		ch <- nil
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch <- nil
		return
	}
	val := &isbndbResponse{}
	r := bytes.NewReader([]byte(string(body)))
	decoder := json.NewDecoder(r)
	err = decoder.Decode(val)
	if err != nil || (val.Book.ISBN != isbn && val.Book.ISBN13 != isbn) {
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
		Source:    Provider_Isbndb,
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
			if (k == provider_Goodreads && gi.goodreadAPIKey == "") || (k == Provider_Isbndb && gi.isbndbAPIKey == "") {
				continue
			}
			res = append(res, k)
		}
	}
	return res
}
