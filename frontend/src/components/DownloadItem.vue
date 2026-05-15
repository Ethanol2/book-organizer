<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { getBaseName, getDownloadName, getTimeAdded, type Download } from '@/types/download';

const showAudioFiles = ref(false)
const showBookFiles = ref(false)
const hasMetadata = ref(false)
const showOtherFiles = ref(false)

const props = defineProps<{
  download: Download
  openModalFunc: (download: Download) => void
}>();

onMounted(() => {
  hasMetadata.value = props.download.files.hasMetadata ?? false;
});

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

    <div class="file-section">
      <div v-if="download.files.text_files && download.files.text_files.length > 0">
        <button class="toggle-files" @click="showBookFiles = !showBookFiles">
          <span>View Files</span>
          <span class="count">{{download.files.text_files.length}}</span>
        </button>
        <div class="file-list">
          <div class="file-row" v-if="showBookFiles" v-for="file in download.files.text_files" :key="file">{{
            getBaseName(file) }}</div>
        </div>
      </div>
      <div v-if="download.files.audio_files && download.files.audio_files.length > 0" class="file-row"
        v-for="file in download.files.audio_files" :key="file">
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

.green-button {
  margin-top: 1rem;
  height: fit-content;
}

.file-list {
  flex: 1;
  padding: 1rem;
}

.file-section {
  margin-top: 1rem;
}

.toggle-files {
  width: 100%;

  display: flex;
  align-items: center;
  justify-content: space-between;

  background: var(--color-background-mute);

  border: 1px solid var(--color-border);

  border-radius: 10px;

  padding: 0.75rem 1rem;

  cursor: pointer;

  color: var(--color-text);

  font-size: 0.9rem;
  font-weight: 500;

  transition:
    background 0.15s ease,
    border-color 0.15s ease;
}

.toggle-files:hover {
  background: var(--color-nav-hover-bg);
  border-color: var(--color-border-hover);
}

.count {
  color: var(--color-text-subtle);
  font-size: 0.8rem;
}

.file-list {
  margin-top: 0.75rem;

  display: flex;
  flex-direction: column;

  gap: 0.35rem;
}

.file-row {
  background: var(--color-background);

  border: 1px solid var(--color-border);

  border-radius: 8px;

  padding: 0.65rem 0.85rem;

  font-size: 0.88rem;

  color: var(--color-text);

  transition: background 0.15s ease;
}

.file-row:hover {
  background: var(--color-background-mute);
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