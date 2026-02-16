<script setup lang="ts">
import BookItem from '@/components/BookItem.vue';
import type { Book, BookSummary } from '@/types/book';
import { onMounted, ref } from 'vue';

const books = ref<BookSummary[]>([]);

async function fetchBooks() {
    books.value = [];
    try {
        const resp = await fetch('api/books?view=summary');
        if (!resp.ok) {
            throw new Error(`HTTP error with status: ${resp.status}`);
        }

        books.value = await resp.json();

    } catch (error) {
        console.error("Error fetching books list:", error)
    }
}

onMounted(async () => {
    await fetchBooks();
});

</script>

<template>
    <section>
        <div class="library">
            <BookItem v-for="book in books" :key="book.id" :book="book"></BookItem>
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
</style>