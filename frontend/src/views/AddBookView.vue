<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AddBookModal from '../components/AddBookModal.vue'
import ResultItem from '../components/ResultItem.vue'

// Type definitions for metadata search functionality
type MetadataSource = 'open library' | 'google books'

type Category = { id?: string; name?: string }

type MetadataItem = {
  title?: string | null
  subtitle?: string | null
  description?: string | null
  year?: number | null
  isbn?: string | null
  publisher?: string | null
  authors?: Category[] | null
  genres?: Category[] | null
  cover?: string | null
}

type SearchResults = {
  total_count: number
  count: number
  offset: number
  items: MetadataItem[]
}

// Router utilities for URL parameter management
const route = useRoute()
const router = useRouter()

// Search filter state
const source = ref<MetadataSource>('open library')
const title = ref('')
const author = ref('')
const year = ref('')
const publisher = ref('')
const isbn = ref('')
const genres = ref('')
const languages = ref('')

// Pagination and result state
const page = ref(1)
const limit = 10
const results = ref<MetadataItem[]>([])
const totalCount = ref(0)
const offset = ref(0)
const count = ref(0)

// UI state
const loading = ref(false)
const error = ref('')
const showAdvanced = ref(false)
const showModal = ref(false)
const selectedItem = ref<MetadataItem | null>(null)

// Modal form state for adding books
const modalTitle = ref('')
const modalSubtitle = ref('')
const modalDescription = ref('')
const modalYear = ref('')
const modalIsbn = ref('')
const modalPublisher = ref('')
const modalAuthors = ref('')
const modalGenres = ref('')
const modalCover = ref('')

// Computed properties and utility functions
const sourceLabel = (s: MetadataSource) => (s === 'open library' ? 'Open Library' : 'Google Books')
const selectedSourceName = computed(() => sourceLabel(source.value))

// Pagination computed properties
const hasMultiplePages = computed(() => totalCount.value > limit)
const isFirstPage = computed(() => page.value === 1)
const isLastPage = computed(() => (page.value - 1) * limit + count.value >= totalCount.value)

// Build URL query parameters from current search filters
function buildQueryParams() {
  const params = new URLSearchParams()
  params.append('source', source.value)
  if (title.value.trim()) params.append('title', title.value.trim())
  if (author.value.trim()) params.append('author', author.value.trim())
  if (year.value.trim()) params.append('year', year.value.trim())
  if (publisher.value.trim()) params.append('publisher', publisher.value.trim())
  if (isbn.value.trim()) params.append('isbn', isbn.value.trim())
  if (genres.value.trim()) {
    genres.value.split(',').forEach(g => params.append('genre', g.trim()))
  }
  if (languages.value.trim()) {
    languages.value.split(',').forEach(l => params.append('language', l.trim()))
  }
  params.append('page', page.value.toString())
  params.append('limit', limit.toString())
  return params
}

