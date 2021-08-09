package goisbn

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Custom type that allows setting the func that our Mock Do func will run instead
type MockDoType func(req *http.Request) (*http.Response, error)

// MockClient is the mock client
type MockClient struct {
	MockDo MockDoType
}

// Overriding what the Do function should "do" in our MockClient
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.MockDo(req)
}

func TestNewGoISBN(t *testing.T) {
	type testCase struct {
		name            string
		desc            string
		providers       []string
		lenProviders    int
		goodreadsAPIKey string
		isbnDBAPIKey    string
		lenResolvers    []string
	}
	testCases := []testCase{
		{
			name:            "Happy Case",
			desc:            "all ok",
			providers:       DEFAULT_PROVIDERS,
			lenProviders:    4,
			goodreadsAPIKey: "dummy goodread key",
			isbnDBAPIKey:    "dummy isbnapi key",
			lenResolvers:    []string{"goodreads", "google", "isbndb", "openlibrary"},
		},
		{
			name:            "Happy Case",
			desc:            "no providers supplied, defaulting to default providers",
			providers:       []string{},
			lenProviders:    4,
			goodreadsAPIKey: "dummy goodread key",
			isbnDBAPIKey:    "dummy isbnapi key",
			lenResolvers:    []string{"goodreads", "google", "isbndb", "openlibrary"},
		},
		{
			name:            "Happy Case",
			desc:            "goodreads api key not supplied, removing goodreads from provider list",
			providers:       DEFAULT_PROVIDERS,
			lenProviders:    3,
			goodreadsAPIKey: "",
			isbnDBAPIKey:    "dummy isbnapi key",
			lenResolvers:    []string{"google", "isbndb", "openlibrary"},
		},
		{
			name:            "Happy Case",
			desc:            "isbndb api key not supplied, removing isbndb from provider list",
			providers:       DEFAULT_PROVIDERS,
			lenProviders:    3,
			goodreadsAPIKey: "dummy goodread key",
			isbnDBAPIKey:    "",
			lenResolvers:    []string{"goodreads", "google", "openlibrary"},
		},
	}
	for _, v := range testCases {
		defer unsetEnv()
		os.Setenv(goodreadsAPIKey, v.goodreadsAPIKey)
		os.Setenv(isbndbAPIKey, v.isbnDBAPIKey)
		gi := NewGoISBN(v.providers)
		assert.Equal(t, len(gi.providers), v.lenProviders)
		assert.Equal(t, gi.goodreadAPIKey, v.goodreadsAPIKey)
		assert.Equal(t, gi.isbndbAPIKey, v.isbnDBAPIKey)
		for _, val := range v.lenResolvers {
			_, ok := gi.resolvers[val]
			assert.Equal(t, ok, true)
		}
	}
}

func TestValidateISBN(t *testing.T) {
	type testCase struct {
		name   string
		desc   string
		isbn   string
		expRes bool
	}
	testCases := []testCase{
		{
			name:   "Happy Case",
			desc:   "all ok, valid isbn 13",
			isbn:   "9780099588986",
			expRes: true,
		},
		{
			name:   "Happy Case",
			desc:   "all ok, valid isbn 10",
			isbn:   "0099588986",
			expRes: true,
		},
		{
			name:   "Sad Case",
			desc:   "invalid length",
			isbn:   "00995889862",
			expRes: false,
		},
	}

	gi := NewGoISBN(DEFAULT_PROVIDERS)

	for _, v := range testCases {

		actRes := gi.ValidateISBN(v.isbn)

		assert.Equal(t, v.expRes, actRes)
	}
}

