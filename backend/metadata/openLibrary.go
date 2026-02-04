package metadata

type OpenLibrarySearchResults struct {
	Start            int         `json:"start"`
	NumFoundExact    bool        `json:"numFoundExact"`
	NumFound         int         `json:"num_found"`
	DocumentationURL string      `json:"documentation_url"`
	Q                string      `json:"q"`
	Offset           interface{} `json:"offset"`
	Docs             []struct {
		AuthorKey          []string `json:"author_key,omitempty"`
		AuthorName         []string `json:"author_name,omitempty"`
		CoverEditionKey    string   `json:"cover_edition_key,omitempty"`
		CoverI             int      `json:"cover_i,omitempty"`
		EbookAccess        string   `json:"ebook_access"`
		EditionCount       int      `json:"edition_count"`
		FirstPublishYear   int      `json:"first_publish_year,omitempty"`
		HasFulltext        bool     `json:"has_fulltext"`
		Ia                 []string `json:"ia,omitempty"`
		IaCollection       []string `json:"ia_collection,omitempty"`
		Key                string   `json:"key"`
		Language           []string `json:"language,omitempty"`
		LendingEditionS    string   `json:"lending_edition_s,omitempty"`
		LendingIdentifierS string   `json:"lending_identifier_s,omitempty"`
		PublicScanB        bool     `json:"public_scan_b"`
		Title              string   `json:"title"`
		Subtitle           string   `json:"subtitle,omitempty"`
	} `json:"docs"`
}

type OpenLibrarySearchParams struct {
	Title     *string `json:"title"`
	Author    *string `json:"author"`
	Year      *int    `json:"year"`
	ISBN      *string `json:"isbn"`
	ASIN      *string `json:"asin"`
	Publisher *string `json:"publisher"`
}

func SearchOpenLibrary()
