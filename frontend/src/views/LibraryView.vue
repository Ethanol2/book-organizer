<script setup lang="ts">
import BookItem from '@/components/LibraryItem.vue';
import type { BookSummary } from '@/types/book';
import { onMounted, ref } from 'vue';
import { useRoute, useRouter, type LocationQueryValue } from 'vue-router';
import { useNotificationsStore } from '@/stores/notifications';
import api from '@/services/api';
import SearchBar from '@/components/SearchBar.vue';
import { AddAdvancedTermsToQuery, AdvancedTermsAreEmpty, Empty as EmptyAdvancedSearchTerms, TrimAdvancedTerms, type AdvancedSearchFields, type AdvancedSearchTerms } from '@/types/advancedSearch';
import AdvancedSearch from '@/components/AdvancedSearch.vue';

const route = useRoute();
const router = useRouter();
const books = ref<BookSummary[]>([]);

const searchTerm = ref('');
const advancedSearchComponent = ref<InstanceType<typeof AdvancedSearch> | null>(null);
const advancedSearchFields: AdvancedSearchFields = {
    year: true, isbn: true, asin: true, publisher: true, series: true, tags: true, genres: true, authors: true, narrators: true, keywords: false, languages: false
}
const advancedSearchTerms = ref<AdvancedSearchTerms>({
    year: '', isbn: '', asin: '', publisher: '', series: '', tags: '', genres: '', authors: '', narrators: '', keywords: '', languages: ''
});

const sortBy = ref('title');
const sortOrder = ref('asc');
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
    searchTerm.value = normalizeQueryValue(route.query.search);
    sortBy.value = normalizeQueryValue(route.query.sortBy) || 'title';
    sortOrder.value = normalizeQueryValue(route.query.sortOrder) || 'asc';
    advancedSearchTerms.value.year = normalizeQueryValue(route.query.year);
    advancedSearchTerms.value.isbn = normalizeQueryValue(route.query.isbn);
    advancedSearchTerms.value.asin = normalizeQueryValue(route.query.asin);
    advancedSearchTerms.value.publisher = normalizeQueryValue(route.query.publisher);
    advancedSearchTerms.value.series = normalizeQueryValue(route.query.series);
    advancedSearchTerms.value.tags = normalizeQueryValue(route.query.tags);
    advancedSearchTerms.value.genres = normalizeQueryValue(route.query.genres);
    advancedSearchTerms.value.authors = normalizeQueryValue(route.query.authors);
    advancedSearchTerms.value.narrators = normalizeQueryValue(route.query.narrators);

    showAdvanced.value = !AdvancedTermsAreEmpty(advancedSearchTerms.value);
}

function buildQueryString(page?: number) {
    var params = new URLSearchParams();

    params.append('view', 'summary');
    const pageNum = page ?? currentPage.value;
    params.append('page', pageNum.toString());
    params.append('count', itemsPerPage.value.toString());
    if (searchTerm.value.trim()) params.append('search', searchTerm.value.trim());
    if (sortBy.value) params.append('sortBy', sortBy.value);
    if (sortOrder.value) params.append('sortOrder', sortOrder.value);
    if (files.value) params.append('files', files.value);

    params = AddAdvancedTermsToQuery(advancedSearchTerms.value, params);

    const queryString = params.toString();
    return queryString ? `?${queryString}` : '';
}

