<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import type { Download } from '@/types/download';
import DownloadItem from '@/components/DownloadItem.vue';
import ImportDownloadModal from '@/components/ImportDownloadModal.vue';
import { useNotificationsStore } from '@/stores/notifications';
import api from '@/services/api';
import type { BookParams } from '@/types/book';
import AddBookModal from '@/components/AddBookModal.vue';
import { AxiosError } from 'axios';

const downloads = ref<Download[]>([]);
const lastRefresh = ref<string>("Never refreshed");
const lastRefreshParams: { time: number, intervalId: number } = {
    time: -1,
    intervalId: 0
}
const showImportModal = ref(false);
const showAddModal = ref(false);
const newBookParams = ref<BookParams | null>(null);
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
        lastRefreshParams.time = Date.now();
        updateLastRefresh();

    } catch (error) {
        console.error("Error fetching downloads list:", error);
        useNotificationsStore().notifyError('Something went wrong while fetching downloads')
    }
    finally {
        loading.value = false;
        updateLastRefresh();
    }
}

async function fetchDownloadMetadata(): Promise<BookParams | null> {
    var params = null;
    try {
        const resp = await api.get("/api/downloads/" + selectedDownload.value.id + "/metadata");
        params = resp.data;
    } catch (err) {
        console.error("Error fetching download metadata:", err);
    }
    return params;
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

async function showModal(download: Download, usingMetadata?: boolean) {

    selectedDownload.value = download;

    if (usingMetadata) {
        const params = await fetchDownloadMetadata();
        if (params === null) {
            return;
        }
        newBookParams.value = params;
        showAddModal.value = true;
        return;
    }

    showImportModal.value = true;
}

async function importWithNewBook(params: BookParams) {

    var newBookId = '';
    try {
        const resp = await api.post("/api/books", params);
        useNotificationsStore().notifySuccess('Book added successfully!');
        newBookId = resp.data.id;
    }
    catch (err) {
        console.error('Add book error', err);

        if (err instanceof AxiosError) {
            useNotificationsStore().notifyError('Failed to add book: ' + err.response?.data.error);
        }
        else {
            useNotificationsStore().notifyError('Failed to add book');
        }
    }

    if (newBookId !== '') {

        await new Promise(resolve => setTimeout(resolve, 500));

        try {
            const resp = await api.post("/api/downloads/" + selectedDownload.value.id + "/associate", { book_id: newBookId, use_downloaded_cover: true });
            useNotificationsStore().notifySuccess('Download associated successfully!');
            downloads.value = downloads.value.filter(download => download.id !== selectedDownload.value.id);
        }
        catch (err) {
            console.error('Associate download error', err);
            useNotificationsStore().notifyError('Failed to associate download');
        }
    }

    showAddModal.value = false;

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
            <DownloadItem v-for="download in downloads" :key="download.id" :download="download"
                :openModalFunc="showModal" />
        </div>
    </section>
    <div>
        <AddBookModal v-model="showAddModal" :show="showAddModal" :isEditMode="false" :has-files="false"
            :params="newBookParams" @add-book="importWithNewBook" @close="showAddModal = false" />
        <ImportDownloadModal v-model="showImportModal" :download="selectedDownload" :show="showImportModal"
            :refreshFunc="fetchDownloads" />
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

@media (max-width: 410px) {
    .vue-heading {
        display: block;
        justify-content: left;
        gap: 0.5rem;
    }

    .refresh {
        justify-content: space-between;
        width: 100%;
    }
}
</style>