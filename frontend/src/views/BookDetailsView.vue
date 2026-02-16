<script setup lang="ts">
import type { Book } from '@/types/book';
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute()
const book = ref<Book>()

onMounted(async () => {
    try {
        const resp = await fetch("/api/books/" + route.params.id);
        if (!resp.ok) {
            throw new Error(`HTTP error with status: ${resp.status}`);
        }

        book.value = await resp.json();
    } catch (error) {
        console.error("Error fetching book details:", error)
    }
})

</script>

<template>
    <div v-if="book" class="book-details">
        <section class="book-header">
            <!-- Cover -->
            <div class="cover-wrapper">
                <img :src="book.files?.cover ?? `/media/metadata/${book.id}.jpg`" :alt="book.title" />
            </div>

            <!-- Info -->
            <div class="info">
                <h1>{{ book.title }}</h1>
                <h3 v-if="book.subtitle">{{ book.subtitle }}</h3>

                <div class="meta">
                    <p><strong>Year:</strong> {{ book.year }}</p>
                    <p><strong>Publisher:</strong> {{ book.publisher }}</p>
                    <p><strong>ISBN:</strong> {{ book.isbn }}</p>

                    <p v-if="book.authors?.length">
                        <strong>Authors:</strong>
                        {{book.authors.map(a => a.name).join(', ')}}
                    </p>

                    <p v-if="book.genres?.length">
                        <strong>Genres:</strong>
                        {{book.genres.map(g => g.name).join(', ')}}
                    </p>
                </div>
            </div>
        </section>

        <!-- File section -->
        <section class="files">
            <h2>Files</h2>

            <div class="file-group">
                <p><strong>Root:</strong> {{ book.files?.root ?? 'None' }}</p>

                <p>
                    <strong>Audio:</strong>
                    {{ book.files?.audio_files?.length ?? 0 }}
                </p>

                <p>
                    <strong>Text:</strong>
                    {{ book.files?.text_files?.length ?? 0 }}
                </p>
            </div>
        </section>
    </div>
</template>


<style scoped>
.book-details {
    padding: 1.5rem;
    overflow-y: auto;
}

/* HEADER LAYOUT */
.book-header {
    display: flex;
    gap: 1.5rem;
    margin-bottom: 2rem;
}

/* COVER */
.cover-wrapper {
    width: 180px;
    max-height: 270px;
    flex-shrink: 0;
    border: 1px solid var(--color-border);
    border-radius: 4px;
    overflow: hidden;
    background: var(--color-background-soft);
}

.cover-wrapper img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

/* INFO */
.info {
    flex: 1;
}

.info h1 {
    margin: 0;
    font-size: 1.8rem;
}

.info h3 {
    margin-top: 0.25rem;
    font-weight: 400;
    color: var(--color-text-soft);
}

/* META GRID */
.meta {
    margin-top: 1rem;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 0.5rem;
}

/* FILE SECTION */
.files {
    border-top: 1px solid var(--color-border);
    padding-top: 1rem;
}

.file-group {
    margin-top: 0.5rem;
}
</style>