function syncRouteQuery() {
    TrimAdvancedTerms(advancedSearchTerms.value);
    router.replace({
        query: {
            ...(searchTerm.value.trim() ? { search: searchTerm.value.trim() } : {}),
            ...(sortBy.value ? { sortBy: sortBy.value } : {}),
            ...(sortOrder.value ? { sortOrder: sortOrder.value } : {}),
            ...(advancedSearchTerms.value.year ? { publish_year: advancedSearchTerms.value.year } : {}),
            ...(advancedSearchTerms.value.isbn ? { isbn: advancedSearchTerms.value.isbn } : {}),
            ...(advancedSearchTerms.value.asin ? { asin: advancedSearchTerms.value.asin } : {}),
            ...(advancedSearchTerms.value.publisher ? { publisher: advancedSearchTerms.value.publisher } : {}),
            ...(advancedSearchTerms.value.series ? { series: advancedSearchTerms.value.series } : {}),
            ...(advancedSearchTerms.value.tags ? { tags: advancedSearchTerms.value.tags } : {}),
            ...(advancedSearchTerms.value.genres ? { genres: advancedSearchTerms.value.genres } : {}),
            ...(advancedSearchTerms.value.authors ? { authors: advancedSearchTerms.value.authors } : {}),
            ...(advancedSearchTerms.value.narrators ? { narrators: advancedSearchTerms.value.narrators } : {}),
        },
    });
}

function handleFetchData(data: any, append = false) {
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
        ? books.value?.length
        : data.items?.length;

    if (loadedCount >= data.results_count) {
        hasLoadedAll.value = true;
    }
}


async function fetchBooks(append = false) {
    if (isLoading.value || hasLoadedAll.value) return;

    isLoading.value = true;
    try {
        const queryString = buildQueryString();
        const resp = await api.get(`/api/books${queryString}`);
        handleFetchData(resp.data, append);

    } catch (error) {
        console.error('Error fetching books list:', error);
    } finally {
        isLoading.value = false;
    }
}

async function scanLibrary() {
    isLoading.value = true;
    try {
        const queryString = buildQueryString();
        const resp = await api.get(`/api/library/scan${queryString}`);
        const data = resp.data;
        handleFetchData(data.results);

        useNotificationsStore().notifySuccess('Scan Complete')

        const errors = data.errors;
        if (errors == undefined || errors.length === 0) {
            useNotificationsStore().notifySuccess('No errors')
        }
        else {
            useNotificationsStore().notifyError(`${errors.length} errors occured during the scan. See console for details`)
            console.error("Scan complete with " + errors.length + " errors \n\n" + errors)
        }

    } catch (error) {
        console.error('Error scanning library:', error);
        useNotificationsStore().notifyError('Something went wrong while scanning the library')
    } finally {
        isLoading.value = false;
    }
}

function searchBooksWithNewAdvancedTerms(terms: AdvancedSearchTerms) {
    advancedSearchTerms.value = terms;
    searchBooks();
}
function searchBooksWithNewSearch(terms: string) {
    searchTerm.value = terms;
    searchBooks();
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
    if (advancedSearchComponent.value) {
        advancedSearchComponent.value.reset();
        advancedSearchTerms.value = EmptyAdvancedSearchTerms;
    }
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
    <div class="vue-heading">
        <h2>Library</h2>
        <button @click="scanLibrary">Scan Library Folder</button>
    </div>

    <section>
        <div class="library-controls">
            <div class="search-panel">
                <SearchBar placeholder="Search title, subtitle, or description" :showAdvanced="showAdvanced"
                    :searchTerm="searchTerm" @searchTerm="searchBooksWithNewSearch($event)"
                    @showAdvanced="showAdvanced = $event" @reset="resetFilters">
                </SearchBar>

                <div v-if="showAdvanced" class="advanced-panel">
                    <AdvancedSearch ref="advancedSearchComponent" :fields="advancedSearchFields"
                        :searchTerms="advancedSearchTerms" @search="searchBooksWithNewAdvancedTerms($event)"
                        :reset="resetFilters"></AdvancedSearch>
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

.sort-row {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    justify-content: flex-start;
}

@media (min-width: 769px) {

    .search-row,
    .sort-row {
        grid-template-columns: 1fr;
    }
}

@media (max-width: 768px) {
    .advanced-panel {
        grid-template-columns: 1fr;
    }

    .sort-row {
        flex-direction: row;
        justify-content: center;
    }
}
</style>