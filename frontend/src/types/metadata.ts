import type { BookParams } from "./book"
import { useNotificationsStore } from "@/stores/notifications"

const pageLimit = 10

export enum MetadataType {
  OpenLibrary = "Open Library",
  GoogleBooks = "Google Books",
}

export type MetadataSearchParams = {
    source: MetadataType
    pageLimit: number | null
    page: number
    
    title?: string
    author?: string
    year?: string
    publisher?: string
    isbn?: string
    asin?: string
    genres?: string
    languages?: string
}
export type MetadataSearchResults = {
    items: BookParams[]
    total_count: number
    count: number
    offset: number
    error: string
}

function hasParams(params: MetadataSearchParams) {
  return params.title?.trim()
  || params.author?.trim()
  || params.year?.trim()
  || params.publisher?.trim()
  || params.isbn?.trim()
  || params.asin?.trim()
  || params.genres?.trim()
  || params.languages?.trim()
}

function buildQueryParams(params: MetadataSearchParams) {
    const urlParams = new URLSearchParams()
    urlParams.append('source', params.source.toLowerCase())
    if (params.title?.trim()) urlParams.set('title', params.title)
    if (params.author?.trim()) urlParams.set('author', params.author)
    if (params.year?.trim()) urlParams.set('year', params.year)
    if (params.publisher?.trim()) urlParams.set('publisher', params.publisher)
    if (params.isbn?.trim()) urlParams.set('isbn', params.isbn)
    if (params.asin?.trim()) urlParams.set('asin', params.asin)
    if (params.genres?.trim()) params.genres.split(',').forEach(g => urlParams.append('genre', g.trim()))
    if (params.languages?.trim()) params.languages.split(',').forEach(l => urlParams.append('language', l.trim()))
  
    if (params.pageLimit) urlParams.append('limit', params.pageLimit?.toString())
    urlParams.append('page', params.page.toString())

    return urlParams
}

export async function searchMetadataSource(params: MetadataSearchParams): Promise<MetadataSearchResults | null> {
  if (!hasParams(params)) {
    useNotificationsStore().notifyError('Enter at least one search term.')
    return null
  }

  const endpoint = '/api/metadata/'
  const queryParams = buildQueryParams(params)

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
    useNotificationsStore().notifyError(`Metadata search failed: ${err}`)
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