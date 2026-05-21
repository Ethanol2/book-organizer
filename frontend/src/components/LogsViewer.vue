<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import api from '../services/api'

interface Log {
    time: string
    level: string
    msg: string
    error?: string | null
}

const waitingForStart = ref(true)
const counts = [25, 50, 100, 200]
const count = ref<number>(25)
const logs = ref<Log[]>([])
const loading = ref(false)
const errorMsg = ref<string | null>(null)
const offset = ref(0)
const noMore = ref(false)

const emit = defineEmits(['scrollToTop'])

async function fetchLogs(reset = true) {
    loading.value = true
    errorMsg.value = null
    try {
        const resp = await api.get('/api/logs', { params: { count: count.value, offset: offset.value } })
        const fetched: Log[] = resp.data?.logs ?? []

        const now = Date.now();
        fetched.reverse();
        fetched.forEach((log) => {
            const time = new Date(log.time);
            var diff = now - time.getTime();
            if (diff < 1000) {
                log.time = "Just now";
                return;
            }
            diff /= 1000;
            if (diff < 60) {
                log.time = `${Math.round(diff)} seconds ago`;
                return;
            }
            diff /= 60;
            if (diff < 60) {
                log.time = `${Math.round(diff)} minutes ago`;
                return;
            }
            log.time = time.getFullYear() + "-" + (time.getMonth() + 1) + "-" + time.getDate() + " " + time.getHours() + ":" + time.getMinutes()
        });

        if (reset) {
            logs.value = fetched
            offset.value = 0
        } else {
            logs.value = logs.value.concat(fetched)
        }

        noMore.value = fetched.length < count.value
    } catch (err) {
        console.error(err)
        errorMsg.value = 'Failed to load logs'
    } finally {
        loading.value = false
    }
}

function refresh() {
    waitingForStart.value = false
    offset.value = 0
    fetchLogs(true)
}

async function loadOlder() {
    offset.value += count.value
    await fetchLogs(false)
}

function backToTop() {
    document.getElementById('logs')?.scrollIntoView({ behavior: 'smooth' })
}

watch(count, () => {
    if (waitingForStart.value) return
    refresh();
});

</script>

<template>
    <div class="logs-viewer">
        <div class="controls">
            <div v-if="waitingForStart">
                <button class="green-button" @click="refresh" :disabled="loading">View Logs</button>
            </div>
            <div v-else>
                <button class="green-button" @click="refresh" :disabled="loading">Refresh</button>
            </div>

            <div class="lines-select">
                <select v-model="count" class="dropdown-select">
                    <option v-for="n in counts" :key="n" :value="n">{{ n }}</option>
                </select>
                Lines
            </div>
        </div>

        <div v-if="errorMsg" class="error" style="margin:0.5rem 0;color:var(--color-danger);">{{ errorMsg }}</div>

        <ul v-if="logs.length > 0" class="logs-list">
            <li v-for="(log, idx) in logs" :key="idx" class="log-row">
                <div v-if="log.error" class="error-text">Error: {{ log.error }}</div>
                <div v-else class="log-main">
                    <span class="time">{{ log.time }}</span>
                    <span :class="['msg', log.level.toLowerCase()]">{{ log.msg }}</span>
                </div>
            </li>
            <span v-if="noMore" class="last-msg">End of Logs</span>
        </ul>

        <div v-if="!waitingForStart">
            <div v-if="loading" style="text-align:center;margin-top:0.5rem">Loading…</div>
            <div v-else-if="logs.length === 0" style="text-align:center;margin-top:0.5rem">No logs available</div>

            <div class="controls">
                <button class="purple-button" @click="loadOlder" :disabled="loading || noMore">Load older</button>
                <button class="purple-button" @click="backToTop">Back to Top</button>
            </div>
        </div>
    </div>
</template>

<style scoped>
.logs-viewer {
    padding: 1rem;
    border-radius: 8px;
    background: var(--color-surface);
}

.controls {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    margin-bottom: 0.5rem
}

.controls>div {
    display: flex;
    align-items: center;
    width: 100%;
    gap: 0.5rem
}

.lines-select {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    max-width: 300px
}

.logs-list {
    list-style: none;
    padding: 1rem;
    margin-bottom: 1rem;
    overflow: auto;
    border: 1px solid var(--color-border);
    border-radius: 6px;
}

.log-row {
    padding: 0.6rem 0;
    border-bottom: 1px solid rgba(0, 0, 0, 0.04);
}

.log-main {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    flex-wrap: wrap
}

.time {
    color: var(--color-muted);
    font-size: 0.9rem;
    width: 80px
}

.level-badge {
    padding: 0.15rem 0.5rem;
    border-radius: 6px;
    font-weight: 600;
    font-size: 0.75rem;
    color: white
}

.msg {
    flex: 1;
}

.msg.debug {
    color: #7e57c2
}

.msg.info {
    color: var(--color-text)
}

.msg.warn {
    color: #ffb300
}

.msg.error {
    color: #e53935
}

.last-msg {
    display: flex;
    width: 100%;
    justify-content: center;
    align-content: center;
}

.error-text {
    color: var(--color-danger);
    font-size: 0.85rem;
    margin-top: 0.25rem
}

@media (max-width: 768px) {
    .controls {
        flex-direction: column;
    }
}
</style>
