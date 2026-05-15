<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { useRoute, type LocationQueryValue } from 'vue-router';
import { AudibleRegion, metadataSearchFields, MetadataSource } from '@/types/metadata';
import { IsEmpty, librarySearchFields, type SearchParams, type SearchTerms } from '@/types/search';

const route = useRoute();

const searchTerm = ref('');
const showAdvanced = ref(false);

const year = ref('');
const isbn = ref('');
const asin = ref('');
const publisher = ref('');
const series = ref('');
const tags = ref('');
const genres = ref('');
const authors = ref('');
const narrators = ref('');
const languages = ref('');
const keywords = ref('');

const files = ref('all');
const showSort = ref(true);
const sortBy = ref('');
const showOrder = ref(true);
const sortOrder = ref('');

const metadataSource = ref<MetadataSource>(MetadataSource.OpenLibrary);
const audibleRegion = ref<AudibleRegion>(AudibleRegion.US);
const availableFields = ref<SearchParams>()

const props = defineProps<{
    placeholder?: string;
    metadata: boolean;
}>();

const emit = defineEmits(['search', 'reset']);

const search = () => {
    emit('search', buildSearchTerms());
};

const reset = () => {
    searchTerm.value = '';
    year.value = '';
    isbn.value = '';
    asin.value = '';
    publisher.value = '';
    series.value = '';
    tags.value = '';
    genres.value = '';
    authors.value = '';
    narrators.value = '';
    languages.value = '';
    keywords.value = '';
    files.value = 'all';
    sortBy.value = 'title';
    sortOrder.value = 'asc';
    emit('reset');
}

const toggleShowAdvanced = () => {
    showAdvanced.value = !showAdvanced.value;
}

function normalizeQueryValue(value: LocationQueryValue | LocationQueryValue[] | undefined): string {
    if (Array.isArray(value)) {
        return String(value[0] ?? '');
    }
    return String(value ?? '');
}

function resetFiltersFromRoute() {
    searchTerm.value = normalizeQueryValue(route.query.search);
    sortBy.value = normalizeQueryValue(route.query.sortBy) || '';
    sortOrder.value = normalizeQueryValue(route.query.sortOrder) || '';
    year.value = normalizeQueryValue(route.query.year);
    isbn.value = normalizeQueryValue(route.query.isbn);
    asin.value = normalizeQueryValue(route.query.asin);
    publisher.value = normalizeQueryValue(route.query.publisher);
    series.value = normalizeQueryValue(route.query.series);
    tags.value = normalizeQueryValue(route.query.tags);
    genres.value = normalizeQueryValue(route.query.genres);
    authors.value = normalizeQueryValue(route.query.authors);
    narrators.value = normalizeQueryValue(route.query.narrators);
}

function buildSearchTerms(): SearchTerms {
    return {
        search: searchTerm.value,
        year: year.value,
        isbn: isbn.value,
        asin: asin.value,
        publisher: publisher.value,
        series: series.value,
        tags: tags.value,
        genres: genres.value,
        authors: authors.value,
        narrators: narrators.value,
        languages: languages.value,
        keywords: keywords.value,
        files: files.value,
        sort: sortBy.value,
        order: sortOrder.value,
        metadataSource: props.metadata ? metadataSource.value : null,
        audibleRegion: props.metadata ? audibleRegion.value : null
    }
}

function hasAdvancedSearchTerms(): boolean {
    return authors.value !== ''
        || narrators.value !== ''
        || languages.value !== ''
        || keywords.value !== ''
        || isbn.value !== ''
        || asin.value !== ''
        || publisher.value !== ''
        || series.value !== ''
        || tags.value !== ''
        || genres.value !== ''
        || year.value !== '';
}

function changeSortValue(overwrite: boolean = false) {

    if (!availableFields.value) {
        return;
    }

    if (overwrite || (!overwrite && sortBy.value === '')) {
        const sortOptions = Object.keys(availableFields.value?.sortOptions)
        showSort.value = sortOptions.length > 0
        sortBy.value = showSort.value ? sortOptions[0] as string : '';
    }
    if (overwrite || (!overwrite && sortOrder.value === '')) {
        const orderOptions = Object.keys(availableFields.value?.orderOptions)
        showOrder.value = orderOptions.length > 0
        sortOrder.value = showOrder.value ? orderOptions[0] as string : '';
    }
}

onMounted(() => {
    resetFiltersFromRoute();
    showAdvanced.value = hasAdvancedSearchTerms();
    availableFields.value = props.metadata ? metadataSearchFields.get(metadataSource.value) : librarySearchFields;
    changeSortValue();
});

watch(() => metadataSource.value, () => {
    availableFields.value = props.metadata ? metadataSearchFields.get(metadataSource.value) : librarySearchFields;
    changeSortValue(true);
    if (IsEmpty(buildSearchTerms())) {
        return;
    }
    search();
});

watch(() => audibleRegion.value, () => {
    changeSortValue(true);
    if (IsEmpty(buildSearchTerms())) {
        return;
    }
    search();
});

</script>

