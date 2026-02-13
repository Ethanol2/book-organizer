<script setup lang="ts">
import { getDownloadName, getTimeAdded, type Download } from '@/types/download';

const props = defineProps<{
  download: Download
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
      </div>
    </div>
  </div>
</template>

<style scoped>
.card {
  height: 140px;
  width: 100%;
  border: 1px solid grey;
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
  border: 1px solid grey;
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
</style>