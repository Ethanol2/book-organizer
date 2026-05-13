<script setup lang="ts">
import BookItem from '@/components/LibraryItem.vue';
import type { BookSummary } from '@/types/book';
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useNotificationsStore } from '@/stores/notifications';
import api from '@/services/api';
import SearchControls from '@/components/SearchControls.vue';
import { AddSearchTermsToQuery, NewSearchTerms, type SearchTerms, TrimSearchTerms} from '@/types/search';

const router = useRouter();
const books = ref<BookSummary[]>([]);
const searchTerms = ref<SearchTerms>(NewSearchTerms());

// Pagination state
const currentPage = ref(1);
const itemsPerPage = ref(20);
const totalResults = ref(0);
const isLoading = ref(false);
const hasLoadedAll = ref(false);

function buildQueryString(page?: number) {
    var params = new URLSearchParams();

    params.append('view', 'summary');
    const pageNum = page ?? currentPage.value;
    params.append('page', pageNum.toString());
    params.append('count', itemsPerPage.value.toString());

    params = AddSearchTermsToQuery(searchTerms.value, params);

    const queryString = params.toString();
    return queryString ? `?${queryString}` : '';
}

function syncRouteQuery() {
    TrimSearchTerms(searchTerms.value);
    router.replace({
        query: {
            ...(searchTerms.value.search ? { search: searchTerms.value.search } : {}),
            ...(searchTerms.value.sort ? { sortBy: searchTerms.value.sort } : {}),
            ...(searchTerms.value.order ? { sortOrder: searchTerms.value.order } : {}),
            ...(searchTerms.value.year ? { publish_year: searchTerms.value.year } : {}),
            ...(searchTerms.value.isbn ? { isbn: searchTerms.value.isbn } : {}),
            ...(searchTerms.value.asin ? { asin: searchTerms.value.asin } : {}),
            ...(searchTerms.value.publisher ? { publisher: searchTerms.value.publisher } : {}),
            ...(searchTerms.value.series ? { series: searchTerms.value.series } : {}),
            ...(searchTerms.value.tags ? { tags: searchTerms.value.tags } : {}),
            ...(searchTerms.value.genres ? { genres: searchTerms.value.genres } : {}),
            ...(searchTerms.value.authors ? { authors: searchTerms.value.authors } : {}),
            ...(searchTerms.value.narrators ? { narrators: searchTerms.value.narrators } : {}),
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
function searchBooksWithNewSearch(terms: SearchTerms) {
    searchTerms.value = terms;
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

onMounted(async () => {
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
        <SearchControls placeholder="Search title, subtitle, or description" :metadata="false"
            @search="searchBooksWithNewSearch($event)" @reset="searchBooksWithNewSearch(NewSearchTerms())">
        </SearchControls>

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
    margin-top: 2rem;
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

</style>