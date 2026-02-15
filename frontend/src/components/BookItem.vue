<script setup lang="ts">
import { getBookCoverSrc, type Book } from '@/types/book';

const props = defineProps<{
    book: Book
}>();

const cover = getBookCoverSrc(props.book)

</script>

<template>
    <div class="book">
        <div class="cover-wrapper">
            <img v-if="cover != ''" :src="cover" :alt="book.title + ', cover'" class="cover">
            <p v-else class="no-cover">{{ book.title + ', cover' }}</p>
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
    display: flex;
    flex-direction: column;
    width: 180px;
    height: 320px;
    text-align: center;
    padding: 0.2rem;
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
    height: 256px;
    align-items: center;
    border: 1px solid grey;
    box-shadow: 0 0 5px black;
}

.info {
    padding-top: 0.5rem;
}
</style>