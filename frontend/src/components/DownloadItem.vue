<script setup lang="ts">
import { getDownloadName, getTimeAdded, type Download } from '@/types/download';
import FilesView from './FilesView.vue';

const props = defineProps<{
  download: Download
  openModalFunc: (download: Download, usingMetadata?: boolean) => void
}>();

</script>

<template>
  <div class="folder">

    <div class="preview">

      <div class="cover-wrapper">
        <img v-if="download.files.cover !== null" :src="download.files.cover" class="cover">
        <span v-else class="cover missing"></span>
      </div>
      <div class="info">
        <div>
          <h2> {{ getDownloadName(download) }} </h2>
          <p class="meta"> {{ getTimeAdded(download) }} </p>
        </div>

        <div class="import-buttons">
          <button @click="openModalFunc(download)" class="green-button">Import</button>
          <button v-if="download.files.has_metadata" @click="openModalFunc(download, true)" class="purple-button">Import
            with Metadata</button>
        </div>
      </div>
    </div>

    <FilesView :files="props.download.files" />

    <div class="import-buttons mobile">
      <button @click="openModalFunc(download)" class="green-button">Import</button>
      <button v-if="download.files.has_metadata" @click="openModalFunc(download, true)" class="purple-button">Import
        with Metadata</button>
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

.purple-button,
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
  height: 100%;
  object-fit: contain;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cover.missing {
  background-color: transparent;
}

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

.import-buttons {
  display: flex;
  gap: 0.5rem;
}

.import-buttons.mobile {
  display: none;
}

@media (max-width: 903px) {

  .folder {
    flex-direction: column;
  }

  .import-buttons {
    display: none;
  }

  .preview {
    border-right: none;
  }

  .import-buttons.mobile {
    display: flex;
    gap: 0;
  }

  .purple-button,
  .green-button {
    height: unset;
  }

}
</style>