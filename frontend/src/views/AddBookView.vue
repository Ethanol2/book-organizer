<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useNotificationsStore } from '@/stores/notifications'
import type { Book, BookParams } from '@/types/book'
import AddBookModal from '../components/AddBookModal.vue'
import ResultItem from '../components/ResultItem.vue'
import { MetadataType, searchMetadataSource, getMetadataDetails } from '@/types/metadata'
import { postBook } from '@/types/book'

// Router utilities for URL parameter management
const route = useRoute()

// Search filter state
const source = ref<MetadataType>(MetadataType.OpenLibrary)
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
const results = ref<BookParams[]>([])
const totalCount = ref(0)
const offset = ref(0)
const count = ref(0)

// UI state
const loading = ref(false)
const error = ref('')
const showAdvanced = ref(false)
const showModal = ref(false)
const selectedItem = ref<BookParams | null>(null)

const notifications = useNotificationsStore()

// Pagination computed properties
const hasMultiplePages = computed(() => totalCount.value > limit)
const isFirstPage = computed(() => page.value === 1)
const isLastPage = computed(() => (page.value - 1) * limit + count.value >= totalCount.value)

function resetSearch() {
  title.value = ''
  author.value = ''
  year.value = ''
  publisher.value = ''
  isbn.value = ''
  genres.value = ''
  languages.value = ''
  page.value = 1
  results.value = []
  totalCount.value = 0
  offset.value = 0
  count.value = 0
}

async function addBook(book: BookParams) {
  const ok = await postBook(book)
  if (ok != null) closeModal()
}

async function searchBooks() {
  loading.value = true

  const searchResults = await searchMetadataSource({
    source: source.value,
    title: title.value,
    author: author.value,
    year: year.value,
    publisher: publisher.value,
    isbn: isbn.value,
    genres: genres.value,
    languages: languages.value,
    page: page.value,
    pageLimit: limit,
  })

  //console.log(searchResults)

  results.value = searchResults?.items ?? []
  totalCount.value = searchResults?.total_count ?? 0
  offset.value = searchResults?.offset ?? 0
  count.value = searchResults?.count ?? 0

  loading.value = false
}

// Fetch search results from metadata API
async function searchBooksAndResetPage() {
  page.value = 1
  await searchBooks()
}

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
async function openModal(item: BookParams) {
  const details = await getMetadataDetails(item)
  selectedItem.value = details ?? item
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  selectedItem.value = null
}

</script>

<template>
  <section class="add-book">
    <h2>Add Book</h2>

    <div class="search-panel">
      <div class="search-row">
        <select class="search-select" v-model="source" aria-label="Metadata source">
          <option v-for="(type, value) in MetadataType" :key="value" :value="type">
            {{ type }}
          </option>
        </select>
        <input
          class="search-input"
          v-model="title"
          type="text"
          placeholder="Enter title"
          @keyup.enter="searchBooksAndResetPage"
          aria-label="Book title"
        />
        <button class="search-button" type="button" @click="searchBooksAndResetPage" :disabled="loading">Search</button>
        <button class="toggle-button" type="button" @click="showAdvanced = !showAdvanced">
          {{ showAdvanced ? 'Hide Advanced' : 'Advanced Search' }}
        </button>
        <button class="toggle-button" type="button" @click="resetSearch">Reset</button>
      </div>

      <div v-if="showAdvanced" class="advanced-panel">
        <label class="search-field">
          <span>Author</span>
          <input
            class="search-input"
            v-model="author"
            type="text"
            placeholder="Author"
            aria-label="Author"
            @keyup.enter="searchBooksAndResetPage"
          />
        </label>
        <label class="search-field">
          <span>Year</span>
          <input
            class="search-input"
            v-model="year"
            type="text"
            placeholder="Year"
            aria-label="Year"
            @keyup.enter="searchBooksAndResetPage"
          />
        </label>
        <label class="search-field">
          <span>Publisher</span>
          <input
            class="search-input"
            v-model="publisher"
            type="text"
            placeholder="Publisher"
            aria-label="Publisher"
            @keyup.enter="searchBooksAndResetPage"
          />
        </label>
        <label class="search-field">
          <span>ISBN</span>
          <input
            class="search-input"
            v-model="isbn"
            type="text"
            placeholder="ISBN"
            aria-label="ISBN"
            @keyup.enter="searchBooksAndResetPage"
          />
        </label>
        <label class="search-field">
          <span>Genres</span>
          <input
            class="search-input"
            v-model="genres"
            type="text"
            placeholder="Genres (comma-separated)"
            aria-label="Genres"
            @keyup.enter="searchBooksAndResetPage"
          />
        </label>
        <label class="search-field">
          <span>Languages</span>
          <input
            class="search-input"
            v-model="languages"
            type="text"
            placeholder="Languages (comma-separated)"
            aria-label="Languages"
            @keyup.enter="searchBooksAndResetPage"
          />
        </label>
      </div>
    </div>

    <div class="search-meta">
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
    :params="selectedItem"
    @close="closeModal"
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

.add-book .search-row {
  display: grid;
  grid-template-columns: minmax(150px, 190px) minmax(280px, 1fr) auto auto auto;
  gap: 0.7rem;
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

  .results {
    gap: 0.4rem;
  }
}
</style>
