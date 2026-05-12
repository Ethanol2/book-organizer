<script setup lang="ts">
import type { AdvancedSearchFields, AdvancedSearchTerms } from '@/types/advancedSearch';
import { ref } from 'vue'

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

const props = defineProps<{
    fields: AdvancedSearchFields;
    reset: () => void;
}>();
const emit = defineEmits(['search', 'reset'])

const search = () => {
    emit('search', buildSearchTerms());
}

function buildSearchTerms(): AdvancedSearchTerms {
    return {
        authors: authors.value,
        genres: genres.value,
        narrators: narrators.value,
        series: series.value,
        keywords: keywords.value,
        tags: tags.value,
        year: year.value,
        publisher: publisher.value,
        isbn: isbn.value,
        asin: asin.value,
        languages: languages.value
    };
}

defineExpose({reset});
function reset() {
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
    emit('reset');
}

</script>

<template>
    <div class="advanced-search">
        <label class="search-field">
            <span>Year</span>
            <input class="text-input" v-model="year" type="text" placeholder="Year" @keyup.enter="search" />
        </label>
        <label class="search-field">
            <span>ISBN</span>
            <input class="text-input" v-model="isbn" type="text" placeholder="ISBN" @keyup.enter="search" />
        </label>
        <label class="search-field">
            <span>ASIN</span>
            <input class="text-input" v-model="asin" type="text" placeholder="ASIN" @keyup.enter="search" />
        </label>
        <label class="search-field">
            <span>Publisher</span>
            <input class="text-input" v-model="publisher" type="text" placeholder="Publisher" @keyup.enter="search" />
        </label>
        <label class="search-field">
            <span>Series</span>
            <input class="text-input" v-model="series" type="text" placeholder="Series" @keyup.enter="search" />
        </label>
        <label class="search-field">
            <span>Tags</span>
            <input class="text-input" v-model="tags" type="text" placeholder="tag1, tag2" @keyup.enter="search" />
        </label>
        <label class="search-field">
            <span>Genres</span>
            <input class="text-input" v-model="genres" type="text" placeholder="genre1, genre2"
                @keyup.enter="search" />
        </label>
        <label class="search-field">
            <span>Authors</span>
            <input class="text-input" v-model="authors" type="text" placeholder="author1, author2"
                @keyup.enter="search" />
        </label>
        <label class="search-field">
            <span>Narrators</span>
            <input class="text-input" v-model="narrators" type="text" placeholder="narrator1, narrator2"
                @keyup.enter="search" />
        </label>
    </div>
</template>

<style scoped>
.advanced-search {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 1rem;
}

.search-field {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    font-size: 0.95rem;
}

.search-field span {
    font-size: 0.82rem;
}
</style>