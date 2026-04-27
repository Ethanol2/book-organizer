<script setup lang="ts">
import { ref } from 'vue';
import { getBaseName, getDownloadName, getTimeAdded, type Download } from '@/types/download';
import ImportDownload from './ImportDownloadModal.vue';

const props = defineProps<{
  download: Download
  openModalFunc: (download: Download) => void
}>()

</script>

<template>
  <div class="folder">

    <div class="preview">

      <div class="cover-wrapper">
        <img :src="download.files.cover === null ? '' : download.files.cover" class="cover">
      </div>
      <div class="info">
        <h2> {{ getDownloadName(download) }} </h2>
        <p class="meta"> {{ getTimeAdded(download) }} </p>

        <button @click="openModalFunc(download)" class="import-button">Import</button>
      </div>
    </div>

    <div class="file-list">
      <div v-if="download.files.cover !== null" class="file-row">
        {{ getBaseName(download.files.cover) }}
      </div>
      <div
        v-if="download.files.text_files && download.files.text_files.length > 0" 
        class="file-row" 
        v-for="file in download.files.text_files" 
        :key="file"
      >
        {{ getBaseName(file) }}
      </div>
      <div 
        v-if="download.files.audio_files && download.files.audio_files.length > 0"
        class="file-row" 
        v-for="file in download.files.audio_files" 
        :key="file"
      >
        {{ getBaseName(file) }}
      </div>
    </div>

  </div>
</template>

<style scoped>

.folder {
  display: flex;
  border: 1px solid var(--color-border);
  border-radius: 12px;
  overflow: hidden;
  margin-bottom: 1rem;
  background: var(--color-background-soft);
}

.preview {
  display: flex;
  gap: 1rem;
  padding: 1rem;
  min-width: 70%;
  border-right: 1px solid var(--color-border);
}

.info {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  width: 100%;
}

.meta {
  color: var(--color-text-subtle);
  font-size: 0.8rem;
}

.import-button {
  margin-top: 0.5rem;
  background: var(--color-primary-blue);
  border: none;
  padding: 0.4rem 0.8rem;
  cursor: pointer;
  transition: background-color 0.2s;
}

.import-button:hover {
  background: var(--color-primary-blue-dark);
}

.file-list {
  flex: 1;
  padding: 1rem;
}

.file-row {
  padding: 0.4rem 0;
  border-bottom: 1px solid var(--color-border);
  font-size: 0.9rem;
}

.file-row:last-child {
  border-bottom: none;
}

.cover-wrapper {
  height: 100%;
  display: flex;
  text-align: center;
}

.cover {
  width: 100px;
  aspect-ratio: 1;
  object-fit: contain;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ===================================================================== */

.card {
  height: 140px;
  width: 100%;
  border: 1px solid var(--color-gray-700);
  border-radius: 6px;
  padding: 6px;
  margin-bottom: 1rem;
}

.card-content {
  display: flex;
  height: 100%;
}

.details {
  flex: 66%;
  flex: 1;
  margin-left: 1rem;
  position: relative;
}

h3 {
  font-size: 1.2rem;
  font-weight: 500;
  margin-bottom: 0.4rem;
  color: var(--color-heading);
}


.import-button:hover {
  background-color: var(--color-primary-blue-dark);
}
</style>