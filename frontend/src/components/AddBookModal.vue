<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { BookParams } from '@/types/book'
import { getCategoriesString, getCategoriesArray, getSeriesArray, getSeriesString } from '@/types/book'

enum deleteMode {
  BookAndFiles = "Delete book and files",
  BookOnly = "Delete book only",
  FilesOnly = "Delete files only"
}

// Modal component for adding/editing books with form fields and submission
const titleInput = ref('')
const subtitleInput = ref('')
const descriptionInput = ref('')
const yearInput = ref('')
const isbnInput = ref('')
const asinInput = ref('')
const publisherInput = ref('')
const coverInput = ref('')
const authorsInput = ref('')
const narratorsInput = ref('')
const seriesInput = ref('')
const genresInput = ref('')
const tagsInput = ref('')
const showDeleteConfirm = ref(false)
const deleteBookText = ref('Delete Book')
const deleteModeSelection = ref<deleteMode>(deleteMode.BookAndFiles)

interface ModalProps {
  show: boolean
  params: BookParams | null
  isEditMode?: boolean
  hasFiles?: boolean
}

const props = withDefaults(defineProps<ModalProps>(), {
  show: false,
  params: null,
  showDeleteButton: false,
  hasFiles: false
})

const emit = defineEmits<{
  close: []
  'add-book': [BookParams]
  'delete-book': [{ deleteBook: boolean, deleteFiles: boolean }]
}>()

function resetConfirmState() {
  showDeleteConfirm.value = false
  deleteModeSelection.value = deleteMode.BookAndFiles
}

function resetFormFields() {
  const params = props.params

  titleInput.value = params?.title ?? ''
  subtitleInput.value = params?.subtitle ?? ''
  descriptionInput.value = params?.description ?? ''
  yearInput.value = params?.year != null ? String(params.year) : ''
  isbnInput.value = params?.isbn ?? ''
  asinInput.value = params?.asin ?? ''
  publisherInput.value = params?.publisher ?? ''
  coverInput.value = params?.cover ?? ''
  authorsInput.value = getCategoriesString(params?.authors ?? [])
  narratorsInput.value = getCategoriesString(params?.narrators ?? [])
  seriesInput.value = getSeriesString(params?.series ?? [])
  genresInput.value = getCategoriesString(params?.genres ?? [])
  tagsInput.value = params?.tags ? params.tags.join(', ') : ''
  resetConfirmState()
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
    asin: asinInput.value.trim() || null,
    cover: coverInput.value.trim() || null,
    year: yearValue ? Number(yearValue) : null,
    authors: getCategoriesArray(authorsInput.value),
    narrators: getCategoriesArray(narratorsInput.value),
    series: getSeriesArray(seriesInput.value),
    genres: getCategoriesArray(genresInput.value),
    tags: tagsInput.value ? tagsInput.value.split(',').map(tag => tag.trim()).filter(tag => tag) : null,
  }

  if (params.year !== null && Number.isNaN(params.year)) {
    params.year = null
  }

  emit('add-book', params)
}

function handleOverlayClick() {
  resetConfirmState()
  emit('close')
}

function handleDeleteClick() {
  showDeleteConfirm.value = !showDeleteConfirm.value
  deleteBookText.value = showDeleteConfirm.value ? 'Nevermind' : 'Delete Book'
}

function confirmDelete() {
  emit('delete-book',
    {
      deleteBook: deleteModeSelection.value === deleteMode.BookAndFiles || deleteModeSelection.value === deleteMode.BookOnly,
      deleteFiles: deleteModeSelection.value === deleteMode.BookAndFiles || deleteModeSelection.value === deleteMode.FilesOnly
    }
  )
  resetConfirmState()
}

function handleModalClick(e: Event) {
  e.stopPropagation()
}

function getDeleteConfirmText(): string {

  if (!props.hasFiles) {
    return 'Are you sure you want to delete this book?'
  }

  switch (deleteModeSelection.value) {
    case deleteMode.BookAndFiles:
      return 'Are you sure you want to delete this book and its files?'
    case deleteMode.BookOnly:
      return 'Are you sure you want to delete this book?'
    case deleteMode.FilesOnly:
      return 'Are you sure you want to delete these files?'
    default:
      return 'Something went wrong'
  }
}

