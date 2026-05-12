
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

export type AdvancedSearchTerms = {
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
}

export const Empty = {
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
    languages: ''
}

export function AdvancedTermsAreEmpty(terms: AdvancedSearchTerms): boolean {
    return terms.authors === '' && terms.genres === '' && terms.narrators === '' && terms.series === '' && terms.keywords === '' && terms.tags === ''
}
export function AddAdvancedTermsToQuery(terms: AdvancedSearchTerms, params: URLSearchParams): URLSearchParams {

    if (terms.year.trim()) params.append('publish_year', terms.year.trim());
    if (terms.isbn.trim()) params.append('isbn', terms.isbn.trim());
    if (terms.asin.trim()) params.append('asin', terms.asin.trim());
    if (terms.publisher.trim()) params.append('publisher', terms.publisher.trim());
    if (terms.series.trim()) params.append('series', terms.series.trim());
    if (terms.tags.trim()) params.append('tags', terms.tags.trim());
    if (terms.genres.trim()) params.append('genres', terms.genres.trim());
    if (terms.authors.trim()) params.append('authors', terms.authors.trim());
    if (terms.narrators.trim()) params.append('narrators', terms.narrators.trim());
    if (terms.languages.trim()) params.append('languages', terms.languages.trim());

    return params
}

export function TrimAdvancedTerms(terms: AdvancedSearchTerms): AdvancedSearchTerms {
    return {
        authors: terms.authors.trim(),
        genres: terms.genres.trim(),
        narrators: terms.narrators.trim(),
        series: terms.series.trim(),
        keywords: terms.keywords.trim(),
        tags: terms.tags.trim(),
        year: terms.year.trim(),
        publisher: terms.publisher.trim(),
        isbn: terms.isbn.trim(),
        asin: terms.asin.trim(),
        languages: terms.languages.trim()
    }
}