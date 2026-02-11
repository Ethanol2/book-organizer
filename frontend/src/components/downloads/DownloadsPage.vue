<script setup lang="ts">
import { ref, onMounted } from 'vue'
import DownloadItem from './DownloadItem.vue';
import type { Download } from '@/types/download';

const downloads = ref<Download[]>([])

onMounted(async () => {
    try {
        const resp = await fetch("/api/downloads");

        if (!resp.ok) {
            throw new Error(`HTTP error with status: ${resp.status}`);
        }

        downloads.value = await resp.json();

    } catch (error) {
        console.error("Error fetching downloads list:", error)
    }
})

</script>

<template>
    <section>
        <DownloadItem v-for="download in downloads" :key="download.id" :download="download" />
    </section>
</template>