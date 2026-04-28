<script setup lang="ts">
import BookItem from '@/components/LibraryItem.vue';
import type { BookSummary } from '@/types/book';
import { onMounted, ref } from 'vue';
import { useRoute, useRouter, type LocationQueryValue } from 'vue-router';

const route = useRoute();
const router = useRouter();
const books = ref<BookSummary[]>([]);

const search = ref('');
const sortBy = ref('title');
const sortOrder = ref('asc');
const subtitle = ref('');
const description = ref('');
const year = ref('');
const isbn = ref('');
const asin = ref('');
const publisher = ref('');
const series = ref('');
const tags = ref('');
const genres = ref('');
const authors = ref('');
const narrators = ref('');
const files = ref(null);
const showAdvanced = ref(false);

// Pagination state
const currentPage = ref(1);
const itemsPerPage = ref(20);
const totalResults = ref(0);
const isLoading = ref(false);
const hasLoadedAll = ref(false);

function normalizeQueryValue(value: LocationQueryValue | LocationQueryValue[] | undefined): string {
    if (Array.isArray(value)) {
        return String(value[0] ?? '');
    }
    return String(value ?? '');
}

function resetFiltersFromRoute() {
    search.value = normalizeQueryValue(route.query.search);
    sortBy.value = normalizeQueryValue(route.query.sortBy) || 'title';
    sortOrder.value = normalizeQueryValue(route.query.sortOrder) || 'asc';
    subtitle.value = normalizeQueryValue(route.query.subtitle);
    description.value = normalizeQueryValue(route.query.description);
    year.value = normalizeQueryValue(route.query.year);
    isbn.value = normalizeQueryValue(route.query.isbn);
    asin.value = normalizeQueryValue(route.query.asin);
    publisher.value = normalizeQueryValue(route.query.publisher);
    series.value = normalizeQueryValue(route.query.series);
    tags.value = normalizeQueryValue(route.query.tags);
    genres.value = normalizeQueryValue(route.query.genres);
    authors.value = normalizeQueryValue(route.query.authors);
    narrators.value = normalizeQueryValue(route.query.narrators);

    showAdvanced.value = Boolean(
        subtitle.value ||
        description.value ||
        year.value ||
        isbn.value ||
        asin.value ||
        publisher.value ||
        series.value ||
        tags.value ||
        genres.value ||
        authors.value ||
        narrators.value,
    );
}

function buildQueryString(page?: number) {
    const params = new URLSearchParams();

    params.append('view', 'summary');
    const pageNum = page ?? currentPage.value;
    params.append('page', pageNum.toString());
    params.append('count', itemsPerPage.value.toString());
    if (search.value.trim()) params.append('search', search.value.trim());
    if (sortBy.value) params.append('sortBy', sortBy.value);
    if (sortOrder.value) params.append('sortOrder', sortOrder.value);
    if (subtitle.value.trim()) params.append('subtitle', subtitle.value.trim());
    if (description.value.trim()) params.append('description', description.value.trim());
    if (year.value.trim()) params.append('publish_year', year.value.trim());
    if (isbn.value.trim()) params.append('isbn', isbn.value.trim());
    if (asin.value.trim()) params.append('asin', asin.value.trim());
    if (publisher.value.trim()) params.append('publisher', publisher.value.trim());
    if (series.value.trim()) params.append('series', series.value.trim());
    if (tags.value.trim()) params.append('tags', tags.value.trim());
    if (genres.value.trim()) params.append('genres', genres.value.trim());
    if (authors.value.trim()) params.append('authors', authors.value.trim());
    if (narrators.value.trim()) params.append('narrators', narrators.value.trim());
    if (files.value) params.append('files', files.value);

    const queryString = params.toString();
    return queryString ? `?${queryString}` : '';
}