func TestResolveGoogle(t *testing.T) {
	type testCase struct {
		name string
		desc string
		jsonResp string
		expRes   *Book
		err      error
		respCode int
	}
	testCases := []testCase{
		{
			name: "Happy Case",
			desc: "All ok",
			jsonResp: `{
				"kind": "books#volumes",
				"totalItems": 1,
				"items": [
					{
						"kind": "books#volume",
						"id": "_iMqjwEACAAJ",
						"etag": "bM3T/1CmX50",
						"selfLink": "https://www.googleapis.com/books/v1/volumes/_iMqjwEACAAJ",
						"volumeInfo": {
							"title": "China Rich Girlfriend",
							"authors": [
								"Kevin Kwan"
							],
							"publisher": "Anchor Books",
							"publishedDate": "2016-05-31",
							"description": "It's the eve of her wedding to Nicholas Young, heir to one of the greatest fortunes in Asia, and Rachel Chu should be over the moon. She has a flawless Asscher-cut diamond from JAR, a wedding dress plucked from the best salon in Paris, and a fiancé willing to sacrifice his entire inheritance in order to marry her. But Rachel still mourns the fact that the father she never knew won't be there to walk her down the aisle ... until a shocking revelation draws Rachel into a world of Shanghai splendor beyond anything she has ever imagined. A romp through Asia's most exclusive clubs, auction houses, and estates, China Rich Girlfriend brings us into the elite circles of Mainland China and offers an inside glimpse at what it's like to be gloriously, crazily China rich.",
							"industryIdentifiers": [
								{
									"type": "ISBN_10",
									"identifier": "1101973390"
								},
								{
									"type": "ISBN_13",
									"identifier": "9781101973394"
								}
							],
							"readingModes": {
								"text": false,
								"image": false
							},
							"pageCount": 496,
							"printType": "BOOK",
							"categories": [
								"Fiancées"
							],
							"averageRating": 3,
							"ratingsCount": 32,
							"maturityRating": "NOT_MATURE",
							"allowAnonLogging": false,
							"contentVersion": "preview-1.0.0",
							"panelizationSummary": {
								"containsEpubBubbles": false,
								"containsImageBubbles": false
							},
							"imageLinks": {
								"smallThumbnail": "http://books.google.com/books/content?id=_iMqjwEACAAJ&printsec=frontcover&img=1&zoom=5&source=gbs_api",
								"thumbnail": "http://books.google.com/books/content?id=_iMqjwEACAAJ&printsec=frontcover&img=1&zoom=1&source=gbs_api"
							},
							"language": "en",
							"previewLink": "http://books.google.com.my/books?id=_iMqjwEACAAJ&dq=9781101973394&hl=&cd=1&source=gbs_api",
							"infoLink": "http://books.google.com.my/books?id=_iMqjwEACAAJ&dq=9781101973394&hl=&source=gbs_api",
							"canonicalVolumeLink": "https://books.google.com/books/about/China_Rich_Girlfriend.html?hl=&id=_iMqjwEACAAJ"
						},
						"saleInfo": {
							"country": "MY",
							"saleability": "NOT_FOR_SALE",
							"isEbook": false
						},
						"accessInfo": {
							"country": "MY",
							"viewability": "NO_PAGES",
							"embeddable": false,
							"publicDomain": false,
							"textToSpeechPermission": "ALLOWED",
							"epub": {
								"isAvailable": false
							},
							"pdf": {
								"isAvailable": false
							},
							"webReaderLink": "http://play.google.com/books/reader?id=_iMqjwEACAAJ&hl=&printsec=frontcover&source=gbs_api",
							"accessViewStatus": "NONE",
							"quoteSharingAllowed": false
						},
						"searchInfo": {
							"textSnippet": "A romp through Asia&#39;s most exclusive clubs, auction houses, and estates, China Rich Girlfriend brings us into the elite circles of Mainland China and offers an inside glimpse at what it&#39;s like to be gloriously, crazily China rich."
						}
					}
				]
			}`,
			expRes: &Book{
				Title:         "China Rich Girlfriend",
				PublishedYear: "2016-05-31",
				Authors:       []string{"Kevin Kwan"},
				Description:   "It's the eve of her wedding to Nicholas Young, heir to one of the greatest fortunes in Asia, and Rachel Chu should be over the moon. She has a flawless Asscher-cut diamond from JAR, a wedding dress plucked from the best salon in Paris, and a fiancé willing to sacrifice his entire inheritance in order to marry her. But Rachel still mourns the fact that the father she never knew won't be there to walk her down the aisle ... until a shocking revelation draws Rachel into a world of Shanghai splendor beyond anything she has ever imagined. A romp through Asia's most exclusive clubs, auction houses, and estates, China Rich Girlfriend brings us into the elite circles of Mainland China and offers an inside glimpse at what it's like to be gloriously, crazily China rich.",
				IndustryIdentifiers: &Identifier{
					ISBN:   "1101973390",
					ISBN13: "9781101973394",
				},
				PageCount:  496,
				Categories: []string{"Fiancées"},
				ImageLinks: &ImageLinks{
					SmallImageURL: "http://books.google.com/books/content?id=_iMqjwEACAAJ&printsec=frontcover&img=1&zoom=5&source=gbs_api",
					ImageURL:      "http://books.google.com/books/content?id=_iMqjwEACAAJ&printsec=frontcover&img=1&zoom=1&source=gbs_api",
				},
				Publisher: "Anchor Books",
				Language:  "en",
				Source:    "google",
			},
			respCode: 200,
		},
		{
			name: "Sad Case",
			desc: "return non 2xx response code",
			respCode: 999,
		},
		{
			name: "Sad Case",
			desc: "API return 0 items",
			jsonResp: `{
				"kind": "books#volumes",
				"totalItems": 0
			}`,
			respCode: 200,
		},
		{
			name: "Sad Case",
			desc: "API return random book",
			jsonResp: `{
				"kind": "books#volumes",
				"totalItems": 1,
				"items": [
					{
						"kind": "books#volume",
						"id": "_iMqjwEACAAJ",
						"etag": "bM3T/1CmX50",
						"selfLink": "https://www.googleapis.com/books/v1/volumes/_iMqjwEACAAJ",
						"volumeInfo": {
							"title": "China Rich Girlfriend",
							"authors": [
								"Kevin Kwan"
							],
							"publisher": "Anchor Books",
							"publishedDate": "2016-05-31",
							"description": "It's the eve of her wedding to Nicholas Young, heir to one of the greatest fortunes in Asia, and Rachel Chu should be over the moon. She has a flawless Asscher-cut diamond from JAR, a wedding dress plucked from the best salon in Paris, and a fiancé willing to sacrifice his entire inheritance in order to marry her. But Rachel still mourns the fact that the father she never knew won't be there to walk her down the aisle ... until a shocking revelation draws Rachel into a world of Shanghai splendor beyond anything she has ever imagined. A romp through Asia's most exclusive clubs, auction houses, and estates, China Rich Girlfriend brings us into the elite circles of Mainland China and offers an inside glimpse at what it's like to be gloriously, crazily China rich.",
							"industryIdentifiers": [
								{
									"type": "ISBN_10",
									"identifier": "1101973213213390"
								},
								{
									"type": "ISBN_13",
									"identifier": "97811431243101973394"
								}
							],
							"readingModes": {
								"text": false,
								"image": false
							},
							"pageCount": 496,
							"printType": "BOOK",
							"categories": [
								"Fiancées"
							],
							"averageRating": 3,
							"ratingsCount": 32,
							"maturityRating": "NOT_MATURE",
							"allowAnonLogging": false,
							"contentVersion": "preview-1.0.0",
							"panelizationSummary": {
								"containsEpubBubbles": false,
								"containsImageBubbles": false
							},
							"imageLinks": {
								"smallThumbnail": "http://books.google.com/books/content?id=_iMqjwEACAAJ&printsec=frontcover&img=1&zoom=5&source=gbs_api",
								"thumbnail": "http://books.google.com/books/content?id=_iMqjwEACAAJ&printsec=frontcover&img=1&zoom=1&source=gbs_api"
							},
							"language": "en",
							"previewLink": "http://books.google.com.my/books?id=_iMqjwEACAAJ&dq=9781101973394&hl=&cd=1&source=gbs_api",
							"infoLink": "http://books.google.com.my/books?id=_iMqjwEACAAJ&dq=9781101973394&hl=&source=gbs_api",
							"canonicalVolumeLink": "https://books.google.com/books/about/China_Rich_Girlfriend.html?hl=&id=_iMqjwEACAAJ"
						},
						"saleInfo": {
							"country": "MY",
							"saleability": "NOT_FOR_SALE",
							"isEbook": false
						},
						"accessInfo": {
							"country": "MY",
							"viewability": "NO_PAGES",
							"embeddable": false,
							"publicDomain": false,
							"textToSpeechPermission": "ALLOWED",
							"epub": {
								"isAvailable": false
							},
							"pdf": {
								"isAvailable": false
							},
							"webReaderLink": "http://play.google.com/books/reader?id=_iMqjwEACAAJ&hl=&printsec=frontcover&source=gbs_api",
							"accessViewStatus": "NONE",
							"quoteSharingAllowed": false
						},
						"searchInfo": {
							"textSnippet": "A romp through Asia&#39;s most exclusive clubs, auction houses, and estates, China Rich Girlfriend brings us into the elite circles of Mainland China and offers an inside glimpse at what it&#39;s like to be gloriously, crazily China rich."
						}
					}
				]
			}`,
			respCode: 200,
		},
		{
			name: "Sad Case",
			desc: "error decoding resp",
			jsonResp: `{
				"kind": "books#volumes",
				"totalItems": 1,
				"items": [
					{
						"kind": "books#volume",
						"id": "_iMqjwEACAAJ",
						"etag": "bM3T/1CmX50",
						"selfLink": "https://www.googleapis.com/books/v1/volumes/_iMqjwEACAAJ",
						"volumeInfo": {
							"title": "China Rich Girlfriend",
							"authors": [
								"Kevin Kwan"
							],
							"publisher": "Anchor Books",
							"publishedDate": "2016-05-31",
							"description": "It's the eve of her wedding to Nicholas Young, heir to one of the greatest fortunes in Asia, and Rachel Chu should be over the moon. She has a flawless Asscher-cut diamond from JAR, a wedding dress plucked from the best salon in Paris, and a fiancé willing to sacrifice his entire inheritance in order to marry her. But Rachel still mourns the fact that the father she never knew won't be there to walk her down the aisle ... until a shocking revelation draws Rachel into a world of Shanghai splendor beyond anything she has ever imagined. A romp through Asia's most exclusive clubs, auction houses, and estates, China Rich Girlfriend brings us into the elite circles of Mainland China and offers an inside glimpse at what it's like to be gloriously, crazily China rich.",
							"industryIdentifiers": [
								{
									"type": "ISBN_10",
									"identifier": "1101973390"
								},
								{
									"type": "ISBN_13",
									"identifier": "9781101973394"
								}
							],
							"readingModes": {
								"text": false,
								"image": false
							},
							"pageCount": 496.1234,
							"printType": "BOOK",
							"categories": [
								"Fiancées"
							],
							"averageRating": 3,
							"ratingsCount": 32,
							"maturityRating": "NOT_MATURE",
							"allowAnonLogging": false,
							"contentVersion": "preview-1.0.0",
							"panelizationSummary": {
								"containsEpubBubbles": false,
								"containsImageBubbles": false
							},
							"imageLinks": {
								"smallThumbnail": "http://books.google.com/books/content?id=_iMqjwEACAAJ&printsec=frontcover&img=1&zoom=5&source=gbs_api",
								"thumbnail": "http://books.google.com/books/content?id=_iMqjwEACAAJ&printsec=frontcover&img=1&zoom=1&source=gbs_api"
							},
							"language": "en",
							"previewLink": "http://books.google.com.my/books?id=_iMqjwEACAAJ&dq=9781101973394&hl=&cd=1&source=gbs_api",
							"infoLink": "http://books.google.com.my/books?id=_iMqjwEACAAJ&dq=9781101973394&hl=&source=gbs_api",
							"canonicalVolumeLink": "https://books.google.com/books/about/China_Rich_Girlfriend.html?hl=&id=_iMqjwEACAAJ"
						},
						"saleInfo": {
							"country": "MY",
							"saleability": "NOT_FOR_SALE",
							"isEbook": false
						},
						"accessInfo": {
							"country": "MY",
							"viewability": "NO_PAGES",
							"embeddable": false,
							"publicDomain": false,
							"textToSpeechPermission": "ALLOWED",
							"epub": {
								"isAvailable": false
							},
							"pdf": {
								"isAvailable": false
							},
							"webReaderLink": "http://play.google.com/books/reader?id=_iMqjwEACAAJ&hl=&printsec=frontcover&source=gbs_api",
							"accessViewStatus": "NONE",
							"quoteSharingAllowed": false
						},
						"searchInfo": {
							"textSnippet": "A romp through Asia&#39;s most exclusive clubs, auction houses, and estates, China Rich Girlfriend brings us into the elite circles of Mainland China and offers an inside glimpse at what it&#39;s like to be gloriously, crazily China rich."
						}
					}
				]
			}`,
			respCode: 200,
		},
		{
			name: "Sad Case",
			desc: "client returns error",
			err: fmt.Errorf("mock error"),
		},
	}

	gi := NewGoISBN([]string{ProviderGoogle})
	ch := make(chan *Book, 1)
	for _, v := range testCases {
		gi.client = &MockClient{
			MockDo: func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: v.respCode,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(v.jsonResp))),
				}, v.err
			},
		}
		gi.resolveGoogle("9781101973394", ch)
		actRes := <-ch
		assert.Equal(t, v.expRes, actRes)
	}
}