// Fetch search results from metadata API
async function searchBooksAndResetPage() {
  page.value = 1
  await searchBooks()
}
async function searchBooks() {
  error.value = ''
  results.value = []

  const hasQuery = title.value.trim() || author.value.trim() || year.value.trim() || publisher.value.trim() || isbn.value.trim() || genres.value.trim() || languages.value.trim()
  if (!hasQuery) {
    error.value = 'Enter at least one search term.'
    return
  }

  const endpoint = '/api/metadata/'
  const queryParams = buildQueryParams()

  loading.value = true
  try {
    const resp = await fetch(`${endpoint}?${queryParams.toString()}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    if (!resp.ok) {
      const body = await resp.text()
      throw new Error(`${resp.status} ${resp.statusText}: ${body}`)
    }

    const body = (await resp.json()) as SearchResults
    totalCount.value = body.total_count ?? body.count ?? 0
    offset.value = body.offset ?? 0
    count.value = body.count ?? body.items?.length ?? 0
    results.value = body.items ?? []

    if (results.value.length === 0) {
      error.value = `No results from ${selectedSourceName.value}`
    }

    // Update URL with pagination params
    router.push({ query: { ...route.query, page: page.value.toString(), limit: limit.toString() } })
  } catch (err) {
    console.error('Search API error', err)
    error.value = 'Search failed. ' + (err instanceof Error ? err.message : String(err))
  } finally {
    loading.value = false
  }
}

// Pagination handlers

// Pagination handlers
function prevPage() {
  if (page.value > 1) {
    page.value--
    searchBooks()
  }
}

function nextPage() {
  page.value++
  searchBooks()
}

// Initialize page from URL
page.value = parseInt(route.query.page as string) || 1

// Modal management functions
function openModal(item: MetadataItem) {
  selectedItem.value = item
  modalTitle.value = item.title || ''
  modalSubtitle.value = item.subtitle || ''
  modalDescription.value = item.description || ''
  modalYear.value = item.year?.toString() || ''
  modalIsbn.value = item.isbn || ''
  modalPublisher.value = item.publisher || ''
  modalAuthors.value = item.authors?.map(a => a.name).join(', ') || ''
  modalGenres.value = item.genres?.map(g => g.name).join(', ') || ''
  modalCover.value = item.cover || ''
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  selectedItem.value = null
}

// Post book data to backend API
async function addBook(bookData: {
  title: string | null
  subtitle: string | null
  description: string | null
  year: number | null
  isbn: string | null
  publisher: string | null
  authors: Array<{ name: string }> | null
  genres: Array<{ name: string }> | null
  cover: string | null
}) {
  try {
    const resp = await fetch('/api/books', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(bookData),
    })

    if (!resp.ok) {
      const body = await resp.text()
      throw new Error(`${resp.status} ${resp.statusText}: ${body}`)
    }

    alert('Book added successfully!')
    closeModal()
  } catch (err) {
    console.error('Add book error', err)
    alert('Failed to add book: ' + (err instanceof Error ? err.message : String(err)))
  }
}

</script>

<template>
  <section class="add-book">
    <h2>Add Book</h2>

    <div class="search-row">
      <select v-model="source" aria-label="Metadata source">
        <option value="open library">Open Library</option>
        <option value="google books">Google Books</option>
      </select>
      <input
        v-model="title"
        type="text"
        placeholder="Enter title"
        @keyup.enter="searchBooksAndResetPage"
        aria-label="Book title"
      />
      <button type="button" @click="searchBooksAndResetPage" :disabled="loading">Search</button>
    </div>

    <div class="advanced-toggle">
      <button type="button" @click="showAdvanced = !showAdvanced">
        {{ showAdvanced ? 'Hide' : 'Show' }} Advanced Search
      </button>
    </div>

    <div v-if="showAdvanced" class="advanced-search">
      <div class="search-row">
        <input
          v-model="author"
          type="text"
          placeholder="Author"
          aria-label="Author"
        />
        <input
          v-model="year"
          type="text"
          placeholder="Year"
          aria-label="Year"
        />
      </div>
      <div class="search-row">
        <input
          v-model="publisher"
          type="text"
          placeholder="Publisher"
          aria-label="Publisher"
        />
        <input
          v-model="isbn"
          type="text"
          placeholder="ISBN"
          aria-label="ISBN"
        />
      </div>
      <div class="search-row">
        <input
          v-model="genres"
          type="text"
          placeholder="Genres (comma-separated)"
          aria-label="Genres"
        />
        <input
          v-model="languages"
          type="text"
          placeholder="Languages (comma-separated)"
          aria-label="Languages"
        />
      </div>
    </div>

    <div class="search-meta">
      <span>Source: {{ selectedSourceName }}</span>
      <span v-if="count > 0">Results: {{ count }} / {{ totalCount }}</span>
      <span v-if="loading">Loading ...</span>
    </div>

    <div v-if="error" class="error">{{ error }}</div>

    <ul class="results" v-if="results.length > 0">
      <ResultItem
        v-for="(item, index) in results"
        :key="index"
        :item="item"
        @click="openModal(item)"
      />
    </ul>

    <div v-if="hasMultiplePages" class="pagination">
      <button @click="prevPage" :disabled="isFirstPage">Previous</button>
      <span>Page {{ page }}</span>
      <button @click="nextPage" :disabled="isLastPage">Next</button>
    </div>

    <div v-if="!loading && results.length === 0 && !error" class="empty-state">
      Start typing a title above and press Search.
    </div>
  </section>

  <!-- Modal component for adding books -->
  <AddBookModal
    :show="showModal"
    :title="modalTitle"
    :subtitle="modalSubtitle"
    :description="modalDescription"
    :year="modalYear"
    :isbn="modalIsbn"
    :publisher="modalPublisher"
    :authors="modalAuthors"
    :genres="modalGenres"
    :cover="modalCover"
    @close="closeModal"
    @update:title="modalTitle = $event"
    @update:subtitle="modalSubtitle = $event"
    @update:description="modalDescription = $event"
    @update:year="modalYear = $event"
    @update:isbn="modalIsbn = $event"
    @update:publisher="modalPublisher = $event"
    @update:authors="modalAuthors = $event"
    @update:genres="modalGenres = $event"
    @update:cover="modalCover = $event"
    @add-book="addBook"
  />
</template>

<style scoped>
/* Main container - entire view scrolls vertically */
.add-book {
  display: block;
  overflow-y: auto;
  padding: 1rem;
  padding-bottom: 10rem;
  box-sizing: border-box;
}

/* Search controls section */
.search-row {
  display: grid;
  grid-template-columns: 180px 1fr auto;
  gap: 0.7rem;
  margin-bottom: 0.8rem;
}

.search-row select,
.search-row input,
.search-row button {
  padding: 0.65rem;
  border: 1px solid #ccc;
  border-radius: 6px;
  font-size: 0.95rem;
}

.search-row button {
  min-width: 120px;
}

/* Advanced search toggle button */
.advanced-toggle {
  margin-bottom: 0.8rem;
}

.advanced-toggle button {
  padding: 0.5rem 1rem;
  border: 1px solid #ccc;
  border-radius: 6px;
  background: #f9f9f9;
  cursor: pointer;
}

.advanced-toggle button:hover {
  background: #e9e9e9;
}

/* Advanced search fields - two columns */
.advanced-search .search-row {
  grid-template-columns: 1fr 1fr;
  margin-bottom: 0.5rem;
}

.advanced-search .search-row:last-child {
  margin-bottom: 0.8rem;
}

/* Search metadata display */
.search-meta {
  margin-bottom: 0.8rem;
  font-size: 0.9rem;
  color: #555;
  display: flex;
  gap: 1rem;
}

/* Error message styling */
.error {
  color: #a00;
  background: #ffe5e5;
  border: 1px solid #ddaaaa;
  border-radius: 6px;
  padding: 0.7rem;
  margin-bottom: 1rem;
}

/* Results list - scrollable, fills remaining space */
.results {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 0.8rem;
}

/* Pagination controls */
.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin-top: 1rem;
}

.pagination button {
  padding: 0.5rem 1rem;
  border: 1px solid #ccc;
  border-radius: 6px;
  background: #f9f9f9;
  cursor: pointer;
}

.pagination button:disabled {
  background: #e0e0e0;
  cursor: not-allowed;
}

.pagination button:hover:not(:disabled) {
  background: #e9e9e9;
}

.pagination span {
  font-size: 0.9rem;
  color: #555;
}

/* Empty state message */
.empty-state {
  margin-top: 0.8rem;
  color: #666;
}

/* Responsive design for landscape orientation */
@media (orientation: landscape) and (max-height: 500px) {
  .add-book {
    padding: 0.5rem;
  }

  .search-row {
    margin-bottom: 0.4rem;
  }

  .advanced-toggle {
    margin-bottom: 0.4rem;
  }

  .results {
    gap: 0.4rem;
  }
}

/* Responsive design for smaller screens */
@media (max-width: 600px) {
  .search-row {
    grid-template-columns: 1fr;
  }

  .advanced-search .search-row {
    grid-template-columns: 1fr;
  }
}
</style>
