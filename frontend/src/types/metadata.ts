import { AddSearchTermsToQuery, HasSearchAdvancedTerms, type AdvancedSearchFields, type SearchTerms } from "@/types/search"
import type { BookParams } from "./book"
import { useNotificationsStore } from "@/stores/notifications"

const pageLimit = 10

export enum MetadataSource {
  OpenLibrary = "Open Library",
  GoogleBooks = "Google Books",
  Audible = "Audible"
}

// "au", "ca", "de", "es", "fr", "in", "it", "jp", "us", "uk"
export enum AudibleRegion {
  US = "com",
  CA = "ca",
  DE = "de",
  ES = "es",
  FR = "fr",
  IN = "co.in",
  IT = "it",
  JP = "co.jp",
  UK = "co.uk",
  AU = "co.au"
}

export const metadataSearchFields = new Map<MetadataSource, AdvancedSearchFields>()
metadataSearchFields.set(MetadataSource.OpenLibrary, {
  authors: true,
  narrators: false,
  tags: false,
  year: true,
  publisher: true,
  isbn: true,
  genres: true,
  languages: true,
  asin: false,
  keywords: false,
  series: false
})
metadataSearchFields.set(MetadataSource.GoogleBooks, {
  authors: true,
  narrators: false,
  tags: false,
  year: true,
  publisher: true,
  isbn: true,
  genres: true,
  languages: true,
  asin: false,
  keywords: false,
  series: false
})
metadataSearchFields.set(MetadataSource.Audible, {
  authors: true,
  narrators: false,
  tags: false,
  year: false,
  publisher: true,
  isbn: false,
  genres: false,
  languages: false,
  asin: true,
  keywords: true,
  series: true
})

export type MetadataSearchResults = {
  items: BookParams[]
  total_count: number
  count: number
  offset: number
  error: string
}

function buildQueryParams(params: SearchTerms, pageLimit: number, page: number): URLSearchParams {
  const urlParams = new URLSearchParams()
  if (params.metadataSource) {
    urlParams.append('source', params.metadataSource.toLowerCase())
    if (params.audibleRegion && params.metadataSource === MetadataSource.Audible) urlParams.append('region', params.audibleRegion)
  }

  if (params.search?.trim()) urlParams.set('title', params.search)
  if (params.authors?.trim()) urlParams.set('author', params.authors)
  if (params.year?.trim()) urlParams.set('year', params.year)
  if (params.publisher?.trim()) urlParams.set('publisher', params.publisher)
  if (params.isbn?.trim()) urlParams.set('isbn', params.isbn)
  if (params.asin?.trim()) urlParams.set('asin', params.asin)
  if (params.keywords?.trim()) urlParams.set('keyword', params.keywords)
  if (params.genres?.trim()) params.genres.split(',').forEach(g => urlParams.append('genre', g.trim()))
  if (params.languages?.trim()) params.languages.split(',').forEach(l => urlParams.append('language', l.trim()))

  urlParams.append('limit', pageLimit?.toString())
  urlParams.append('page', page.toString())

  return urlParams
}

export async function searchMetadataSource(params: SearchTerms, pageLimit: number, page: number): Promise<MetadataSearchResults | null> {
  // if (!HasSearchAdvancedTerms(params) || !params.search) {
  //   useNotificationsStore().notifyError('Enter at least one search term.')
  //   return null
  // }

  const endpoint = '/api/metadata/'
  const queryParams = buildQueryParams(params, pageLimit, page)

  try {
    const resp = await fetch(`${endpoint}?${queryParams.toString()}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
    })

    if (!resp.ok) {
      const body = await resp.text()
      throw new Error(`${resp.status} ${resp.statusText}: ${body}`)
    }

    const body = (await resp.json()) as MetadataSearchResults
    return body

  } catch (err) {
    console.error(err)
    useNotificationsStore().notifyError(`Something went wrong`)
    return null
  }
}

export async function getMetadataDetails(item: BookParams | null): Promise<BookParams | null> {
  if (!item) return null
  if (!item.key) {
    useNotificationsStore().notifyError('Selected item does not have a valid key for fetching details.')
    return null
  }

  try {
    const resp = await fetch(item.key, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
    })

    if (!resp.ok) {
      const body = await resp.text()
      throw new Error(`${resp.status} ${resp.statusText}: ${body}`)
    }

    var details: BookParams = (await resp.json()) as BookParams

    details.title = details.title ?? item.title
    details.subtitle = details.subtitle ?? item.subtitle
    details.authors = details.authors ?? item.authors
    details.genres = details.genres ?? item.genres
    details.series = details.series ?? item.series
    details.year = details.year ?? item.year
    details.publisher = details.publisher ?? item.publisher
    details.isbn = details.isbn ?? item.isbn
    details.cover = details.cover ?? item.cover

    return details

  } catch (err) {
    console.error('Get book details error', err)
    useNotificationsStore().notifyError('Failed to get book details: ' + (err instanceof Error ? err.message : String(err)))
    return null
  }
}