func TestResolveGoodreads(t *testing.T) {
	type TestCase struct {
		name     string
		desc     string
		isbn     string
		xmlResp  string
		expRes   *Book
		err      error
		respCode int
	}
	testCases := []TestCase{
		{
			name: "Happy Case",
			desc: "all ok, isbn13",
			isbn: "9780751562774",
			expRes: &Book{
				Title:         "The Secrets She Keeps",
				PublishedYear: "2017",
				Authors:       []string{"Michael Robotham"},
				ImageLinks: &ImageLinks{
					SmallImageURL: "https://s.gr-assets.com/assets/nophoto/book/50x75-a91bf249278a81aabab721ef782c4a74.png",
					ImageURL:      "https://s.gr-assets.com/assets/nophoto/book/111x148-bcc042a9c91a29c1d680899eff700a03.png",
				},
				IndustryIdentifiers: &Identifier{
					ISBN13: "9780751562774",
				},
				Source: "goodreads",
			},
			xmlResp: `<?xml version="1.0" encoding="UTF-8"?>
			<GoodreadsResponse>
				<Request>
					<authentication>true</authentication>
					<key>
						<![CDATA[6qVbqOjnzhHws97M5gYYA]]>
					</key>
					<method>
						<![CDATA[search_index]]>
					</method>
				</Request>
				<search>
					<query>
						<![CDATA[9780751562774]]>
					</query>
					<results-start>1</results-start>
					<results-end>1</results-end>
					<total-results>1</total-results>
					<source>Goodreads</source>
					<query-time-seconds>0.01</query-time-seconds>
					<results>
						<work>
							<id type="integer">54397694</id>
							<books_count type="integer">48</books_count>
							<ratings_count type="integer">29675</ratings_count>
							<text_reviews_count type="integer">3192</text_reviews_count>
							<original_publication_year type="integer">2017</original_publication_year>
							<original_publication_month type="integer">7</original_publication_month>
							<original_publication_day type="integer">11</original_publication_day>
							<average_rating>4.02</average_rating>
							<best_book type="Book">
								<id type="integer">36283464</id>
								<title>The Secrets She Keeps</title>
								<author>
									<id type="integer">266945</id>
									<name>Michael Robotham</name>
								</author>
								<image_url>https://s.gr-assets.com/assets/nophoto/book/111x148-bcc042a9c91a29c1d680899eff700a03.png</image_url>
								<small_image_url>https://s.gr-assets.com/assets/nophoto/book/50x75-a91bf249278a81aabab721ef782c4a74.png</small_image_url>
							</best_book>
						</work>
					</results>
				</search>
			</GoodreadsResponse>`,
			respCode: 200,
		},
		{
			name: "Happy Case",
			desc: "all ok, isbn10",
			isbn: "1982149000",
			expRes: &Book{
				Title:         "The Secrets She Keeps",
				PublishedYear: "2017",
				Authors:       []string{"Michael Robotham"},
				ImageLinks: &ImageLinks{
					SmallImageURL: "https://s.gr-assets.com/assets/nophoto/book/50x75-a91bf249278a81aabab721ef782c4a74.png",
					ImageURL:      "https://s.gr-assets.com/assets/nophoto/book/111x148-bcc042a9c91a29c1d680899eff700a03.png",
				},
				IndustryIdentifiers: &Identifier{
					ISBN: "1982149000",
				},
				Source: "goodreads",
			},
			xmlResp: `<?xml version="1.0" encoding="UTF-8"?>
			<GoodreadsResponse>
				<Request>
					<authentication>true</authentication>
					<key>
						<![CDATA[6qVbqOjnzhHws97M5gYYA]]>
					</key>
					<method>
						<![CDATA[search_index]]>
					</method>
				</Request>
				<search>
					<query>
						<![CDATA[9780751562774]]>
					</query>
					<results-start>1</results-start>
					<results-end>1</results-end>
					<total-results>1</total-results>
					<source>Goodreads</source>
					<query-time-seconds>0.01</query-time-seconds>
					<results>
						<work>
							<id type="integer">54397694</id>
							<books_count type="integer">48</books_count>
							<ratings_count type="integer">29675</ratings_count>
							<text_reviews_count type="integer">3192</text_reviews_count>
							<original_publication_year type="integer">2017</original_publication_year>
							<original_publication_month type="integer">7</original_publication_month>
							<original_publication_day type="integer">11</original_publication_day>
							<average_rating>4.02</average_rating>
							<best_book type="Book">
								<id type="integer">36283464</id>
								<title>The Secrets She Keeps</title>
								<author>
									<id type="integer">266945</id>
									<name>Michael Robotham</name>
								</author>
								<image_url>https://s.gr-assets.com/assets/nophoto/book/111x148-bcc042a9c91a29c1d680899eff700a03.png</image_url>
								<small_image_url>https://s.gr-assets.com/assets/nophoto/book/50x75-a91bf249278a81aabab721ef782c4a74.png</small_image_url>
							</best_book>
						</work>
					</results>
				</search>
			</GoodreadsResponse>`,
			respCode: 200,
		},
		{
			name: "Sad Case",
			desc: "client returns error",
			err:  fmt.Errorf("mock error"),
			isbn: "9780751562774",
		},
		{
			name:     "Sad Case",
			desc:     "client returns non 2XX response code",
			isbn:     "9780751562774",
			respCode: 999,
		},
		{
			name: "Sad Case",
			desc: "error decoding response",
			isbn: "9780751562774",
			xmlResp: `<?xml version="1.0" encoding="UTF-8"?>
			<GoodreadsResponse>
				<Request>
					<authentication>true</authentication>
					<key>
						<![CDATA[6qVbqOjnzhHws97M5gYYA]]>
					</key>
					<method>
						<![CDATA[search_index]]>
					</method>
				</Request>
				<search>
					<query>
						<![CDATA[9780751562774]]>
					</query>
					<results-start>1</results-start>
					<results-end>1</results-end>
					<total-results>1</total-results>
					<source>Goodreads</source>
					<query-time-seconds>0.01</query-time-seconds>
					<results>
						<work>
							<id type="integer">54397694</id>
							<books_count type="integer">48</books_count>
							<ratings_count type="integer">29675</ratings_count>
							<text_reviews_count type="integer">3192</text_reviews_count>
							<original_publication_year type="integer">2017</original_publication_year>
							<original_publication_month type="integer">7</original_publication_month>
							<original_publication_day type="integer">11</original_publication_day>
							<average_rating>4.02</average_rating>
							<best_book type="Book">
								<id type="integer">36283464</id>
								<title>The Secrets She Keeps</title>
								<author>
									<id type="integer">266945.12</id>
									<name>Michael Robotham</name>
								</author>
								<image_url>https://s.gr-assets.com/assets/nophoto/book/111x148-bcc042a9c91a29c1d680899eff700a03.png</image_url>
								<small_image_url>https://s.gr-assets.com/assets/nophoto/book/50x75-a91bf249278a81aabab721ef782c4a74.png</small_image_url>
							</best_book>
						</work>
					</results>
				</search>
			</GoodreadsResponse>`,
			respCode: 200,
		},
		{
			name: "Sad Case",
			desc: "API returns 0 item",
			isbn: "9780751562774",
			xmlResp: `<?xml version="1.0" encoding="UTF-8"?>
			<GoodreadsResponse>
				<Request>
					<authentication>true</authentication>
					<key>
						<![CDATA[6qVbqOjnzhHws97M5gYYA]]>
					</key>
					<method>
						<![CDATA[search_index]]>
					</method>
				</Request>
				<search>
					<query>
						<![CDATA[9780751562774]]>
					</query>
					<results-start>1</results-start>
					<results-end>1</results-end>
					<total-results>1</total-results>
					<source>Goodreads</source>
					<query-time-seconds>0.01</query-time-seconds>
					<results>
					</results>
				</search>
			</GoodreadsResponse>`,
			respCode: 200,
		},
	}
	defer unsetEnv()
	os.Setenv(goodreadsAPIKey, "mock goodread key")
	gi := NewGoISBN([]string{ProviderGoodreads})
	ch := make(chan *Book, 1)
	for _, v := range testCases {
		gi.client = &MockClient{
			MockDo: func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: v.respCode,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(v.xmlResp))),
				}, v.err
			},
		}
		gi.resolveGoodreads(v.isbn, ch)
		actRes := <-ch
		assert.Equal(t, v.expRes, actRes)
	}
}

