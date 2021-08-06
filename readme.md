# go-isbn

[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/abx123/go-isbn)
[![Go Report Card](https://goreportcard.com/badge/github.com/abx123/go-isbn?style=flat-square)](https://goreportcard.com/report/github.com/abx123/go-isbn)
[![Codecov](https://img.shields.io/codecov/c/github/abx123/go-isbn.svg?style=flat-square)](https://codecov.io/gh/labx123/go-isbn)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/abx123/go-isbn/master/LICENSE)
[![Sourcegraph](https://sourcegraph.com/github.com/labx123/go-isbn/-/badge.svg?style=flat-square)](https://sourcegraph.com/github.com/abx123/go-isbn?badge)

## Feature Overview

- Retrieves book details using ISBN10 / ISBN13 via multiple providers:
  - Google Books
  - Open Library
  - Goodreads _(requires env var GOODREAD_APIKEY to be set) [free](https://www.goodreads.com/api)_
  - ISBNDB _(requires env var ISBNDB_APIKEY to be set) [7-day trial](https://isbndb.com/isbn-database)_
- Validates if a string is in valid ISBN10 / ISBN13 format

go-isbn will spawn equal number of go routines each querying a single provider with a max timeout of 3 seconds. First valid result will then be returned. Will return book not found only if all providers fail.

## Guide

### Installation

```sh
go get github.com/abx123/go-isbn
```

### Example

Querying on all 4 providers:

```go
package main

import (
  "fmt"
  "log"

  goisbn "github.com/abx123/go-isbn"
)

func main() {
  // go-isbn instance
  gi := goisbn.NewGoISBN(goisbn.DEFAULT_PROVIDERS)

  // Get book details
  book, err := gi.Get("9780099588986")
  if err != nil{
    log.Fatalln(err)
  }
  fmt.Println(book)
```

Querying on all selected providers:

```go
package main

import (
  "fmt"
  "log"

  goisbn "github.com/abx123/go-isbn"
)

func main() {
  // go-isbn instance
  gi := goisbn.NewGoISBN([]string{
    goisbn.ProviderGoogle,
    goisbn.ProviderGoodreads,
  })

  // Get book details
  book, err := gi.Get("9780099588986")
  if err != nil{
    log.Fatalln(err)
  }
  fmt.Println(book)
```
