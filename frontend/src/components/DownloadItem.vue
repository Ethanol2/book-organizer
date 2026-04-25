<script setup lang="ts">
import { ref } from 'vue';
import { getDownloadName, getTimeAdded, type Download } from '@/types/download';
import ImportDownload from './ImportDownloadModal.vue';

const props = defineProps<{
  download: Download
  openModalFunc: (download: Download) => void
}>()

</script>

<template>
  <div class="card">
    <div class="card-content">
      <div class="cover-wrapper">
        <img :src="download.files.cover === null ? '' : download.files.cover" alt="no cover found" class="cover">
      </div>
      <div class="details">
        <h3>
          {{ getDownloadName(download) }}
        </h3>
        <p>
          Text File Count: {{ download.files.text_files == null ? 0 : download.files.text_files.length }} <br>
          Audio File Count: {{ download.files.audio_files == null ? 0 : download.files.audio_files.length }} <br>
          {{ getTimeAdded(download) }}
        </p>
        <button @click="openModalFunc(download)" class="import-button">Import</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
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

.cover-wrapper {
  height: 100%;
  display: flex;
  text-align: center;
  border: 1px solid var(--color-gray-700);
}

.cover {
  height: 100%;
  aspect-ratio: 1;
  object-fit: contain;
  display: flex;
  align-items: center;
  justify-content: center;
}

h3 {
  font-size: 1.2rem;
  font-weight: 500;
  margin-bottom: 0.4rem;
  color: var(--color-heading);
}

.import-button {
  position: absolute;
  bottom: 0;
  right: 0;
  padding: 0.5rem 1rem;
  background-color: var(--color-primary-blue);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.import-button:hover {
  background-color: var(--color-primary-blue-dark);
}
</style>