func TestOpenLibrary(t *testing.T) {
	type testCase struct {
		name     string
		desc     string
		jsonResp string
		expRes   *Book
		err      error
		respCode int
		isbn     string
	}
	testCases := []testCase{
		{
			name: "Happy Case",
			desc: "all ok, isbn13",
			jsonResp: `
			{
				"ISBN:9780099588986": {
					"url": "https://openlibrary.org/books/OL32026810M/The_Confession",
					"key": "/books/OL32026810M",
					"title": "The Confession",
					"authors": [
						{
							"url": "https://openlibrary.org/authors/OL6829207A/John_Grisham",
							"name": "John Grisham"
						},
						{
							"url": "https://openlibrary.org/authors/OL9325475A/John_Grisham",
							"name": "John Grisham"
						}
					],
					"identifiers": {
						"isbn_13": [
							"9780099588986"
						],
						"openlibrary": [
							"OL32026810M"
						]
					},
					"publishers": [
						{
							"name": "Arrow Books"
						}
					],
					"publish_date": "2010",
					"subjects": [
						{
							"name": "nyt:mass_market_paperback=2011-07-23",
							"url": "https://openlibrary.org/subjects/nyt:mass_market_paperback=2011-07-23"
						},
						{
							"name": "New York Times bestseller",
							"url": "https://openlibrary.org/subjects/new_york_times_bestseller"
						},
						{
							"name": "death penalty",
							"url": "https://openlibrary.org/subjects/death_penalty"
						},
						{
							"name": "Judicial error",
							"url": "https://openlibrary.org/subjects/judicial_error"
						},
						{
							"name": "Fiction",
							"url": "https://openlibrary.org/subjects/fiction"
						},
						{
							"name": "Death row inmates",
							"url": "https://openlibrary.org/subjects/death_row_inmates"
						},
						{
							"name": "Executions and executioners",
							"url": "https://openlibrary.org/subjects/executions_and_executioners"
						},
						{
							"name": "Fiction, legal",
							"url": "https://openlibrary.org/subjects/fiction,_legal"
						},
						{
							"name": "Fiction, thrillers, general",
							"url": "https://openlibrary.org/subjects/fiction,_thrillers,_general"
						},
						{
							"name": "Large type books",
							"url": "https://openlibrary.org/subjects/large_type_books"
						},
						{
							"name": "Fiction, thrillers",
							"url": "https://openlibrary.org/subjects/fiction,_thrillers"
						},
						{
							"name": "Suspense fiction",
							"url": "https://openlibrary.org/subjects/suspense_fiction"
						},
						{
							"name": "Legal stories",
							"url": "https://openlibrary.org/subjects/legal_stories"
						},
						{
							"name": "Large print books",
							"url": "https://openlibrary.org/subjects/large_print_books"
						},
						{
							"name": "Roman policier",
							"url": "https://openlibrary.org/subjects/roman_policier"
						},
						{
							"name": "Erreur judiciaire",
							"url": "https://openlibrary.org/subjects/erreur_judiciaire"
						},
						{
							"name": "Romans, nouvelles",
							"url": "https://openlibrary.org/subjects/romans,_nouvelles"
						},
						{
							"name": "Condamnés à mort",
							"url": "https://openlibrary.org/subjects/condamnés_à_mort"
						}
					],
					"ebooks": [
						{
							"preview_url": "https://archive.org/details/confession0000gris_l3e5",
							"availability": "borrow",
							"formats": {},
							"borrow_url": "https://openlibrary.org/books/OL32026810M/The_Confession/borrow",
							"checkedout": true
						}
					],
					"cover": {
						"small": "https://covers.openlibrary.org/b/id/10693197-S.jpg",
						"medium": "https://covers.openlibrary.org/b/id/10693197-M.jpg",
						"large": "https://covers.openlibrary.org/b/id/10693197-L.jpg"
					}
				}
			}`,
			expRes: &Book{
				Title:         "The Confession",
				PublishedYear: "2010",
				Authors:       []string{"John Grisham", "John Grisham"},
				IndustryIdentifiers: &Identifier{
					ISBN13: "9780099588986",
				},
				ImageLinks: &ImageLinks{
					SmallImageURL: "https://covers.openlibrary.org/b/id/10693197-S.jpg",
					ImageURL:      "https://covers.openlibrary.org/b/id/10693197-M.jpg",
					LargeImageURL: "https://covers.openlibrary.org/b/id/10693197-L.jpg",
				},
				Publisher: "Arrow Books",
				Source:    "openlibrary",
			},
			respCode: 200,
			isbn:     "9780099588986",
		},
		{
			name: "Happy Case",
			desc: "all ok, isbn10",
			jsonResp: `
			{
				"ISBN:0099588986": {
					"url": "https://openlibrary.org/books/OL32026810M/The_Confession",
					"key": "/books/OL32026810M",
					"title": "The Confession",
					"authors": [
						{
							"url": "https://openlibrary.org/authors/OL6829207A/John_Grisham",
							"name": "John Grisham"
						},
						{
							"url": "https://openlibrary.org/authors/OL9325475A/John_Grisham",
							"name": "John Grisham"
						}
					],
					"identifiers": {
						"isbn_10": [
							"0099588986"
						],
						"openlibrary": [
							"OL32026810M"
						]
					},
					"publishers": [
						{
							"name": "Arrow Books"
						}
					],
					"publish_date": "2010",
					"subjects": [
						{
							"name": "nyt:mass_market_paperback=2011-07-23",
							"url": "https://openlibrary.org/subjects/nyt:mass_market_paperback=2011-07-23"
						},
						{
							"name": "New York Times bestseller",
							"url": "https://openlibrary.org/subjects/new_york_times_bestseller"
						},
						{
							"name": "death penalty",
							"url": "https://openlibrary.org/subjects/death_penalty"
						},
						{
							"name": "Judicial error",
							"url": "https://openlibrary.org/subjects/judicial_error"
						},
						{
							"name": "Fiction",
							"url": "https://openlibrary.org/subjects/fiction"
						},
						{
							"name": "Death row inmates",
							"url": "https://openlibrary.org/subjects/death_row_inmates"
						},
						{
							"name": "Executions and executioners",
							"url": "https://openlibrary.org/subjects/executions_and_executioners"
						},
						{
							"name": "Fiction, legal",
							"url": "https://openlibrary.org/subjects/fiction,_legal"
						},
						{
							"name": "Fiction, thrillers, general",
							"url": "https://openlibrary.org/subjects/fiction,_thrillers,_general"
						},
						{
							"name": "Large type books",
							"url": "https://openlibrary.org/subjects/large_type_books"
						},
						{
							"name": "Fiction, thrillers",
							"url": "https://openlibrary.org/subjects/fiction,_thrillers"
						},
						{
							"name": "Suspense fiction",
							"url": "https://openlibrary.org/subjects/suspense_fiction"
						},
						{
							"name": "Legal stories",
							"url": "https://openlibrary.org/subjects/legal_stories"
						},
						{
							"name": "Large print books",
							"url": "https://openlibrary.org/subjects/large_print_books"
						},
						{
							"name": "Roman policier",
							"url": "https://openlibrary.org/subjects/roman_policier"
						},
						{
							"name": "Erreur judiciaire",
							"url": "https://openlibrary.org/subjects/erreur_judiciaire"
						},
						{
							"name": "Romans, nouvelles",
							"url": "https://openlibrary.org/subjects/romans,_nouvelles"
						},
						{
							"name": "Condamnés à mort",
							"url": "https://openlibrary.org/subjects/condamnés_à_mort"
						}
					],
					"ebooks": [
						{
							"preview_url": "https://archive.org/details/confession0000gris_l3e5",
							"availability": "borrow",
							"formats": {},
							"borrow_url": "https://openlibrary.org/books/OL32026810M/The_Confession/borrow",
							"checkedout": true
						}
					],
					"cover": {
						"small": "https://covers.openlibrary.org/b/id/10693197-S.jpg",
						"medium": "https://covers.openlibrary.org/b/id/10693197-M.jpg",
						"large": "https://covers.openlibrary.org/b/id/10693197-L.jpg"
					}
				}
			}`,
			expRes: &Book{
				Title:         "The Confession",
				PublishedYear: "2010",
				Authors:       []string{"John Grisham", "John Grisham"},
				IndustryIdentifiers: &Identifier{
					ISBN: "0099588986",
				},
				ImageLinks: &ImageLinks{
					SmallImageURL: "https://covers.openlibrary.org/b/id/10693197-S.jpg",
					ImageURL:      "https://covers.openlibrary.org/b/id/10693197-M.jpg",
					LargeImageURL: "https://covers.openlibrary.org/b/id/10693197-L.jpg",
				},
				Publisher: "Arrow Books",
				Source:    "openlibrary",
			},
			respCode: 200,
			isbn:     "0099588986",
		},
		{
			name: "Sad Case",
			desc: "client returns error",
			err:  fmt.Errorf("mock error"),
			isbn: "9780099588986",
		},
		{
			name:     "Sad Case",
			desc:     "client returns non 2XX response code",
			respCode: 999,
			isbn:     "9780099588986",
		},
		{
			name: "Sad Case",
			desc: "error decoding response",
			jsonResp: `
			{
				"ISBN:9780099588986": {
					"url": "https://openlibrary.org/books/OL32026810M/The_Confession",
					"key": "/books/OL32026810M",
					"title": ["The Confession"],
					"authors": [
						{
							"url": "https://openlibrary.org/authors/OL6829207A/John_Grisham",
							"name": "John Grisham"
						},
						{
							"url": "https://openlibrary.org/authors/OL9325475A/John_Grisham",
							"name": "John Grisham"
						}
					],
					"identifiers": {
						"isbn_13": [
							"9780099588986"
						],
						"openlibrary": [
							"OL32026810M"
						]
					},
					"publishers": [
						{
							"name": "Arrow Books"
						}
					],
					"publish_date": "2010",
					"subjects": [
						{
							"name": "nyt:mass_market_paperback=2011-07-23",
							"url": "https://openlibrary.org/subjects/nyt:mass_market_paperback=2011-07-23"
						},
						{
							"name": "New York Times bestseller",
							"url": "https://openlibrary.org/subjects/new_york_times_bestseller"
						},
						{
							"name": "death penalty",
							"url": "https://openlibrary.org/subjects/death_penalty"
						},
						{
							"name": "Judicial error",
							"url": "https://openlibrary.org/subjects/judicial_error"
						},
						{
							"name": "Fiction",
							"url": "https://openlibrary.org/subjects/fiction"
						},
						{
							"name": "Death row inmates",
							"url": "https://openlibrary.org/subjects/death_row_inmates"
						},
						{
							"name": "Executions and executioners",
							"url": "https://openlibrary.org/subjects/executions_and_executioners"
						},
						{
							"name": "Fiction, legal",
							"url": "https://openlibrary.org/subjects/fiction,_legal"
						},
						{
							"name": "Fiction, thrillers, general",
							"url": "https://openlibrary.org/subjects/fiction,_thrillers,_general"
						},
						{
							"name": "Large type books",
							"url": "https://openlibrary.org/subjects/large_type_books"
						},
						{
							"name": "Fiction, thrillers",
							"url": "https://openlibrary.org/subjects/fiction,_thrillers"
						},
						{
							"name": "Suspense fiction",
							"url": "https://openlibrary.org/subjects/suspense_fiction"
						},
						{
							"name": "Legal stories",
							"url": "https://openlibrary.org/subjects/legal_stories"
						},
						{
							"name": "Large print books",
							"url": "https://openlibrary.org/subjects/large_print_books"
						},
						{
							"name": "Roman policier",
							"url": "https://openlibrary.org/subjects/roman_policier"
						},
						{
							"name": "Erreur judiciaire",
							"url": "https://openlibrary.org/subjects/erreur_judiciaire"
						},
						{
							"name": "Romans, nouvelles",
							"url": "https://openlibrary.org/subjects/romans,_nouvelles"
						},
						{
							"name": "Condamnés à mort",
							"url": "https://openlibrary.org/subjects/condamnés_à_mort"
						}
					],
					"ebooks": [
						{
							"preview_url": "https://archive.org/details/confession0000gris_l3e5",
							"availability": "borrow",
							"formats": {},
							"borrow_url": "https://openlibrary.org/books/OL32026810M/The_Confession/borrow",
							"checkedout": true
						}
					],
					"cover": {
						"small": "https://covers.openlibrary.org/b/id/10693197-S.jpg",
						"medium": "https://covers.openlibrary.org/b/id/10693197-M.jpg",
						"large": "https://covers.openlibrary.org/b/id/10693197-L.jpg"
					}
				}
			}`,
			respCode: 200,
			isbn:     "9780099588986",
		},
		{
			name: "Sad Case",
			desc: "API return incorrect book",
			jsonResp: `
			{
				"ISBN:9780099588986321": {
					"url": "https://openlibrary.org/books/OL32026810M/The_Confession",
					"key": "/books/OL32026810M",
					"title": "The Confession",
					"authors": [
						{
							"url": "https://openlibrary.org/authors/OL6829207A/John_Grisham",
							"name": "John Grisham"
						},
						{
							"url": "https://openlibrary.org/authors/OL9325475A/John_Grisham",
							"name": "John Grisham"
						}
					],
					"identifiers": {
						"isbn_13": [
							"9780099588986"
						],
						"openlibrary": [
							"OL32026810M"
						]
					},
					"publishers": [
						{
							"name": "Arrow Books"
						}
					],
					"publish_date": "2010",
					"subjects": [
						{
							"name": "nyt:mass_market_paperback=2011-07-23",
							"url": "https://openlibrary.org/subjects/nyt:mass_market_paperback=2011-07-23"
						},
						{
							"name": "New York Times bestseller",
							"url": "https://openlibrary.org/subjects/new_york_times_bestseller"
						},
						{
							"name": "death penalty",
							"url": "https://openlibrary.org/subjects/death_penalty"
						},
						{
							"name": "Judicial error",
							"url": "https://openlibrary.org/subjects/judicial_error"
						},
						{
							"name": "Fiction",
							"url": "https://openlibrary.org/subjects/fiction"
						},
						{
							"name": "Death row inmates",
							"url": "https://openlibrary.org/subjects/death_row_inmates"
						},
						{
							"name": "Executions and executioners",
							"url": "https://openlibrary.org/subjects/executions_and_executioners"
						},
						{
							"name": "Fiction, legal",
							"url": "https://openlibrary.org/subjects/fiction,_legal"
						},
						{
							"name": "Fiction, thrillers, general",
							"url": "https://openlibrary.org/subjects/fiction,_thrillers,_general"
						},
						{
							"name": "Large type books",
							"url": "https://openlibrary.org/subjects/large_type_books"
						},
						{
							"name": "Fiction, thrillers",
							"url": "https://openlibrary.org/subjects/fiction,_thrillers"
						},
						{
							"name": "Suspense fiction",
							"url": "https://openlibrary.org/subjects/suspense_fiction"
						},
						{
							"name": "Legal stories",
							"url": "https://openlibrary.org/subjects/legal_stories"
						},
						{
							"name": "Large print books",
							"url": "https://openlibrary.org/subjects/large_print_books"
						},
						{
							"name": "Roman policier",
							"url": "https://openlibrary.org/subjects/roman_policier"
						},
						{
							"name": "Erreur judiciaire",
							"url": "https://openlibrary.org/subjects/erreur_judiciaire"
						},
						{
							"name": "Romans, nouvelles",
							"url": "https://openlibrary.org/subjects/romans,_nouvelles"
						},
						{
							"name": "Condamnés à mort",
							"url": "https://openlibrary.org/subjects/condamnés_à_mort"
						}
					],
					"ebooks": [
						{
							"preview_url": "https://archive.org/details/confession0000gris_l3e5",
							"availability": "borrow",
							"formats": {},
							"borrow_url": "https://openlibrary.org/books/OL32026810M/The_Confession/borrow",
							"checkedout": true
						}
					],
					"cover": {
						"small": "https://covers.openlibrary.org/b/id/10693197-S.jpg",
						"medium": "https://covers.openlibrary.org/b/id/10693197-M.jpg",
						"large": "https://covers.openlibrary.org/b/id/10693197-L.jpg"
					}
				}
			}`,
			respCode: 200,
			isbn:     "9780099588986",
		},
	}
	gi := NewGoISBN([]string{ProviderOpenLibrary})
	ch := make(chan *Book, 1)
	for _, v := range testCases {
		gi.client = &MockClient{
			MockDo: func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: v.respCode,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(v.jsonResp))),
				}, v.err
			},
		}
		gi.resolveOpenLibrary(v.isbn, ch)
		actRes := <-ch
		assert.Equal(t, v.expRes, actRes)
	}
}

