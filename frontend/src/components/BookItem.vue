<script setup lang="ts">
import { getAuthorsList, getBookCoverSrc, type Book, type BookSummary } from '@/types/book';

const props = defineProps<{
    book: BookSummary
}>();

const cover = props.book.cover == null ? "" : props.book.cover

</script>

<template>
    <RouterLink :to="`/books/${book.id}`" class="book">
        <div class="cover-wrapper">
            <img v-if="cover != ''" :src="cover" :alt="book.title" class="cover">
            <img v-else :src="'/media/metadata/' + book.id + '.jpg'" :alt="book.title" class="no-cover">
        </div>
        <div class="info">
            <h3>
                {{ book.title }}
            </h3>
            <small>
                {{ book.subtitle }}
                {{ getAuthorsList(book.authors) }}
            </small>
        </div>
    </RouterLink>
</template>

<style scoped>
.book {
    display: flex;
    flex-direction: column;
    width: 180px;
    height: 320px;
    text-align: center;
    padding: 0.2rem;
    text-decoration: none;
    color: inherit
}

.cover-wrapper {
    flex: 0 0 256px;
    display: flex;
    justify-content: center;
    align-items: center;
    text-align: center;
    justify-content: center;
    margin-left: auto;
    margin-right: auto;
}

.cover {
    width: 160px;
    max-height: 256px;
    border: 1px solid grey;
    box-shadow: 0 0 5px black;
}

.no-cover {
    display: flex;
    width: 160px;
    min-height: 160px;
    align-items: center;
    justify-content: center;
    border: 1px solid grey;
    box-shadow: 0 0 5px black;
}

.info {
    padding-top: 0.5rem;
}
</style>