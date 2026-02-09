<script setup lang="ts">
import type { Download } from '@/types/download';

const props = defineProps<{
    download: Download
}>()

</script>

<template>
  <div class="card">
    <div>
        <div class="cover">
            <div v-if="download.files.cover !== null">
                <img :src="download.files.cover">
            </div>
        </div>
        <div class="details">
            <h3>
                {{ download.files.root }}
            </h3>
            <p>
                Text File Count: {{ download.files.text_files == null ? 0 : download.files.text_files.length }} <br>
                Audio File Count: {{ download.files.audio_files == null ? 0 : download.files.audio_files.length }} <br>
                Added {{ download.created_at }}
            </p>
        </div>
    </div>
  </div>
</template>

<style scoped>
.card {
  display: flex;
  margin-top: 2rem;
}

.cover {
  flex: 33%;
  padding: 2rem;
  position: relative;
}

img {
  object-fit: cover;
  aspect-ratio: 1;
  overflow: hidden;
  width: 100%;
  height: 100%;
}

.details {
  flex: 66%;
  flex: 1;
  margin-left: 1rem;
  position: relative;
}

i {
  display: flex;
  place-items: center;
  place-content: center;
  width: 32px;
  height: 32px;

  color: var(--color-text);
}

h3 {
  font-size: 1.2rem;
  font-weight: 500;
  margin-bottom: 0.4rem;
  color: var(--color-heading);
}
</style>