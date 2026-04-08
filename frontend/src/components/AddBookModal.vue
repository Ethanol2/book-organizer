<script setup lang="ts">
import { ref, watch } from 'vue'
import type { BookParams } from '@/types/book'
import { getCategoriesString, getCategoriesArray, getSeriesArray, getSeriesString } from '@/types/book'
// Modal component for adding/editing books with form fields and submission

const authorsInput = ref('')
const narratorsInput = ref('')
const seriesInput = ref('')
const genresInput = ref('')

interface ModalProps {
  show: boolean
  params: BookParams | null
}

const props = withDefaults(defineProps<ModalProps>(), {
  show: false,
    params: null
})

// Emits for closing modal and submitting book data
const emit = defineEmits<{
  close: []
  'update:title': [value: string]
  'update:subtitle': [value: string]
  'update:description': [value: string]
  'update:year': [value: string]
  'update:isbn': [value: string]
  'update:publisher': [value: string]
  'update:authors': [value: string]
    'update:narrators': [value: string]
    'update:series': [value: string]
  'update:genres': [value: string]
  'update:cover': [value: string]
  'add-book': [bookData: {
    title: string | null
    subtitle: string | null
    description: string | null
    year: number | null
    isbn: string | null
    publisher: string | null
    authors: Array<{ name: string }> | null
    genres: Array<{ name: string }> | null
    cover: string | null
  }]
}>()

// Handle form submission
function handleSubmit() {
  if (props.params) {
    props.params.authors = getCategoriesArray(authorsInput.value)
    props.params.narrators = getCategoriesArray(narratorsInput.value)
    props.params.series = getSeriesArray(seriesInput.value)
    props.params.genres = getCategoriesArray(genresInput.value)
  }

  const bookData = {
    title: props.params?.title || null,
    subtitle: props.params?.subtitle || null,
    description: props.params?.description || null,
    year: props.params?.year || null,
    isbn: props.params?.isbn || null,
    publisher: props.params?.publisher || null,
    authors: props.params?.authors || null,
    narrators: props.params?.narrators || null,
    series: props.params?.series || null,
    genres: props.params?.genres || null,
    cover: props.params?.cover || null,
  }
  emit('add-book', bookData)
}

// Close modal when clicking overlay
function handleOverlayClick() {
  emit('close')
}

// Prevent closing when clicking inside modal
function handleModalClick(e: Event) {
  e.stopPropagation()
}

watch(
  () => props.show,
  (visible) => {
    if (!visible || !props.params) {
      return
    }

    authorsInput.value = getCategoriesString(props.params.authors ?? [])
    narratorsInput.value = getCategoriesString(props.params.narrators ?? [])
    seriesInput.value = getSeriesString(props.params.series ?? [])
    genresInput.value = getCategoriesString(props.params.genres ?? [])
  },
  { immediate: true }
)
</script>

<template>
  <!-- Modal overlay - closes when clicked -->
  <div v-if="show" class="modal-overlay" @click="handleOverlayClick">
    <!-- Modal dialog - prevents closing from interior clicks -->
    <div class="modal" @click="handleModalClick">
      <h3>Add Book</h3>
      <form @submit.prevent="handleSubmit">
        <!-- Book metadata form fields -->
        <label>Title: <input :value="params?.title" @input="emit('update:title', ($event.target as HTMLInputElement).value)" type="text" required /></label>
        <label>Subtitle: <input :value="params?.subtitle" @input="emit('update:subtitle', ($event.target as HTMLInputElement).value)" type="text" /></label>
        <label>Series (comma-separated, use #1 for first in series): <input v-model="seriesInput" type="text" /></label>
        <label>Description: <textarea :value="params?.description" @input="emit('update:description', ($event.target as HTMLTextAreaElement).value)"></textarea></label>
        <label>Year: <input :value="params?.year" @input="emit('update:year', ($event.target as HTMLInputElement).value)" type="number" /></label>
        <label>ISBN: <input :value="params?.isbn" @input="emit('update:isbn', ($event.target as HTMLInputElement).value)" type="text" /></label>
        <label>Publisher: <input :value="params?.publisher" @input="emit('update:publisher', ($event.target as HTMLInputElement).value)" type="text" /></label>
        <label>Authors (comma-separated): <input v-model="authorsInput" type="text" /></label>
        <label>Narrators (comma-separated): <input v-model="narratorsInput" type="text" /></label>
        <label>Genres (comma-separated): <input v-model="genresInput" type="text" /></label>
        <label>Cover URL: <input :value="params?.cover" @input="emit('update:cover', ($event.target as HTMLInputElement).value)" type="text" /></label>

        <!-- Action buttons -->
        <div class="modal-buttons">
          <button type="button" @click="handleOverlayClick">Cancel</button>
          <button type="submit">Add Book</button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
/* Modal overlay background */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

/* Modal dialog container */
.modal {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  max-width: 600px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}

.modal h3 {
  margin-top: 0;
  margin-bottom: 1rem;
}

/* Form labels and inputs */
.modal label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.modal input,
.modal textarea {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  margin-bottom: 1rem;
  box-sizing: border-box;
  font-family: inherit;
  font-size: 0.95rem;
}

.modal textarea {
  height: 100px;
  resize: vertical;
}

/* Modal action buttons */
.modal-buttons {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}

.modal-buttons button {
  padding: 0.5rem 1rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  background: #f9f9f9;
  cursor: pointer;
  font-size: 0.95rem;
  transition: background-color 0.2s;
}

.modal-buttons button:hover {
  background: #e9e9e9;
}

/* Submit button styling */
.modal-buttons button[type="submit"] {
  background: #4CAF50;
  color: white;
  border-color: #45a049;
}

.modal-buttons button[type="submit"]:hover {
  background: #45a049;
}
</style>
