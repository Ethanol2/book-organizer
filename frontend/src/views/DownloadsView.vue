<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import type { Download } from '@/types/download';
import DownloadItem from '@/components/DownloadItem.vue';
import ImportDownloadModal from '@/components/ImportDownloadModal.vue';
import { useNotificationsStore } from '@/stores/notifications';
import api from '@/services/api';

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
        audio_files: null,
        has_metadata: null
    },
    created_at: ""
});

const loading = ref(false);


async function fetchDownloads() {
    downloads.value = [];
    try {
        loading.value = true;
        const resp = await api.get('/api/downloads');

        downloads.value = resp.data;
        lastRefreshParams.time = Date.now()
        updateLastRefresh()

    } catch (error) {
        console.error("Error fetching downloads list:", error)
        useNotificationsStore().notifyError('Something went wrong while fetching downloads')
    }
    finally {
        loading.value = false;
        updateLastRefresh()
    }
}

function updateLastRefresh() {
    if (loading.value) {
        lastRefresh.value = "Loading..."
        return
    }
    if (lastRefreshParams.time <= 0) {
        return
    }
    let timeSince = (Date.now() - lastRefreshParams.time) / 1000
    if (timeSince / 60 > 60) {
        lastRefresh.value = `${Math.round(timeSince / 60 / 60)} hour(s) since last refresh`
        return
    }
    else if (timeSince / 60 > 1) {
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
        <header class="vue-heading">
            <h2>Downloads</h2>
            <div class="refresh">
                <small>{{ lastRefresh }}</small>
                <button v-show="!loading" @click="fetchDownloads">Refresh</button>
            </div>
        </header>

        <div class="download-items">
            <DownloadItem v-for="download in downloads" :key="download.id" :download="download" :openModalFunc="showModal" />
        </div>
    </section>
    <div>
        <ImportDownloadModal v-model="showImportModal" :download="selectedDownload" :modelShow="showImportModal" :refreshFunc="fetchDownloads" />
    </div>
</template>

<style scoped>

.downloads-view {
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    padding-bottom: 10rem;
    box-sizing: border-box;
}

.download-items {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    width: 100%;
    max-width: 1200px;
    align-self: center;
}

.refresh {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}
</style>