func TestResolveIsbnDB(t *testing.T) {
	type testCase struct {
		name     string
		desc     string
		isbn     string
		jsonResp string
		err      error
		expRes   *Book
		respCode int
	}

	testCases := []testCase{
		{
			name: "Happy Case",
			desc: "all ok",
			isbn: "9781407243207",
			jsonResp: `{
				"book": {
					"publisher": "",
					"language": "en_US",
					"image": "https://images.isbndb.com/covers/32/07/9781407243207.jpg",
					"title_long": "The Bourne Ultimatum",
					"dimensions": "Height: 1.5748 Inches, Length: 7.874 Inches, Weight: 0.83775654089442 Pounds, Width: 5.5118 Inches",
					"date_published": "",
					"authors": [],
					"title": "The Bourne Ultimatum",
					"isbn13": "9781407243207",
					"msrp": "0.00",
					"binding": "Paperback",
					"isbn": "1407243209"
				}
			}`,
			expRes: &Book{
				Title:   "The Bourne Ultimatum",
				Authors: []string{},
				IndustryIdentifiers: &Identifier{
					ISBN:   "1407243209",
					ISBN13: "9781407243207",
				},
				ImageLinks: &ImageLinks{
					SmallImageURL: "https://images.isbndb.com/covers/32/07/9781407243207.jpg",
				},
				Language: "en_US",
				Source:   "isbndb",
			},
			respCode: 200,
		},
		{
			name: "Sad Case",
			desc: "client returns error",
			err:  fmt.Errorf("mock error"),
			isbn: "9781407243207",
		},
		{
			name:     "Sad Case",
			desc:     "client returns non 2XX response code",
			isbn:     "9781407243207",
			respCode: 999,
		},
		{
			name: "Sad Case",
			desc: "error decoding response",
			isbn: "9781407243207",
			jsonResp: `{
				"book": {
					"publisher": "",
					"language": "en_US",
					"image": "https://images.isbndb.com/covers/32/07/9781407243207.jpg",
					"title_long": "The Bourne Ultimatum",
					"dimensions": "Height: 1.5748 Inches, Length: 7.874 Inches, Weight: 0.83775654089442 Pounds, Width: 5.5118 Inches",
					"date_published": 12345,
					"authors": [],
					"title": "The Bourne Ultimatum",
					"isbn13": "9781407243207",
					"msrp": "0.00",
					"binding": "Paperback",
					"isbn": "1407243209"
				}
			}`,
			respCode: 200,
		},
		{
			name: "Sad Case",
			desc: "API returns random book",
			isbn: "9781407243207",
			jsonResp: `{
				"book": {
					"publisher": "",
					"language": "en_US",
					"image": "https://images.isbndb.com/covers/32/07/9781407243207.jpg",
					"title_long": "The Bourne Ultimatum",
					"dimensions": "Height: 1.5748 Inches, Length: 7.874 Inches, Weight: 0.83775654089442 Pounds, Width: 5.5118 Inches",
					"date_published": "",
					"authors": [],
					"title": "The Bourne Ultimatum",
					"isbn13": "12345",
					"msrp": "0.00",
					"binding": "Paperback",
					"isbn": "5431"
				}
			}`,
			respCode: 200,
		},
	}

	defer unsetEnv()
	os.Setenv(isbndbAPIKey, "mock isbndb key")
	gi := NewGoISBN([]string{ProviderIsbndb})
	ch := make(chan *Book, 1)

	for _, v := range testCases {
		gi.client = &MockClient{
			MockDo: func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: v.respCode,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(v.jsonResp))),
				}, v.err
			},
		}
		gi.resolveISBNDB(v.isbn, ch)
		actRes := <-ch
		assert.Equal(t, v.expRes, actRes)

	}
}