function syncRouteQuery() {
    router.replace({
        query: {
            ...(search.value.trim() ? { search: search.value.trim() } : {}),
            ...(sortBy.value ? { sortBy: sortBy.value } : {}),
            ...(sortOrder.value ? { sortOrder: sortOrder.value } : {}),
            ...(subtitle.value.trim() ? { subtitle: subtitle.value.trim() } : {}),
            ...(description.value.trim() ? { description: description.value.trim() } : {}),
            ...(year.value.trim() ? { publish_year: year.value.trim() } : {}),
            ...(isbn.value.trim() ? { isbn: isbn.value.trim() } : {}),
            ...(asin.value.trim() ? { asin: asin.value.trim() } : {}),
            ...(publisher.value.trim() ? { publisher: publisher.value.trim() } : {}),
            ...(series.value.trim() ? { series: series.value.trim() } : {}),
            ...(tags.value.trim() ? { tags: tags.value.trim() } : {}),
            ...(genres.value.trim() ? { genres: genres.value.trim() } : {}),
            ...(authors.value.trim() ? { authors: authors.value.trim() } : {}),
            ...(narrators.value.trim() ? { narrators: narrators.value.trim() } : {}),
        },
    });
}

async function fetchBooks(append = false) {
    if (isLoading.value || hasLoadedAll.value) return;
    
    isLoading.value = true;
    try {
        const queryString = buildQueryString();
        const resp = await fetch(`api/books${queryString}`);
        if (!resp.ok) {
            throw new Error(`HTTP error with status: ${resp.status}`);
        }

        const data = await resp.json();
        
        if (append) {
            books.value.push(...data.items);
        } else {
            if (data.items == null || data.items.length === 0) {
                books.value = [];
            }
            else {
                books.value = data.items;
            }
        }
        
        totalResults.value = data.results_count;
        currentPage.value = data.page;
        
        // Check if we've loaded all items
        const loadedCount = append 
            ? books.value.length 
            : data.items.length;
        
        if (loadedCount >= data.results_count) {
            hasLoadedAll.value = true;
        }

    } catch (error) {
        console.error('Error fetching books list:', error);
    } finally {
        isLoading.value = false;
    }
}

async function searchBooks() {
    currentPage.value = 1;
    hasLoadedAll.value = false;
    syncRouteQuery();
    await fetchBooks(false);
}

async function loadMore() {
    if (isLoading.value || hasLoadedAll.value) return;
    currentPage.value++;
    await fetchBooks(true);
}

function resetFilters() {
    search.value = '';
    subtitle.value = '';
    description.value = '';
    year.value = '';
    isbn.value = '';
    asin.value = '';
    publisher.value = '';
    series.value = '';
    tags.value = '';
    genres.value = '';
    authors.value = '';
    narrators.value = '';
    showAdvanced.value = false;
    currentPage.value = 1;
    hasLoadedAll.value = false;
    syncRouteQuery();
    fetchBooks(false);
}

onMounted(async () => {
    resetFiltersFromRoute();
    await fetchBooks(false);
    
    // Set up intersection observer for infinite scroll
    const observer = new IntersectionObserver(
        (entries) => {
            if (entries[0]?.isIntersecting) {
                loadMore();
            }
        },
        { threshold: 0.1 }
    );
    
    // Get the sentinel element (will be the last child that loads more when visible)
    const sentinel = document.querySelector('.library-sentinel');
    if (sentinel) {
        observer.observe(sentinel);
    }
});

</script>

