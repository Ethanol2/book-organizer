<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useNotificationsStore } from '@/stores/notifications';
import type { Download } from '@/types/download';
import { postBook, type BookSummary, type BookParams, getCategoriesString } from '@/types/book';
import { getDownloadName } from '@/types/download';
import { getMetadataDetails, MetadataType, searchMetadataSource } from '@/types/metadata';

const router = useRouter();
const notificationsStore = useNotificationsStore();

const props = defineProps<{
  download: Download;
  modelShow: boolean;
  refreshFunc: () => void
}>();

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
}>();

const searchQuery = ref('');
const searchResults = ref<BookSummary[] | BookParams[]>([]);
const isLoading = ref(false);
const useDownloadedCover = ref(true);
const source = ref<MetadataType | string>('Library');

const downloadName = computed(() => getDownloadName(props.download));
const isImporting = ref(false);

const metadataSearchBufferTime = 500;
let waitingForInputStop = true;

const associateDownload = async (book: BookSummary | BookParams, isSummary: boolean) => {
  
  if (isImporting.value) {
    return;
  }
  isImporting.value = true;

  let id;
  var notifId;

  if (isSummary) 
  {
    id = (book as BookSummary).id;
  } 
  else
  {
    notifId = notificationsStore.notifyInfo(`Fetching book details from ${source.value}...`);
    const bookDetails = await getMetadataDetails(book as BookParams);    
    notificationsStore.removeNotification(notifId);
    if (bookDetails == null) {
      notificationsStore.notifyError(`Failed to fetch book details from ${source.value}`);
    }
    else {
      notificationsStore.notifySuccess(`Successfully fetched book details from ${source.value}`);
    }

    notifId = notificationsStore.notifyInfo('Adding book...');
    const fullBook = await postBook(bookDetails ?? book);
    notificationsStore.removeNotification(notifId);

    if (fullBook == null) {
      notificationsStore.notifyError('Failed to add book');
      isImporting.value = false;
      return
    }

    notificationsStore.notifySuccess('Successfully added book');
    id = fullBook.id;
  }

  notifId = notificationsStore.notifyInfo(`Associating "${book.title}" with the download...`);
  
  try {
    const response = await fetch(`/api/downloads/${props.download.id}/associate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        book_id: id,
        use_downloaded_cover: useDownloadedCover.value,
      }),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Failed to associate download with book');
    }

    notificationsStore.notifySuccess(`Successfully associated "${book.title}" with the download`);
    closeModal();
    await router.push({ name: 'book-details', params: { id: id } });
    
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Failed to associate download';
    notificationsStore.notifyError(message);
    closeModal();
    
  } finally {
    notificationsStore.removeNotification(notifId);
    isImporting.value = false;
  }
};

const closeModal = () => {
  props.refreshFunc();
  emit('update:modelValue', false);
};

const performSearch = async () => {

  waitingForInputStop = true;
  if (isLoading.value) return

  isLoading.value = true;
  
  // A timeout to prevent spamming endpoints
  while (waitingForInputStop) {
    waitingForInputStop = false;
    await Promise.resolve().then(() => {
      return new Promise((resolve) => setTimeout(resolve, metadataSearchBufferTime));
    });
  }

  if (source.value !== "Library")
  {
    if (searchQuery.value.trim() === "") {
      searchResults.value = [];
      isLoading.value = false;
      return;
    }

    const metadataResults = await searchMetadataSource({
      source: source.value as MetadataType,
      title: searchQuery.value,
      page: 1,
      pageLimit: 10,
    })

    if (metadataResults == null) {
      searchResults.value = [];
    } 
    else
    {
      searchResults.value = metadataResults.items;
    }

    isLoading.value = false;

    return;
  }

  try {
    const params = new URLSearchParams();
    params.append('view', 'summary');
    params.append('files', 'without_files');
    if (searchQuery.value.trim()) {
      params.append('search', searchQuery.value.trim());
    }

    const response = await fetch(`/api/books?${params.toString()}`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    searchResults.value = await response.json();
  } catch (error) {
    console.error('Error searching books:', error);
    searchResults.value = [];
  } finally {
    isLoading.value = false;
  }
};

function getCover(book: BookSummary | BookParams, isSummary: boolean): string {

  if (isSummary) {
    const summary = book as BookSummary; 
    return '/media/metadata/' + summary.id + '.jpg'
  }

  return book.cover ?? '';
}

watch(() => props.modelShow, (newVal) => {
  if (newVal) {
    searchQuery.value = getDownloadName(props.download);
    useDownloadedCover.value = props.download.files.cover == null || props.download.files.cover === '' ? false : useDownloadedCover.value;
    performSearch();
  }
});

</script>

<template>
  <!-- Modal overlay - closes when clicked -->
  <div v-if="modelShow" class="modal-overlay" @click="closeModal">
    <!-- Modal dialog - prevents closing from interior clicks -->
    <div class="modal" @click.stop>
      <h3>Import Download - <strong>{{ downloadName }}</strong></h3>

      <div v-if="download.files.cover != null && download.files.cover !== ''">
        <label>
          <input type="checkbox" v-model="useDownloadedCover" />
          Use this cover
        </label>
        <img :src="download.files.cover" alt="no cover found" class="download-cover">
      </div>

      <!-- Search bar -->
      <div class="search-section">
        <label>Select Book:</label>
        <input
          v-model="searchQuery"
          type="text"
          @input="performSearch"
          placeholder="Search for books..."
        />
      </div>

      <!-- Search results -->
      <div v-if="isLoading" class="loading">Searching...</div>
      <div v-else-if="searchResults == null" class="no-results">
        No books found matching "{{ searchQuery }}"
      </div>
      <div v-else-if="searchResults.length === 0 && searchQuery.trim()" class="no-results">
        No books found matching "{{ searchQuery }}"
      </div>
      <div v-else class="results-list">
        <div v-if="source === 'Library'">
        <button
            v-for="book in searchResults as BookSummary[]"
            :key="book.id"
            @click="associateDownload(book, true)"
            :disabled="isImporting"
            class="result-item result-button"
          >
            <img
              :src="getCover(book, true)"
              alt=""
              class="book-cover"
            />
            <div class="book-info">
              <h4>{{ book.title }}</h4>
              <p v-if="book.subtitle">{{ book.subtitle }}</p>
              <p v-if="book.authors.length > 0">
                By: {{ book.authors.map(a => a.name).join(', ') }}
              </p>
            </div>
          </button>
        </div>
        <div v-else>
          <button
            v-for="(book, index) in searchResults as BookParams[]"
            :key="index"
            @click="associateDownload(book, false)"
            :disabled="isImporting"
            class="result-item result-button"
          >
            <img
              :src="getCover(book, false)"
              alt=""
              class="book-cover"
            />
            <div class="book-info">
              <h4>{{ book.title }}</h4>
              <p v-if="book.subtitle">{{ book.subtitle }}</p>
              <p>
                By: {{ getCategoriesString(book.authors ?? []) }}
              </p>
            </div>
          </button>
        </div>
      </div>

      <!-- Action buttons -->
      <div class="modal-buttons">

        <label class="search-field">
            <select class="search-select" v-model="source" @change="performSearch">
                <option value='Library'>Library</option>
                <option v-for="(type, value) in MetadataType" :key="value" :value="type">
                  {{ type }}
                </option>
            </select>
        </label>

        <span v-if="source !== 'Library'" style="color: var(--vt-c-text-subtle);">Use the Add Book page for more control</span>
        <span v-else style="color: var(--vt-c-text-subtle);">Results don't include books with files</span>

        <button type="button" @click="closeModal">Cancel</button>
      </div>

      <!-- Loading overlay -->
      <div v-if="isImporting" class="import-loading-overlay">
        <div class="loading-content">
          <div class="spinner"></div>
          <p>Importing...</p>
        </div>
      </div>
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
  background: var(--color-background);
  padding: 2rem;
  border-radius: 8px;
  max-width: 600px;
  width: 90%;
  height: 800px;
  display: flex;
  flex-direction: column;
  position: relative;
}

.modal h3 {
  margin-top: 0;
  margin-bottom: 0.25rem;
}

.modal h3 strong {
  font-weight: 900;
}

/* Checkbox label */
.modal > label {
  margin-bottom: 1.5rem;
}
/* Search section */
.search-section {
  margin-bottom: 1rem;
}

.search-section label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.search-section input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid var(--color-gray-600);
  border-radius: 4px;
  box-sizing: border-box;
  font-family: inherit;
  font-size: 0.95rem;
}

/* Loading and no results */
.loading, .no-results {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  color: var(--vt-c-text-subtle);
  border: 1px solid var(--color-gray-300);
  border-radius: 4px;
  margin-bottom: 1rem;
}

.no-results {
  font-style: italic;
}

/* Results list */
.results-list {
  flex: 1;
  overflow-y: auto;
  border: 1px solid var(--color-gray-300);
  border-radius: 4px;
  margin-bottom: 1rem;
}

.result-item {
  display: flex;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid var(--color-gray-300);
  gap: 1rem;
}

.result-button {
  background: none;
  border: none;
  cursor: pointer;
  transition: background-color 0.2s;
  text-align: left;
  width: 100%;
}

.result-button:hover:not(:disabled) {
  background-color: var(--color-gray-150);
}

.result-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.result-button:focus {
  outline: 2px solid var(--color-primary-blue);
  outline-offset: -2px;
}

.book-info {
  flex: 1;
}

.book-info h4 {
  margin: 0 0 0.5rem 0;
  font-size: 1rem;
  font-weight: 600;
}

.book-info p {
  margin: 0.25rem 0;
  color: var(--color-gray-900);
  font-size: 0.9rem;
}

.book-cover {
  height: 100px;
  width: auto;
  aspect-ratio: 2/3;
  object-fit: cover;
  border-radius: 4px;
  flex-shrink: 0;
}

/* Modal action buttons */
.modal-buttons {
  display: flex;
  gap: 1rem;
  justify-content: space-between;
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

.download-cover {
  max-height: 150px;
  aspect-ratio: 1;
  object-fit: contain;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Loading overlay */
.import-loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  z-index: 2000;
}

.loading-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.spinner {
  border: 4px solid var(--color-gray-200);
  border-top: 4px solid var(--color-primary-blue);
  border-radius: 50%;
  width: 40px;
  height: 40px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
</style>