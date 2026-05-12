<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { BookParams } from '@/types/book'
import AddBookModal from '../components/AddBookModal.vue'
import ResultItem from '../components/ResultItem.vue'
import { MetadataType, searchMetadataSource, getMetadataDetails, metadataSearchFields, AudibleRegion } from '@/types/metadata'
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
const asin = ref('')
const genres = ref('')
const languages = ref('')
const keywords = ref('')

const region = ref<AudibleRegion>(AudibleRegion.US)

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
    genres: source.value === MetadataType.Audible ? keywords.value : genres.value,
    languages: languages.value,
    page: page.value,
    pageLimit: limit,
    region: region.value
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
    <h2 class="vue-heading">Add Book</h2>

    <div class="search-panel">
      <div class="search-row">
        <div class="source-region">
          <select class="search-select" v-model="source" aria-label="Metadata source">
            <option v-for="(type, value) in MetadataType" :key="value" :value="type">
              {{ type }}
            </option>
          </select>
          <select class="search-select region" v-model="region" aria-label="Audible Region" v-show="source == MetadataType.Audible">
            <option v-for="(type, value) in AudibleRegion" :key="value" :value="type">
              .{{ type }}
            </option>
          </select>
        </div>
        <input
          class="search-input"
          v-model="title"
          type="text"
          placeholder="Enter title"
          @keyup.enter="searchBooksAndResetPage"
          aria-label="Book title"
        />
        <button class="search-button" type="button" @click="searchBooksAndResetPage" :disabled="loading">Search</button>
        <div class="source-region">
          <button class="toggle-button" type="button" @click="showAdvanced = !showAdvanced">
            {{ showAdvanced ? 'Hide' : 'Advanced' }}
          </button>
          <button class="toggle-button" type="button" @click="resetSearch">Reset</button>
        </div>
      </div>

      <div v-if="showAdvanced" class="advanced-panel">
        <label class="search-field" v-if="metadataSearchFields.get(source)?.authors">
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
        <label class="search-field" v-if="metadataSearchFields.get(source)?.year">
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
        <label class="search-field" v-if="metadataSearchFields.get(source)?.publisher">
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
        <label class="search-field" v-if="metadataSearchFields.get(source)?.isbn">
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
        <label class="search-field" v-if="metadataSearchFields.get(source)?.asin">
          <span>ASIN</span>
          <input
            class="search-input"
            v-model="asin"
            type="text"
            placeholder="ASIN"
            aria-label="ASIN"
            @keyup.enter="searchBooksAndResetPage"
          />
        </label>
        <label class="search-field" v-if="metadataSearchFields.get(source)?.genres">
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
        <label class="search-field" v-if="metadataSearchFields.get(source)?.languages">
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
        <label class="search-field" v-if="metadataSearchFields.get(source)?.keywords">
          <span>Keywords</span>
          <input
            class="search-input"
            v-model="keywords"
            type="text"
            placeholder="Keywords (comma-separated)"
            aria-label="Keywords"
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
  padding-bottom: 10rem;
  box-sizing: border-box;
}

/* Mobile-first: stacked vertical layout */
.add-book .search-row {
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
  margin-bottom: 0.8rem;
  height: 100%;
}

.add-book .search-row .search-input,
.add-book .search-row .search-select {
  font-size: 16px;
  padding: 0.75rem 1rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  min-height: 44px;
  box-sizing: border-box;
  background: var(--color-background);
  color: var(--color-text);
  height: 100%;
}

.add-book .search-row .search-input {
  flex: 1;
  width: 100%;
}

/* Buttons container - flex row on mobile */
.add-book .search-row {
  display: flex;
  flex-direction: column;
}

.add-book .search-row > .search-button,
.add-book .search-row > .toggle-button {
  width: 100%;
}

.search-select.region {
  width: 100%;
}

.source-region {
  display: flex;
  gap: 0.7rem;
  height: 100%;
}

/* Desktop layout: restore multi-column grid */
@media (min-width: 768px) {
  .add-book .search-row {
    display: flex;
    grid-template-columns: minmax(150px, 1fr) auto minmax(280px, 2fr) auto auto auto;
    gap: 0.7rem;
    flex-direction: row;
    align-items: center;
  }

  .add-book .search-row .search-input,
  .add-book .search-row .search-select {
    font-size: 14px;
    padding: 0.5rem 0.8rem;
    min-height: auto;
    border: 1px solid var(--color-border);
    background: var(--color-background);
    color: var(--color-text);
  }

  .add-book .search-row > .search-button,
  .add-book .search-row > .toggle-button {
    width: auto;
  }

  .search-select.region {
    width: 100%;
  }
}

/* Search metadata display */
.search-meta {
  margin-bottom: 0.8rem;
  font-size: 0.9rem;
  color: var(--color-text);
  display: flex;
  gap: 1rem;
  padding-top: 1rem;
}

/* Error message styling */
.error {
  color: var(--color-error);
  background: var(--color-error-background);
  border: 1px solid var(--color-error-border);
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

/* Search button - prominent primary style */
.search-button {
  background: var(--color-primary-blue) !important;
  color: var(--vt-c-text-light-1) !important;
  font-weight: 600 !important;
  padding: 0.75rem 1.5rem !important;
  border: none !important;
  font-size: 16px !important;
  min-height: 44px !important;
  cursor: pointer !important;
  border-radius: 6px !important;
  transition: all 0.2s ease !important;
}

.search-button:hover:not(:disabled) {
  background: var(--color-primary-blue-dark) !important;
  transform: translateY(-1px) !important;
}

.search-button:active:not(:disabled) {
  transform: translateY(0) !important;
}

.search-button:disabled {
  background: var(--color-gray-400) !important;
  cursor: not-allowed !important;
  opacity: 0.6 !important;
}

/* Toggle buttons */
.toggle-button {
  padding: 0.75rem 1rem !important;
  border: 1px solid var(--color-border) !important;
  border-radius: 6px !important;
  background: var(--color-background-soft) !important;
  color: var(--color-text) !important;
  cursor: pointer !important;
  font-size: 16px !important;
  min-height: 44px !important;
  transition: all 0.2s ease !important;
}

.toggle-button:hover:not(:disabled) {
  background: var(--color-background-mute) !important;
  border-color: var(--color-border-hover) !important;
}

@media (min-width: 768px) {
  .search-button {
    font-size: 14px !important;
    padding: 0.5rem 1rem !important;
    min-height: auto !important;
  }

  .toggle-button {
    font-size: 14px !important;
    padding: 0.5rem 0.8rem !important;
    min-height: auto !important;
  }
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
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-background-soft);
  color: var(--color-text);
  cursor: pointer;
}

.pagination button:disabled {
  background: var(--color-gray-300);
  cursor: not-allowed;
}

.pagination button:hover:not(:disabled) {
  background: var(--color-background-mute);
}

.pagination span {
  font-size: 0.9rem;
  color: var(--color-text);
}

/* Empty state message */
.empty-state {
  margin-top: 0.8rem;
  color: var(--vt-c-text-subtle);
  text-align: center;
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
