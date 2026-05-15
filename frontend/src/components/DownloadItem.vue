<script setup lang="ts">
import { getDownloadName, getTimeAdded, type Download } from '@/types/download';
import FilesView from './FilesView.vue';

const props = defineProps<{
  download: Download
  openModalFunc: (download: Download) => void
}>();

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

        <button @click="openModalFunc(download)" class="green-button">Import</button>
      </div>
    </div>

    <FilesView :files="props.download.files" />

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

.green-button {
  margin-top: 1rem;
  height: fit-content;
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

</style>