<template>
    <h2>Library</h2>
    
    <section>
        <div class="library-controls">
            <div class="search-panel">
              <div class="search-row library-search-row">
                  <input
                      class="search-input"
                      v-model="search"
                      type="search"
                      placeholder="Search title, subtitle, or description"
                      aria-label="Search books"
                        @keyup.enter="searchBooks"
                  />

                  <button class="search-button" type="button" @click="searchBooks">Search</button>
                  <button class="toggle-button" type="button" @click="showAdvanced = !showAdvanced">
                      {{ showAdvanced ? 'Hide Advanced' : 'Advanced Search' }}
                  </button>
                  <button class="toggle-button" type="button" @click="resetFilters">Reset</button>
              </div>

              <div v-if="showAdvanced" class="advanced-panel">
                <label class="search-field">
                    <span>Year</span>
                    <input class="search-input" v-model="year" type="text" placeholder="Year" @keyup.enter="searchBooks" />
                </label>
                <label class="search-field">
                    <span>ISBN</span>
                    <input class="search-input" v-model="isbn" type="text" placeholder="ISBN" @keyup.enter="searchBooks" />
                </label>
                <label class="search-field">
                    <span>ASIN</span>
                    <input class="search-input" v-model="asin" type="text" placeholder="ASIN" @keyup.enter="searchBooks" />
                </label>
                <label class="search-field">
                    <span>Publisher</span>
                    <input class="search-input" v-model="publisher" type="text" placeholder="Publisher" @keyup.enter="searchBooks" />
                </label>
                <label class="search-field">
                    <span>Series</span>
                    <input class="search-input" v-model="series" type="text" placeholder="Series" @keyup.enter="searchBooks" />
                </label>
                <label class="search-field">
                    <span>Tags</span>
                    <input class="search-input" v-model="tags" type="text" placeholder="tag1, tag2" @keyup.enter="searchBooks" />
                </label>
                <label class="search-field">
                    <span>Genres</span>
                    <input class="search-input" v-model="genres" type="text" placeholder="genre1, genre2" @keyup.enter="searchBooks" />
                </label>
                <label class="search-field">
                    <span>Authors</span>
                    <input class="search-input" v-model="authors" type="text" placeholder="author1, author2" @keyup.enter="searchBooks" />
                </label>
                <label class="search-field">
                    <span>Narrators</span>
                    <input class="search-input" v-model="narrators" type="text" placeholder="narrator1, narrator2" @keyup.enter="searchBooks" />
                </label>
              </div>

              <div class="sort-row">
                  <label class="search-field">
                      <span>Files</span>
                      <select class="search-select" v-model="files" @change="searchBooks">
                          <option value=null>All</option>
                          <option value="with_files">Has Files</option>
                          <option value="without_files">No Files</option>
                      </select>
                  </label>

                  <label class="search-field">
                      <span>Sort by</span>
                      <select class="search-select" v-model="sortBy" @change="searchBooks">
                          <option value="created_at">Date Added</option>
                          <option value="title">Title</option>
                          <option value="author">Author</option>
                          <option value="publish_year">Year</option>
                          <option value="series">Series</option>
                          <option value="publisher">Publisher</option>
                      </select>
                  </label>

                  <label class="search-field">
                      <span>Order</span>
                      <select class="search-select" v-model="sortOrder" @change="searchBooks">
                          <option value="asc">Ascending</option>
                          <option value="desc">Descending</option>
                      </select>
                  </label>
              </div>
            </div>
        </div>

        <div class="library">
            <BookItem v-for="book in books" :key="book.id" :book="book"></BookItem>
            <div v-if="isLoading" class="library-loading">
                <div class="spinner"></div>
            </div>
            <div v-if="!hasLoadedAll && books.length > 0" class="library-sentinel"></div>
        </div>
    </section>
</template>

<style scoped>
.library {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 16px;
    padding-bottom: 10rem;
}

.library-loading {
    grid-column: 1 / -1;
    display: flex;
    justify-content: center;
    padding: 2rem;
}

.spinner {
    width: 40px;
    height: 40px;
    border: 4px solid var(--color-border);
    border-top-color: var(--color-text);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
}

@keyframes spin {
    to {
        transform: rotate(360deg);
    }
}

.library-sentinel {
    grid-column: 1 / -1;
    height: 1px;
}

.library-controls {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-bottom: 1.5rem;
}

.search-field {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    font-size: 0.95rem;
}

.search-field span {
    color: var(--color-text);
    font-size: 0.82rem;
}

.advanced-panel {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 1rem;
}

.search-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto auto auto;
    gap: 0.75rem;
    align-items: stretch;
}

.search-row button,
.search-row .search-input {
    min-height: 42px;
}

.sort-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
  justify-content: flex-start;
}

@media (max-width: 900px) {
    .search-row,
    .sort-row {
        grid-template-columns: 1fr;
    }
}

@media (max-width: 600px) {
    .advanced-panel {
        grid-template-columns: 1fr;
    }
    .sort-row {
        flex-direction: column;
        align-items: flex-start;
    }
}
</style>