func TestGet(t *testing.T) {
	type testCase struct {
		name        string
		desc        string
		isbn        string
		apiRespCode int
		apiResp     string
		expRes      *Book
		expErr      error
	}
	testCases := []testCase{
		{
			name: "Happy Case",
			desc: "all ok",
			isbn: "9781101973394",
			apiResp: `{
					"kind": "books#volumes",
					"totalItems": 1,
					"items": [
						{
							"kind": "books#volume",
							"id": "_iMqjwEACAAJ",
							"etag": "bM3T/1CmX50",
							"selfLink": "https://www.googleapis.com/books/v1/volumes/_iMqjwEACAAJ",
							"volumeInfo": {
								"title": "China Rich Girlfriend",
								"authors": [
									"Kevin Kwan"
								],
								"publisher": "Anchor Books",
								"publishedDate": "2016-05-31",
								"description": "It's the eve of her wedding to Nicholas Young, heir to one of the greatest fortunes in Asia, and Rachel Chu should be over the moon. She has a flawless Asscher-cut diamond from JAR, a wedding dress plucked from the best salon in Paris, and a fiancé willing to sacrifice his entire inheritance in order to marry her. But Rachel still mourns the fact that the father she never knew won't be there to walk her down the aisle ... until a shocking revelation draws Rachel into a world of Shanghai splendor beyond anything she has ever imagined. A romp through Asia's most exclusive clubs, auction houses, and estates, China Rich Girlfriend brings us into the elite circles of Mainland China and offers an inside glimpse at what it's like to be gloriously, crazily China rich.",
								"industryIdentifiers": [
									{
										"type": "ISBN_10",
										"identifier": "1101973390"
									},
									{
										"type": "ISBN_13",
										"identifier": "9781101973394"
									}
								],
								"readingModes": {
									"text": false,
									"image": false
								},
								"pageCount": 496,
								"printType": "BOOK",
								"categories": [
									"Fiancées"
								],
								"averageRating": 3,
								"ratingsCount": 32,
								"maturityRating": "NOT_MATURE",
								"allowAnonLogging": false,
								"contentVersion": "preview-1.0.0",
								"panelizationSummary": {
									"containsEpubBubbles": false,
									"containsImageBubbles": false
								},
								"imageLinks": {
									"smallThumbnail": "http://books.google.com/books/content?id=_iMqjwEACAAJ&printsec=frontcover&img=1&zoom=5&source=gbs_api",
									"thumbnail": "http://books.google.com/books/content?id=_iMqjwEACAAJ&printsec=frontcover&img=1&zoom=1&source=gbs_api"
								},
								"language": "en",
								"previewLink": "http://books.google.com.my/books?id=_iMqjwEACAAJ&dq=9781101973394&hl=&cd=1&source=gbs_api",
								"infoLink": "http://books.google.com.my/books?id=_iMqjwEACAAJ&dq=9781101973394&hl=&source=gbs_api",
								"canonicalVolumeLink": "https://books.google.com/books/about/China_Rich_Girlfriend.html?hl=&id=_iMqjwEACAAJ"
							},
							"saleInfo": {
								"country": "MY",
								"saleability": "NOT_FOR_SALE",
								"isEbook": false
							},
							"accessInfo": {
								"country": "MY",
								"viewability": "NO_PAGES",
								"embeddable": false,
								"publicDomain": false,
								"textToSpeechPermission": "ALLOWED",
								"epub": {
									"isAvailable": false
								},
								"pdf": {
									"isAvailable": false
								},
								"webReaderLink": "http://play.google.com/books/reader?id=_iMqjwEACAAJ&hl=&printsec=frontcover&source=gbs_api",
								"accessViewStatus": "NONE",
								"quoteSharingAllowed": false
							},
							"searchInfo": {
								"textSnippet": "A romp through Asia&#39;s most exclusive clubs, auction houses, and estates, China Rich Girlfriend brings us into the elite circles of Mainland China and offers an inside glimpse at what it&#39;s like to be gloriously, crazily China rich."
							}
						}
					]
				}`,
			expRes: &Book{
				Title:         "China Rich Girlfriend",
				PublishedYear: "2016-05-31",
				Authors:       []string{"Kevin Kwan"},
				Description:   "It's the eve of her wedding to Nicholas Young, heir to one of the greatest fortunes in Asia, and Rachel Chu should be over the moon. She has a flawless Asscher-cut diamond from JAR, a wedding dress plucked from the best salon in Paris, and a fiancé willing to sacrifice his entire inheritance in order to marry her. But Rachel still mourns the fact that the father she never knew won't be there to walk her down the aisle ... until a shocking revelation draws Rachel into a world of Shanghai splendor beyond anything she has ever imagined. A romp through Asia's most exclusive clubs, auction houses, and estates, China Rich Girlfriend brings us into the elite circles of Mainland China and offers an inside glimpse at what it's like to be gloriously, crazily China rich.",
				IndustryIdentifiers: &Identifier{
					ISBN:   "1101973390",
					ISBN13: "9781101973394",
				},
				PageCount:  496,
				Categories: []string{"Fiancées"},
				ImageLinks: &ImageLinks{
					SmallImageURL: "http://books.google.com/books/content?id=_iMqjwEACAAJ&printsec=frontcover&img=1&zoom=5&source=gbs_api",
					ImageURL:      "http://books.google.com/books/content?id=_iMqjwEACAAJ&printsec=frontcover&img=1&zoom=1&source=gbs_api",
				},
				Publisher: "Anchor Books",
				Language:  "en",
				Source:    "google",
			},
			apiRespCode: 200,
		},
		{
			name: "Sad Case",
			desc: "providers return error",
			isbn: "9781101973394",
			expRes: nil,
			expErr: errBookNotFound,
		},
		{
			name: "Sad Case",
			desc: "invalid isbn error",
			isbn: "9781101973x94",
			expRes: nil,
			expErr: errInvalidISBN,
		},
	}
	gi := NewGoISBN(DEFAULT_PROVIDERS)
	for _, v := range testCases {
		gi.client = &MockClient{
			MockDo: func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: v.apiRespCode,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(v.apiResp))),
				}, nil
			},
		}
		actRes, actErr := gi.Get(v.isbn)

		assert.Equal(t, v.expRes, actRes)
		assert.Equal(t, v.expErr, actErr)

	}
}

func unsetEnv() (restore func()) {
	before := map[string]string{
		goodreadsAPIKey: os.Getenv(goodreadsAPIKey),
		isbndbAPIKey:    os.Getenv(isbndbAPIKey),
	}
	for k := range before {
		os.Unsetenv(k)
	}
	return func() {
		for k, v := range before {
			if os.Getenv(k) != v {
				os.Setenv(k, v)
			}
		}
	}
}
