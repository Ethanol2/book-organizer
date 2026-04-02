<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

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

const route = useRoute()
const router = useRouter()

const source = ref<MetadataSource>('open library')
const title = ref('')
const author = ref('')
const year = ref('')
const publisher = ref('')
const isbn = ref('')
const genres = ref('')
const languages = ref('')
const page = ref(1)
const limit = 10
const results = ref<MetadataItem[]>([])
const totalCount = ref(0)
const offset = ref(0)
const count = ref(0)
const loading = ref(false)
const error = ref('')
const showAdvanced = ref(false)
const showModal = ref(false)
const selectedItem = ref<MetadataItem | null>(null)

const modalTitle = ref('')
const modalSubtitle = ref('')
const modalDescription = ref('')
const modalYear = ref('')
const modalIsbn = ref('')
const modalPublisher = ref('')
const modalAuthors = ref('')
const modalGenres = ref('')
const modalCover = ref('')

const sourceLabel = (s: MetadataSource) => (s === 'open library' ? 'Open Library' : 'Google Books')

const selectedSourceName = computed(() => sourceLabel(source.value))

const hasMultiplePages = computed(() => totalCount.value > limit)
const isFirstPage = computed(() => page.value === 1)
const isLastPage = computed(() => (page.value - 1) * limit + count.value >= totalCount.value)

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

async function addBook() {
  if (!selectedItem.value) return

  const bookData = {
    title: modalTitle.value || null,
    subtitle: modalSubtitle.value || null,
    description: modalDescription.value || null,
    year: modalYear.value ? parseInt(modalYear.value) : null,
    isbn: modalIsbn.value || null,
    publisher: modalPublisher.value || null,
    authors: modalAuthors.value ? modalAuthors.value.split(',').map(name => ({ name: name.trim() })) : null,
    genres: modalGenres.value ? modalGenres.value.split(',').map(name => ({ name: name.trim() })) : null,
    cover: modalCover.value || null,
  }

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

const formattedAuthors = (item: MetadataItem) => {
  if (!item.authors || item.authors.length === 0) {
    return 'Unknown author'
  }
  return item.authors.map(a => a.name || 'unknown').join(', ')
}
</script>

<template>
  <section class="add-book">
    <h2>Add Book</h2>
    <p class="hint">Search metadata from Open Library or Google Books and review results.</p>

    <div class="search-row">
      <select v-model="source" aria-label="Metadata source">
        <option value="open library">Open Library</option>
        <option value="google books">Google Books</option>
      </select>
      <input
        v-model="title"
        type="text"
        placeholder="Enter title"
        @keyup.enter="searchBooks"
        aria-label="Book title"
      />
      <button type="button" @click="searchBooks" :disabled="loading">Search</button>
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
      <li class="result-item" v-for="(item, index) in results" :key="index" @click="openModal(item)">
        <div class="result-cover">
          <img v-if="item.cover" :src="item.cover" :alt="item.title || 'cover'" />
          <div v-else class="cover-placeholder"></div>
        </div>
        <div class="result-details">
          <h3>{{ item.title || 'Untitled' }}</h3>
          <p class="subtitle" v-if="item.subtitle">{{ item.subtitle }}</p>
          <p class="meta">
            <strong>Author:</strong> {{ formattedAuthors(item) }}
            <span v-if="item.year">• {{ item.year }}</span>
            <span v-if="item.isbn">• ISBN {{ item.isbn }}</span>
          </p>
          <p class="publisher" v-if="item.publisher"><strong>Publisher:</strong> {{ item.publisher }}</p>
          <p class="description" v-if="item.description">{{ item.description }}</p>
        </div>
      </li>
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

  <div v-if="showModal" class="modal-overlay" @click="closeModal">
    <div class="modal" @click.stop>
      <h3>Add Book</h3>
      <form @submit.prevent="addBook">
        <label>Title: <input v-model="modalTitle" required /></label>
        <label>Subtitle: <input v-model="modalSubtitle" /></label>
        <label>Description: <textarea v-model="modalDescription"></textarea></label>
        <label>Year: <input v-model="modalYear" type="number" /></label>
        <label>ISBN: <input v-model="modalIsbn" /></label>
        <label>Publisher: <input v-model="modalPublisher" /></label>
        <label>Authors (comma-separated): <input v-model="modalAuthors" /></label>
        <label>Genres (comma-separated): <input v-model="modalGenres" /></label>
        <label>Cover URL: <input v-model="modalCover" /></label>
        <div class="modal-buttons">
          <button type="button" @click="closeModal">Cancel</button>
          <button type="submit">Add Book</button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.add-book {
  max-width: 900px;
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
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
.advanced-search .search-row {
  grid-template-columns: 1fr 1fr;
  margin-bottom: 0.5rem;
}
.advanced-search .search-row:last-child {
  margin-bottom: 0.8rem;
}
.search-meta {
  margin-bottom: 0.8rem;
  font-size: 0.9rem;
  color: #555;
  display: flex;
  gap: 1rem;
}
.error {
  color: #a00;
  background: #ffe5e5;
  border: 1px solid #ddaaaa;
  border-radius: 6px;
  padding: 0.7rem;
  margin-bottom: 1rem;
}
.results {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 0.8rem;
  flex: 1;
  overflow-y: auto;
}
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
.result-item {
  border: 1px solid #ccc;
  border-radius: 8px;
  padding: 0.8rem;
  display: grid;
  grid-template-columns: 160px 1fr;
  gap: 0.8rem;
  cursor: pointer;
}
.result-item:hover {
  background: #f9f9f9;
}
.cover-placeholder {
  width: 100%;
  height: 250px;
  background: #eee;
  border: 1px solid #ccc;
  border-radius: 4px;
}
.result-cover img {
  width: 100%;
  max-height: 250px;
  object-fit: contain;
}
.result-details h3 {
  margin: 0;
}
.result-details .subtitle {
  margin: 0.2rem 0;
  color: #333;
}
.result-details .meta,
.result-details .publisher {
  margin: 0.2rem 0;
  font-size: 0.9rem;
  color: #555;
}
.description {
  margin-top: 0.45rem;
  line-height: 1.35;
  color: #444;
}
.empty-state {
  margin-top: 0.8rem;
  color: #666;
}
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
}
.modal {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  max-width: 600px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}
.modal h3 {
  margin-top: 0;
}
.modal label {
  display: block;
  margin-bottom: 0.5rem;
}
.modal input,
.modal textarea {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  margin-bottom: 1rem;
}
.modal textarea {
  height: 100px;
}
.modal-buttons {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}
.modal-buttons button {
  padding: 0.5rem 1rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  background: #f9f9f9;
  cursor: pointer;
}
.modal-buttons button[type="submit"] {
  background: #4CAF50;
  color: white;
}
</style>
