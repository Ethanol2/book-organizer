<script setup lang="ts">
import { getBookCoverSrc, type Book } from '@/types/book';

const props = defineProps<{
    book: Book
}>();

</script>

<template>
    <div class="book">
        <div class="cover-wrapper">
            <img :src="getBookCoverSrc(book)" :alt="book.title + ', cover'" class="cover">
        </div>
        <div class="info">
            <h3>
                {{ book.title }}
            </h3>
            <small>
                {{ book.subtitle }}
            </small>
        </div>
    </div>
</template>

<style scoped>
.book {
    width: 180px;
    height: 320px;
    text-align: center;
    padding: 0.2rem;
}

.cover-wrapper {
    display: flex;
    width: 160px;
    height: 256px;
    text-align: center;
    margin-left: auto;
    margin-right: auto;
    border: 1px solid grey;
    box-shadow: 0 0 5px black;
}

.cover {
    width: 100%;
    object-fit: contain;
    display: flex;
    align-items: center;
    justify-content: center;
}

.info {
    padding-top: 0.5rem;
}
</style>