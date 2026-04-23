<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import type { Download } from '@/types/download';
import DownloadItem from '@/components/DownloadItem.vue';
import ImportDownloadModal from '@/components/ImportDownloadModal.vue';

const downloads = ref<Download[]>([]);
const lastRefresh = ref<string>("Never refreshed");
const lastRefreshParams: { time: number, intervalId: number } = {
    time: -1,
    intervalId: 0
}
const showImportModal = ref(false);
const selectedDownload = ref<Download>({
    id: "if this shows up something went wrong",
    files: {
        cover: null,
        text_files: null,
        audio_files: null
    },
    created_at: ""
});


async function fetchDownloads() {
    downloads.value = [];
    try {
        const resp = await fetch("/api/downloads");

        if (!resp.ok) {
            throw new Error(`HTTP error with status: ${resp.status}`);
        }

        downloads.value = await resp.json();
        lastRefreshParams.time = Date.now()
        updateLastRefresh()

    } catch (error) {
        console.error("Error fetching downloads list:", error)
    }
}

function updateLastRefresh() {
    if (lastRefreshParams.time <= 0) {
        return
    }
    let timeSince = (Date.now() - lastRefreshParams.time) / 1000
    if (timeSince / 60 > 1) {
        lastRefresh.value = `${Math.round(timeSince / 60)} minute(s) since last refresh`
        return
    }

    lastRefresh.value = `${Math.round(timeSince)} seconds since last refresh`
}

function showModal(download: Download) {
    showImportModal.value = true;
    selectedDownload.value = download;
}

onMounted(async () => {
    await fetchDownloads();
    lastRefreshParams.intervalId = setInterval(updateLastRefresh, 1000)
});

onUnmounted(() => {
    clearInterval(lastRefreshParams.intervalId)
})

</script>

<template>
    <section class="downloads-view">
        <header class="toolbar">
            <h2>Downloads</h2>
            <div class="refresh">
                <small>{{ lastRefresh }}</small>
                <button @click="fetchDownloads">Refresh</button>
            </div>
        </header>

        <div>
            <DownloadItem v-for="download in downloads" :key="download.id" :download="download" :openModalFunc="showModal" />
        </div>
    </section>
    <div>
        <ImportDownloadModal v-model="showImportModal" :download="selectedDownload" :modelShow="showImportModal" />
    </div>
</template>

<style scoped>

.downloads-view {
    display: block;
    overflow-y: auto;
    padding: 1rem;
    padding-bottom: 10rem;
    box-sizing: border-box;
}

.toolbar {
    display: flex;
    padding-bottom: 0.5rem;
    justify-content: space-between;
}

.refresh {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}
</style>