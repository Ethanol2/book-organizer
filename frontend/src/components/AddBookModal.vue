<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { BookParams } from '@/types/book'
import { getCategoriesString, getCategoriesArray, getSeriesArray, getSeriesString } from '@/types/book'

// Modal component for adding/editing books with form fields and submission
const titleInput = ref('')
const subtitleInput = ref('')
const descriptionInput = ref('')
const yearInput = ref('')
const isbnInput = ref('')
const publisherInput = ref('')
const coverInput = ref('')
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
  params: null,
})

const emit = defineEmits<{
  close: []
  'add-book': [BookParams]
}>()

function resetFormFields() {
  const params = props.params

  titleInput.value = params?.title ?? ''
  subtitleInput.value = params?.subtitle ?? ''
  descriptionInput.value = params?.description ?? ''
  yearInput.value = params?.year != null ? String(params.year) : ''
  isbnInput.value = params?.isbn ?? ''
  publisherInput.value = params?.publisher ?? ''
  coverInput.value = params?.cover ?? ''
  authorsInput.value = getCategoriesString(params?.authors ?? [])
  narratorsInput.value = getCategoriesString(params?.narrators ?? [])
  seriesInput.value = getSeriesString(params?.series ?? [])
  genresInput.value = getCategoriesString(params?.genres ?? [])
}

function handleSubmit() {
  const baseParams = props.params ? { ...props.params } : {}
  const yearValue = yearInput.value.trim()
  const params: BookParams = {
    ...baseParams,
    title: titleInput.value.trim() || null,
    subtitle: subtitleInput.value.trim() || null,
    description: descriptionInput.value.trim() || null,
    publisher: publisherInput.value.trim() || null,
    isbn: isbnInput.value.trim() || null,
    cover: coverInput.value.trim() || null,
    year: yearValue ? Number(yearValue) : null,
    authors: getCategoriesArray(authorsInput.value),
    narrators: getCategoriesArray(narratorsInput.value),
    series: getSeriesArray(seriesInput.value),
    genres: getCategoriesArray(genresInput.value),
  }

  if (params.year !== null && Number.isNaN(params.year)) {
    params.year = null
  }

  emit('add-book', params)
}

function handleOverlayClick() {
  emit('close')
}

function handleModalClick(e: Event) {
  e.stopPropagation()
}

const modalTitle = computed(() => (props.params ? 'Edit Book' : 'Add Book'))
const submitLabel = computed(() => (props.params ? 'Save Changes' : 'Add Book'))

watch(
  [() => props.params, () => props.show],
  ([params, visible]) => {
    if (visible) {
      resetFormFields()
    }
  },
  { immediate: true }
)
</script>

<template>
  <!-- Modal overlay - closes when clicked -->
  <div v-if="show" class="modal-overlay" @click="handleOverlayClick">
    <!-- Modal dialog - prevents closing from interior clicks -->
    <div class="modal" @click="handleModalClick">
      <h3>{{ modalTitle }}</h3>
      <form @submit.prevent="handleSubmit">
        <!-- Book metadata form fields -->
        <label>Title: <input v-model="titleInput" type="text" required /></label>
        <label>Subtitle: <input v-model="subtitleInput" type="text" /></label>
        <label>Series (comma-separated, use #1 for first in series): <input v-model="seriesInput" type="text" /></label>
        <label>Description: <textarea v-model="descriptionInput"></textarea></label>
        <label>Year: <input v-model="yearInput" type="number" /></label>
        <label>ISBN: <input v-model="isbnInput" type="text" /></label>
        <label>Publisher: <input v-model="publisherInput" type="text" /></label>
        <label>Authors (comma-separated): <input v-model="authorsInput" type="text" /></label>
        <label>Narrators (comma-separated): <input v-model="narratorsInput" type="text" /></label>
        <label>Genres (comma-separated): <input v-model="genresInput" type="text" /></label>
        <label>Cover URL: <input v-model="coverInput" type="text" /></label>

        <!-- Action buttons -->
        <div class="modal-buttons">
          <button type="button" @click="handleOverlayClick">Cancel</button>
          <button type="submit">{{ submitLabel }}</button>
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
