<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import type { BookParams } from '@/types/book'
import { getCategoriesString, getCategoriesArray, getSeriesArray, getSeriesString } from '@/types/book'

enum deleteMode {
  BookAndFiles = "Delete book and files",
  BookOnly = "Delete book only",
  FilesOnly = "Delete files only"
};

// Modal component for adding/editing books with form fields and submission
const descriptionContent = ref('');
const titleInput = ref('');
const subtitleInput = ref('');
const yearInput = ref<number|null>();
const isbnInput = ref('');
const asinInput = ref('');
const publisherInput = ref('');
const coverInput = ref('');
const authorsInput = ref('');
var authorsStr: string = '';
const narratorsInput = ref('');
const seriesInput = ref('');
var seriesStr: string = '';
const genresInput = ref('');
const tagsInput = ref('');

var hasOddPath = false;
const moveFiles = ref(false);

const showDeleteConfirm = ref(false);
const deleteBookText = ref('Delete Book');
const deleteModeSelection = ref<deleteMode>(deleteMode.BookAndFiles);

interface ModalProps {
  show: boolean
  params: BookParams | null
  isEditMode?: boolean
  bookRoot?: string | null
}

const props = withDefaults(defineProps<ModalProps>(), {
  show: false,
  params: null,
  isEditMode: false,
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
  descriptionContent.value = params?.description ?? ''
  yearInput.value = params?.year != null ? params.year : null
  isbnInput.value = params?.isbn ?? ''
  asinInput.value = params?.asin ?? ''
  publisherInput.value = params?.publisher ?? ''
  coverInput.value = params?.cover ?? ''
  authorsInput.value = authorsStr = getCategoriesString(params?.authors ?? [])
  narratorsInput.value = getCategoriesString(params?.narrators ?? [])
  seriesInput.value = seriesStr = getSeriesString(params?.series ?? [])
  genresInput.value = getCategoriesString(params?.genres ?? [])
  tagsInput.value = params?.tags ? params.tags.join(', ') : ''
  resetConfirmState()
}

function handleSubmit() {
  const baseParams = props.params ? { ...props.params } : {}
  
  const yearValue = yearInput.value != null ? yearInput.value : 0;

  const params: BookParams = {
    ...baseParams,
    title: titleInput.value.trim(),
    subtitle: subtitleInput.value.trim(),
    description: descriptionContent.value,
    publisher: publisherInput.value.trim(),
    isbn: isbnInput.value.trim(),
    asin: asinInput.value.trim(),
    cover: coverInput.value.trim(),
    year: Number(yearValue),
    authors: getCategoriesArray(authorsInput.value),
    narrators: getCategoriesArray(narratorsInput.value),
    series: getSeriesArray(seriesInput.value),
    genres: getCategoriesArray(genresInput.value),
    tags: tagsInput.value ? tagsInput.value.split(',').map(tag => tag.trim()).filter(tag => tag) : null,
    move_files: props.isEditMode ? moveFiles.value : null
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

  if (!!props.bookRoot) {
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

function requireMoveFiles(): boolean {
  return !!props.bookRoot && (
    (authorsInput.value !== authorsStr) ||
    (seriesInput.value !== seriesStr) ||
    (titleInput.value !== props.params?.title) ||
    hasOddPath
  )
}
function getNewPath(): string {
  const authors = getCategoriesArray(authorsInput.value);
  const series = getSeriesArray(seriesInput.value);

  var titlePath = titleInput.value;
  var seriesPath = '';
  if (series.length > 0) {
    seriesPath = series[0]?.name + "/";

    if (series[0]?.index) {
      titlePath = series[0].index + " - " + titlePath
    }
  }

  var authorPath = 'Unknown/';
  if (authors.length > 0) {
    authorPath = authors[0]?.name + "/";
  }

  return "/" + authorPath + seriesPath + titlePath
}
function getOldPath(): string {
  return props.bookRoot == null? "" : props.bookRoot.replace("/media/library", "")
}

const modalTitle = computed(() => (props.isEditMode ? 'Edit Book' : 'Add Book'))

watch(
  [() => props.params, () => props.show],
  ([params, visible]) => {
    if (visible) {
      resetFormFields()
      if (props.isEditMode && (getNewPath() !== getOldPath())) {
        hasOddPath = true;
      } 
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
        <label>Description:
          <QuillEditor v-model:content="descriptionContent" contentType="html" theme="snow" />
        </label>
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

        <div v-if="requireMoveFiles()" class="delete-confirmation">
          <div class="delete-confirmation-row">            
            <div class="delete-confirmation-copy">
              <div class="path-box path-box--old">
                <strong>Old Path</strong>
                <span>{{ getOldPath() }}</span>
              </div>
              <div class="path-box path-box--new">
                <strong>New Path</strong>
                <span>{{ getNewPath() }}</span>
              </div>
            </div>
          </div>
          <label class="delete-files-checkbox">
              <input type="checkbox" v-model="moveFiles" />
              Move Files
            </label>
        </div>

        <!-- Action buttons -->
        <div class="modal-buttons">
          <button type="button" @click="handleOverlayClick">Cancel</button>
          <button type="submit">{{ props.isEditMode ? 'Save Changes' : 'Add Book' }}</button>
        </div>

        <div v-if="props.params && props.isEditMode" class="modal-delete-wrap">
          <button type="button" class="delete-button" @click="handleDeleteClick">{{ deleteBookText }}</button>

          <div v-if="showDeleteConfirm" class="delete-confirmation">
            <div class="delete-confirmation-row">
              <p class="delete-confirmation-copy">{{ getDeleteConfirmText() }}</p>
              <button type="button" class="confirm-delete-button" @click="confirmDelete">Delete</button>
            </div>
            <label class="delete-files-checkbox" v-if="!!props.bookRoot">
              <select class="dropdown-select" v-model="deleteModeSelection" aria-label="Delete Mode">
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

.delete-button {
  padding: 0.5rem 1rem;
  border: 1px solid var(--color-error-border);
  border-radius: 4px;
  background: var(--color-error-background);
  color: white;
  cursor: pointer;
  font-size: 0.95rem;
  transition: background-color 0.2s;
}

.delete-button:hover {
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
  gap: 0.5rem;
  font-weight: 500;
  margin-top: 1rem;
}

.delete-files-checkbox input[type="checkbox"] {
  width: auto;
  margin: 0;
  flex-shrink: 0;
}

.delete-confirmation-copy {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.path-box {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.75rem 0.9rem;
  border-radius: 6px;
  font-size: 0.9rem;
  color: inherit;
}

.path-box strong,
.path-box span {
  color: inherit;
}

.path-box span {
  word-break: break-word;
}

.path-box--old {
  background: rgba(212, 63, 58, 0.12);
  border: 1px solid rgba(212, 63, 58, 0.35);
}

.path-box--new {
  background: rgba(56, 161, 105, 0.12);
  border: 1px solid rgba(56, 161, 105, 0.35);
}

@media (min-width: 700px) {
  .delete-confirmation-copy {
    flex-direction: row;
  }

  .path-box {
    flex: 1;
  }
}

.cover img {
  width: 100%;
  max-height: 250px;
  object-fit: contain;
  flex-shrink: 0;
  border-radius: 6px;
  background: var(--color-background);
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.08);
}

:deep(.ql-container) {
  font-size: 0.95rem;
  border-radius: 0 0 6px 6px;
  border: 1px solid var(--vt-c-divider-light-1);
  background: var(--vt-c-white);
}

:deep(.ql-editor) {
  min-height: 100px;
  color: var(--vt-c-text-light-1);
  padding: 0.75rem 1rem;
}

:deep(.ql-toolbar) {
  background: var(--vt-c-white);
  border: 1px solid var(--vt-c-divider-light-1);
  border-radius: 6px 6px 0 0;
  border-bottom: 1px solid var(--vt-c-divider-light-1);
}

:deep(.ql-toolbar.ql-snow .ql-stroke) {
  stroke: var(--vt-c-text-light-1);
}

:deep(.ql-toolbar.ql-snow .ql-fill) {
  fill: var(--vt-c-text-light-1);
}

:deep(.ql-toolbar.ql-snow button:hover .ql-stroke) {
  stroke: var(--color-primary-green);
}

:deep(.ql-toolbar.ql-snow button:hover .ql-fill) {
  fill: var(--color-primary-green);
}

:deep(.ql-toolbar.ql-snow button.ql-active .ql-stroke) {
  stroke: var(--color-primary-green);
}

:deep(.ql-toolbar.ql-snow button.ql-active .ql-fill) {
  fill: var(--color-primary-green);
}

@media (max-width: 768px) {
  .modal {
    overflow: scroll;
  }
}
</style>
