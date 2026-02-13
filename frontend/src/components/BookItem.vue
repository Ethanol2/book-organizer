<script setup lang="ts">
import { getBookCoverSrc, type Book } from '@/types/book';

const props = defineProps<{
    book: Book
}>();

</script>

<template>
    <div class="book">
        <div class="cover-wrapper">
            <img :src="getBookCoverSrc(book)" :alt="book.title + ', cover'">
        </div>
        <h3>
            {{ book.title }}
        </h3>
        <small>
            {{ book.subtitle }}
        </small>
    </div>
</template>