<template>
    <div class="search-controls">

        <!-- Metadata Controls -->
        <div class="source-region" v-if="props.metadata">
            <select class="dropdown-select" v-model="metadataSource" aria-label="Metadata source">
                <option v-for="(type, value) in MetadataSource" :key="value" :value="type">
                    {{ type }}
                </option>
            </select>
            <select class="dropdown-select region" v-model="audibleRegion" aria-label="Audible Region"
                v-show="metadataSource == MetadataSource.Audible">
                <option v-for="(type, value) in AudibleRegion" :key="value" :value="type">
                    .{{ type }}
                </option>
            </select>
        </div>

        <!-- Search Bar -->
        <input class="text-input" v-model="searchTerm" type="search" :placeholder="props.placeholder || 'Search books'"
            aria-label="Search books" @keyup.enter="search" />
        <button class="green-button" type="button" @click="search()">Search</button>
        <div class="secondary-controls">
            <button class="toggle-button" type="button" @click="toggleShowAdvanced()">
                {{ showAdvanced ? 'Hide' : 'Advanced' }}
            </button>
            <button class="toggle-button" type="button" @click="reset()">Reset</button>
        </div>

    </div>

    <!-- Advanced Search Fields -->
    <div class="advanced-search" v-if="showAdvanced && availableFields">
        <label class="advanced-search-field" v-if="availableFields.year">
            <span>Year</span>
            <input class="text-input" v-model="year" type="text" placeholder="Year" @keyup.enter="search" />
        </label>
        <label class="advanced-search-field" v-if="availableFields.isbn">
            <span>ISBN</span>
            <input class="text-input" v-model="isbn" type="text" placeholder="ISBN" @keyup.enter="search" />
        </label>
        <label class="advanced-search-field" v-if="availableFields.asin">
            <span>ASIN</span>
            <input class="text-input" v-model="asin" type="text" placeholder="ASIN" @keyup.enter="search" />
        </label>
        <label class="advanced-search-field" v-if="availableFields.publisher">
            <span>Publisher</span>
            <input class="text-input" v-model="publisher" type="text" placeholder="Publisher" @keyup.enter="search" />
        </label>
        <label class="advanced-search-field" v-if="availableFields.series">
            <span>Series</span>
            <input class="text-input" v-model="series" type="text" placeholder="Series" @keyup.enter="search" />
        </label>
        <label class="advanced-search-field" v-if="availableFields.tags">
            <span>Tags</span>
            <input class="text-input" v-model="tags" type="text" placeholder="tag1, tag2" @keyup.enter="search" />
        </label>
        <label class="advanced-search-field" v-if="availableFields.genres">
            <span>Genres</span>
            <input class="text-input" v-model="genres" type="text" placeholder="genre1, genre2" @keyup.enter="search" />
        </label>
        <label class="advanced-search-field" v-if="availableFields.authors">
            <span>Authors</span>
            <input class="text-input" v-model="authors" type="text" placeholder="author1, author2"
                @keyup.enter="search" />
        </label>
        <label class="advanced-search-field" v-if="availableFields.narrators">
            <span>Narrators</span>
            <input class="text-input" v-model="narrators" type="text" placeholder="narrator1, narrator2"
                @keyup.enter="search" />
        </label>
        <label class="advanced-search-field" v-if="availableFields.languages">
            <span>Languages</span>
            <input class="text-input" v-model="languages" type="text" placeholder="language1, language2"
                @keyup.enter="search" />
        </label>
        <label class="advanced-search-field" v-if="availableFields.keywords">
            <span>Keywords</span>
            <input class="text-input" v-model="keywords" type="text" placeholder="keyword1, keyword2"
                @keyup.enter="search" />
        </label>
    </div>

    <!-- Sorting and Filtering Controls -->
    <div class="sort-row">
        <label v-if="availableFields?.files">
            <span>Files</span>
            <select class="dropdown-select" v-model="files" @change="search">
                <option value="all">All</option>
                <option value="with_files">Has Files</option>
                <option value="without_files">No Files</option>
            </select>
        </label>

        <label v-if="showSort && availableFields">
            <span>Sort by</span>
            <select class="dropdown-select" v-model="sortBy" @change="search">
                <option v-for="(value, term) in availableFields.sortOptions" :key="term" :value="term">{{ value }}
                </option>
            </select>
        </label>

        <label v-if="showOrder && availableFields">
            <span>Order</span>
            <select class="dropdown-select" v-model="sortOrder" @change="search">
                <option v-for="(value, term) in availableFields.orderOptions" :key="term" :value="term">{{ value }}
                </option>
            </select>
        </label>
    </div>

</template>

<style scoped>
.search-controls {
    display: flex;
    gap: 0.75rem;
}

.green-button {
    max-width: 100px;
}

.secondary-controls {
    min-height: 42px;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    gap: 0.75rem;
}

.source-region {
    display: flex;
    gap: 0.7rem;
    height: 100%;
    width: 500px;
    display: flex;
    gap: 0.7rem;
    height: 100%;
    width: 500px;
}

.dropdown-select.region {
    width: 80px;
    width: 80px;
}

.advanced-search {
    margin-top: 1rem;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 1rem;
}

.advanced-search-field {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    font-size: 0.95rem;
}

.advanced-search-field span {
    font-size: 0.82rem;
}

.sort-row {
    margin-top: 1rem;
    display: flex;
    gap: 0.75rem;
    align-items: center;
    justify-content: flex-start;
}

@media (max-width: 768px) {

    .source-region {
        width: 100%;
    }

    .search-controls {
        flex-direction: column;
    }

    .green-button {
        max-width: unset;
    }

    .advanced-search {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
    }

    .sort-row {
        justify-content: center;
    }
}
</style>