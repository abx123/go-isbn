package goisbn

type googleBooksResponse struct {
	TotalItems int64 `json:"totalItems,omitempty"`
	Items      []struct {
		VolumeInfo struct {
			Title           string   `json:"title,omitempty"`
			Authors         []string `json:"authors,omitempty"`
			PublicationYear string   `json:"publishedDate,omitempty"`
			AverageRating   float64  `json:"averageRating,omitempty"`
			Identifier      []struct {
				Type       string `json:"type,omitempty"`
				Identifier string `json:"identifier,omitempty"`
			} `json:"industryIdentifiers,omitempty"`
			Image struct {
				ImageURL      string `json:"thumbnail,omitempty"`
				SmallImageURL string `json:"smallThumbnail,omitempty"`
			} `json:"imageLinks,omitempty"`
		} `json:"volumeInfo,omitempty"`
	} `json:"items,omitempty"`
}

type goodreadsResponse struct {
	Search struct {
		Results struct {
			Work struct {
				PublicationYear int64   `xml:"original_publication_year"`
				AverageRating   float64 `xml:"average_rating"`
				Book            struct {
					ID     int64  `xml:"id"`
					Title  string `xml:"title"`
					Author struct {
						ID   int64  `xml:"id"`
						Name string `xml:"name"`
					} `xml:"author"`
					ImageURL      string `xml:"image_url"`
					SmallImageURL string `xml:"small_image_url"`
				} `xml:"best_book"`
			} `xml:"work"`
		} `xml:"results"`
	} `xml:"search"`
}

type openLibraryresponse struct {
	Title       string `json:"title,omitempty"`
	Identifiers struct {
		ISBN   []string `json:"isbn_10,omitempty"`
		ISBN13 []string `json:"isbn_13,omitempty"`
	} `json:"identifiers,omitempty"`
	Authors []struct {
		Name string `json:"name,omitempty"`
	} `json:"authors,omitempty"`
	PublishedYear string `json:"publish_date,omitempty"`
	Publishers    []struct {
		Name string `json:"name,omitempty"`
	} `json:"publishers,omitempty"`
	Cover struct {
		Small  string `json:"small,omitempty"`
		Medium string `json:"medium,omitempty"`
		Large  string `json:"large,omitempty"`
	} `json:"cover,omitempty"`
}

type Book struct {
	Title               string     `json:"title"`
	PublishedYear       string     `json:"published_year"`
	Authors             []string   `json:"authors"`
	Description         string     `json:"description"`
	IndustryIdentifiers identifier `json:"industry_identifiers"`
	PageCount           int64      `json:"page_count"`
	// PrintType           string     `json:"print_type"`
	Categories []string `json:"categories"`
	ImageLinks string   `json:"image_links"`
	// PreviewLink string   `json:"preview_link"`
	// InfoLink    string   `json:"info_link"`
	Publisher string `json:"publisher"`
	Language  string `json:"language"`
	Source    string `json:"source"`
}

type identifier struct {
	ISBN   string `json:"isbn"`
	ISBN13 string `json:"isbn_13"`
}
