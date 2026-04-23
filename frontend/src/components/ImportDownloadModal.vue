<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useNotificationsStore } from '@/stores/notifications';
import type { Download } from '@/types/download';
import type { BookSummary } from '@/types/book';
import { getDownloadName } from '@/types/download';
import { MetadataType } from '@/types/metadata';

const router = useRouter();
const notificationsStore = useNotificationsStore();

const props = defineProps<{
  download: Download;
  modelShow: boolean;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
}>();

const searchQuery = ref('');
const searchResults = ref<BookSummary[]>([]);
const isLoading = ref(false);
const replaceCover = ref(true);
const source = ref<MetadataType | string>('Library');

const downloadName = computed(() => getDownloadName(props.download));
const isAssociating = ref(false);

const associateDownload = async (book: BookSummary) => {
  isAssociating.value = true;
  try {
    const response = await fetch(`/api/downloads/${props.download.id}/associate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        book_id: book.id,
        replace_book_cover: replaceCover.value,
      }),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Failed to associate download with book');
    }

    const result = await response.json();
    notificationsStore.notifySuccess(`Successfully associated "${book.title}" with the download`);
    closeModal();
    await router.push({ name: 'book-details', params: { id: book.id } });
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Failed to associate download';
    notificationsStore.notifyError(message);
    closeModal();
  } finally {
    isAssociating.value = false;
  }
};

const closeModal = () => {
  emit('update:modelValue', false);
};

const performSearch = async () => {

  isLoading.value = true;
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

watch(() => props.modelShow, (newVal) => {
  if (newVal) {
    searchQuery.value = getDownloadName(props.download);
    replaceCover.value = props.download.files.cover != null;
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

      <div v-if="download.files.cover != null">
        <label>
          <input type="checkbox" v-model="replaceCover" />
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
        <button
          v-for="book in searchResults"
          :key="book.id"
          @click="associateDownload(book)"
          :disabled="isAssociating"
          class="result-item result-button"
        >
          <img
            :src="'/media/metadata/' + book.id + '.jpg'"
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

      <!-- Action buttons -->
      <div class="modal-buttons">

        <label class="search-field">
            <select class="search-select" v-model="source" @change="performSearch">
                <option value='Library'>Library</option>
                <option v-for="(type, value) in MetadataType" :key="value" :value="value">
                  {{ type }}
                </option>
            </select>
        </label>

        <button type="button" @click="closeModal">Cancel</button>
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
  background: white;
  padding: 2rem;
  border-radius: 8px;
  max-width: 600px;
  width: 90%;
  height: 800px;
  display: flex;
  flex-direction: column;
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
  border: 1px solid #ccc;
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
  color: #666;
  border: 1px solid #eee;
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
  border: 1px solid #eee;
  border-radius: 4px;
  margin-bottom: 1rem;
}

.result-item {
  display: flex;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid #eee;
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
  background-color: #f5f5f5;
}

.result-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.result-button:focus {
  outline: 2px solid #007bff;
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
  color: #666;
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

.download-cover {
  max-height: 150px;
  aspect-ratio: 1;
  object-fit: contain;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>