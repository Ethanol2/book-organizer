import type { AudibleRegion, MetadataSource } from "./metadata"

export type AdvancedSearchFields = {
    authors: boolean
    genres: boolean
    narrators: boolean
    series: boolean
    keywords: boolean
    tags: boolean

    year: boolean
    publisher: boolean

    isbn: boolean
    asin: boolean
    languages: boolean
}

export const librarySearchFields: AdvancedSearchFields = {
    year: true, isbn: true, asin: true, publisher: true, series: true, tags: true, genres: true, authors: true, narrators: true, keywords: false, languages: false
}

export type SearchTerms = {
    // Basic Search Terms
    search: string

    // Advanced Search Terms
    authors: string
    genres: string
    narrators: string
    series: string,
    keywords: string
    tags: string
    languages: string

    year: string
    publisher: string

    isbn: string
    asin: string

    files: string

    // Sorting
    sort: string
    order: string 

    // Metadata
    metadataSource: MetadataSource | null
    audibleRegion: AudibleRegion | null
}
export const NewSearchTerms = (): SearchTerms =>({
    search: '',
    authors: '',
    genres: '',
    narrators: '',
    series: '',
    keywords: '',
    tags: '',
    year: '',
    publisher: '',
    isbn: '',
    asin: '',
    languages: '',
    files: '',
    sort: '',
    order: '',
    metadataSource: null,
    audibleRegion: null
});
export function AddSearchTermsToQuery(terms: SearchTerms, params: URLSearchParams): URLSearchParams {

    terms = TrimSearchTerms(terms);

    if (terms.search) params.append('search', terms.search);
    if (terms.year) params.append('publish_year', terms.year);
    if (terms.isbn) params.append('isbn', terms.isbn);
    if (terms.asin) params.append('asin', terms.asin);
    if (terms.publisher) params.append('publisher', terms.publisher);
    if (terms.series) params.append('series', terms.series);
    if (terms.tags) params.append('tags', terms.tags);
    if (terms.genres) params.append('genres', terms.genres);
    if (terms.authors) params.append('authors', terms.authors);
    if (terms.narrators) params.append('narrators', terms.narrators);
    if (terms.languages) params.append('languages', terms.languages);
    if (terms.keywords) params.append('keywords', terms.keywords);

    if (terms.sort) params.append('sortBy', terms.sort);
    if (terms.order) params.append('sortOrder', terms.order);
    if (terms.files) params.append('files', terms.files);

    return params
}

export function TrimSearchTerms(terms: SearchTerms): SearchTerms {
    if (!terms) return NewSearchTerms();
    
    terms.search = terms.search?.trim();
    terms.authors = terms.authors?.trim();
    terms.genres = terms.genres?.trim();
    terms.narrators = terms.narrators?.trim();
    terms.series = terms.series?.trim();
    terms.keywords = terms.keywords?.trim();
    terms.tags = terms.tags?.trim();
    terms.languages = terms.languages?.trim();
    terms.isbn = terms.isbn?.trim();
    terms.asin = terms.asin?.trim();
    return terms
}

export function HasSearchAdvancedTerms(terms: SearchTerms): boolean {
    terms = TrimSearchTerms(terms);
    return terms.authors !== '' ||
        terms.genres !== '' ||
        terms.narrators !== '' ||
        terms.series !== '' ||
        terms.keywords !== '' ||
        terms.tags !== '' ||
        terms.year !== '' ||
        terms.publisher !== '' ||
        terms.isbn !== '' ||
        terms.asin !== ''
}