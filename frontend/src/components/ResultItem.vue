<script setup lang="ts">
// Result item component - displays a single search result as a clickable card

interface Category {
  id?: string
  name?: string
}

interface MetadataItem {
  title?: string | null
  subtitle?: string | null
  description?: string | null
  year?: number | null
  isbn?: string | null
  publisher?: string | null
  authors?: Category[] | null
  genres?: Category[] | null
  cover?: string | null
}

const props = defineProps<{
  item: MetadataItem
}>()

// Emit click event to parent for modal opening
const emit = defineEmits<{
  click: []
}>()

// Format author list for display
const formattedAuthors = (item: MetadataItem) => {
  if (!item.authors || item.authors.length === 0) {
    return 'Unknown author'
  }
  return item.authors.map(a => a.name || 'unknown').join(', ')
}
</script>

<template>
  <!-- Clickable result item card -->
  <li class="result-item" @click="emit('click')">
    <!-- Book cover or placeholder -->
    <div class="result-cover">
      <img v-if="item.cover" :src="item.cover" :alt="item.title || 'cover'" />
      <div v-else class="cover-placeholder"></div>
    </div>

    <!-- Book metadata details -->
    <div class="result-details">
      <h3>{{ item.title || 'Untitled' }}</h3>
      <p class="subtitle" v-if="item.subtitle">{{ item.subtitle }}</p>
      <p class="meta" v-if="item.authors"><strong>Author:</strong> {{ formattedAuthors(item) }}</p>
      <p class="publisher" v-if="item.year"><strong>Year:</strong> {{ item.year }}</p>
      <p class="publisher" v-if="item.isbn"><strong>ISBN:</strong> {{ item.isbn }}</p>
      <p class="publisher" v-if="item.publisher"><strong>Publisher:</strong> {{ item.publisher }}</p>
      <p class="description" v-if="item.description">{{ item.description }}</p>
    </div>
  </li>
</template>

<style scoped>
/* Individual result item - clickable card */
.result-item {
  border: 1px solid #ccc;
  border-radius: 8px;
  padding: 0.8rem;
  display: grid;
  grid-template-columns: 160px 1fr;
  gap: 0.8rem;
  cursor: pointer;
  list-style: none;
}

.result-item:hover {
  background: #f9f9f9;
}

/* Cover image or placeholder */
.cover-placeholder {
  width: 100%;
  height: 250px;
  background: #eee;
  border: 1px solid #ccc;
  border-radius: 4px;
  flex-shrink: 0;
}

.result-cover img {
  width: 100%;
  max-height: 250px;
  object-fit: contain;
  flex-shrink: 0;
}

/* Result details styling */
.result-details h3 {
  margin: 0;
  font-size: 1rem;
}

.result-details .subtitle {
  margin: 0.2rem 0;
  color: #333;
  font-size: 0.9rem;
}

.result-details .meta,
.result-details .publisher {
  margin: 0.2rem 0;
  font-size: 0.9rem;
  color: #555;
}

/* Truncated description - max 3 lines */
.description {
  margin-top: 0.45rem;
  line-height: 1.35;
  color: #444;
  font-size: 0.85rem;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  max-height: 4rem;
}

/* Responsive design for landscape orientation */
@media (orientation: landscape) and (max-height: 500px) {
  .result-item {
    padding: 0.4rem;
    grid-template-columns: 100px 1fr;
    gap: 0.4rem;
  }

  .cover-placeholder {
    height: 150px;
  }

  .result-cover img {
    max-height: 150px;
  }
}

/* Responsive design for smaller screens */
@media (max-width: 600px) {
  .result-item {
    grid-template-columns: 120px 1fr;
  }
}
</style>