const modalTitle = computed(() => (props.isEditMode ? 'Edit Book' : 'Add Book'))

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
        <label>ASIN: <input v-model="asinInput" type="text" /></label>
        <label>Publisher: <input v-model="publisherInput" type="text" /></label>
        <label>Authors (comma-separated): <input v-model="authorsInput" type="text" /></label>
        <label>Narrators (comma-separated): <input v-model="narratorsInput" type="text" /></label>
        <label>Genres (comma-separated): <input v-model="genresInput" type="text" /></label>
        <label>Tags (comma-separated): <input v-model="tagsInput" type="text" /></label>
        <label>Cover URL: <input v-model="coverInput" type="text" /></label>

        <div v-if="coverInput != ''" class="cover">
          <img :src="coverInput" />
        </div>

        <!-- Action buttons -->
        <div class="modal-buttons">
          <button type="button" @click="handleOverlayClick">Cancel</button>
          <button type="submit">Save</button>
        </div>

        <div v-if="props.params && props.isEditMode" class="modal-delete-wrap">
          <button type="button" class="delete-book-button" @click="handleDeleteClick">{{ deleteBookText }}</button>

          <div v-if="showDeleteConfirm" class="delete-confirmation">
            <div class="delete-confirmation-row">
              <p class="delete-confirmation-copy">{{ getDeleteConfirmText() }}</p>
              <button type="button" class="confirm-delete-button" @click="confirmDelete">Delete</button>
            </div>
            <label class="delete-files-checkbox" v-if="props.hasFiles">
              <select class="search-select" v-model="deleteModeSelection" aria-label="Delete Mode">
                <option v-for="(type, value) in deleteMode" :key="value" :value="type">
                  {{ type }}
                </option>
              </select>
            </label>
          </div>
        </div>

      </form>
    </div>
  </div>
</template>

<style scoped>

.modal input,
.modal textarea {
  width: 100%;
  padding: 0.75rem 1rem;
  border-radius: 6px;
  font-size: 0.95rem;
  border: 1px solid var(--vt-c-divider-light-1);
  box-sizing: border-box;
  background: var(--vt-c-white);
  color: var(--vt-c-text-light-1);
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
  border: 1px solid var(--color-gray-600);
  border-radius: 4px;
  background: var(--color-gray-100);
  cursor: pointer;
  font-size: 0.95rem;
  transition: background-color 0.2s;
}

.modal-buttons button:hover {
  background: var(--color-gray-400);
}

/* Submit button styling */
.modal-buttons button[type="submit"] {
  background: var(--color-primary-green);
  color: var(--vt-c-text-light-1);
  border-color: var(--color-primary-green-dark);
}

.modal-buttons button[type="submit"]:hover {
  background: var(--color-primary-green-dark);
}

/* Cancel button styling */

.modal-buttons button {
  background: var(--color-warning-background);
  color: var(--vt-c-text-light-1);
  border-color: var(--color-warning-border);
}

.modal-buttons button:hover {
  background: var(--color-warning-border);
}

.modal-delete-wrap {
  margin-top: 1rem;
  border-top: 1px solid var(--color-gray-500);
  padding-top: 1rem;
}

.delete-book-button {
  padding: 0.5rem 1rem;
  border: 1px solid var(--color-error-border);
  border-radius: 4px;
  background: var(--color-error-background);
  color: white;
  cursor: pointer;
  font-size: 0.95rem;
  transition: background-color 0.2s;
}

.delete-book-button:hover {
  background: var(--color-error-background-dark);
}

.delete-confirmation {
  margin-top: 1rem;
  padding: 1rem;
  background: var(--color-warning-background);
  border: 1px solid var(--color-warning-border);
  border-radius: 6px;
}

.delete-confirmation-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.delete-confirmation-copy {
  margin: 0;
  color: var(--color-warning-text);
  font-weight: 600;
  text-align: left;
  flex: 1;
}

.delete-files-checkbox {
  display: inline-flex;
  align-items: center;
  gap: 1rem;
  font-weight: 500;
  margin-top: 1rem;
}

.delete-files-checkbox input[type="checkbox"] {
  width: auto;
  margin: 0;
  flex-shrink: 0;
}

.delete-book-button,
.confirm-delete-button {
  padding: 0.5rem 1rem;
  border-radius: 4px;
  color: white;
  cursor: pointer;
  font-size: 0.95rem;
  transition: background-color 0.2s;
}

.delete-book-button {
  border: 1px solid #d43f3a;
  background: #d9534f;
}

.delete-book-button:hover,
.confirm-delete-button:hover {
  background: #c9302c;
}

.confirm-delete-button {
  border: 1px solid #d43f3a;
  background: #d9534f;
}

.cover img{
  width: 100%;
  max-height: 250px;
  object-fit: contain;
  flex-shrink: 0;
  border-radius: 6px;
  background: var(--color-background);
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.08);
}
</style>
