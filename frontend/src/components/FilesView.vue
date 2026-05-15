<script setup lang="ts">
import type { BookFiles } from '@/types/book';
import { getBaseName } from '@/types/download';
import { ref } from 'vue';

const showBookFiles = ref(false);
const showAudioFiles = ref(false);
const showOtherFiles = ref(false);

const props = defineProps<{
    files: BookFiles
}>();

</script>

<template>
    <div class="files-view">

        <!-- Metadata -->
        <div class="file-header metadata" v-if="props.files.has_metadata">
            <svg viewBox="0 0 123 123">
                <path fill-rule="evenodd" clip-rule="evenodd"
                    d="M61.44,0c33.926,0,61.44,27.514,61.44,61.44c0,33.926-27.514,61.439-61.44,61.439 C27.513,122.88,0,95.366,0,61.44C0,27.514,27.513,0,61.44,0L61.44,0z M79.42,98.215H43.46v-6.053h6.757v-36.96H43.46v-4.816h16.808 c4.245,0,8.422-0.51,12.549-1.551v43.328h6.604V98.215L79.42,98.215z M63.859,21.078c2.785,0,4.975,0.805,6.571,2.396 c1.579,1.59,2.377,3.771,2.377,6.581c0,2.848-1.358,5.381-4.093,7.601c-2.751,2.22-5.941,3.338-9.577,3.338 c-2.733,0-4.905-0.765-6.569-2.297c-1.665-1.551-2.497-3.556-2.497-6.05c0-3.143,1.358-5.853,4.059-8.152 C56.83,22.219,60.072,21.078,63.859,21.078L63.859,21.078z" />
            </svg>
            Metadata
        </div>

        <!-- Book Files -->
        <div v-if="props.files.text_files && props.files.text_files.length > 0">
            <div v-if="props.files.text_files.length > 1">
                <span class="file-header" @click="showBookFiles = !showBookFiles">
                    <svg viewBox="0 0 123 123">
                        <path fill-rule="evenodd" clip-rule="evenodd"
                            d="M17.16 0h82.72a3.32 3.32 0 013.31 3.31v92.32c-.15 2.58-3.48 2.64-7.08 2.48H15.94c-4.98 0-9.05 4.07-9.05 9.05s4.07 9.05 9.05 9.05h80.17v-9.63h7.08v12.24c0 2.23-1.82 4.05-4.05 4.05H16.29C7.33 122.88 0 115.55 0 106.59V17.16C0 7.72 7.72 0 17.16 0zm3.19 13.4h2.86c1.46 0 2.66.97 2.66 2.15v67.47c0 1.18-1.2 2.15-2.66 2.15h-2.86c-1.46 0-2.66-.97-2.66-2.15V15.55c.01-1.19 1.2-2.15 2.66-2.15z" />
                    </svg>
                    {{ props.files.text_files.length }} eBook Files
                </span>
            </div>
            <div v-else>
                <span class="file-header no-click">
                    <svg viewBox="0 0 123 123">
                        <path fill-rule="evenodd" clip-rule="evenodd"
                            d="M17.16 0h82.72a3.32 3.32 0 013.31 3.31v92.32c-.15 2.58-3.48 2.64-7.08 2.48H15.94c-4.98 0-9.05 4.07-9.05 9.05s4.07 9.05 9.05 9.05h80.17v-9.63h7.08v12.24c0 2.23-1.82 4.05-4.05 4.05H16.29C7.33 122.88 0 115.55 0 106.59V17.16C0 7.72 7.72 0 17.16 0zm3.19 13.4h2.86c1.46 0 2.66.97 2.66 2.15v67.47c0 1.18-1.2 2.15-2.66 2.15h-2.86c-1.46 0-2.66-.97-2.66-2.15V15.55c.01-1.19 1.2-2.15 2.66-2.15z" />
                    </svg>
                    {{ getBaseName(props.files.text_files[0]) }}
                </span>
            </div>
            <div v-if="showBookFiles">
                <div v-if="props.files.text_files && props.files.text_files.length > 0" class="file-row"
                    v-for="file in props.files.text_files" :key="file">
                    {{ getBaseName(file) }}
                </div>
            </div>
        </div>

        <!-- Audio Files -->
        <div v-if="props.files.audio_files && props.files.audio_files.length > 0">
            <div v-if="props.files.audio_files.length > 1">
                <span class="file-header" @click="showAudioFiles = !showAudioFiles">
                    <svg viewBox="0 0 123 123">
                        <path
                            d="M111.85,108.77c-3.47,4.82-8.39,8.52-14.13,10.48c-0.26,0.12-0.55,0.18-0.84,0.18c-0.28,0-0.56-0.06-0.82-0.17v0.06 c0,1.96-1.6,3.56-3.57,3.56l-7.68,0c-1.96,0-3.57-1.6-3.57-3.56l0-55.13c0-1.96,1.6-3.57,3.57-3.57h7.68c1.96,0,3.57,1.6,3.57,3.57 v0.34c0.26-0.12,0.54-0.18,0.82-0.18c0.22,0,0.44,0.04,0.64,0.1l0,0.01c4.36,1.45,8.26,3.92,11.42,7.11V59.15 c0-14.89-4.99-27.63-13.81-36.6l-3.91,5.83c-7.95-8.75-19.4-14.27-32.08-14.27c-12.76,0-24.29,5.59-32.24,14.45l-4.73-5.78 C13.47,31.65,8.54,44.21,8.54,59.15V73.4c3.4-4.08,7.92-7.22,13.07-8.93l0-0.01c0.21-0.07,0.43-0.11,0.64-0.11 c0.28,0,0.57,0.06,0.82,0.17v-0.34c0-1.96,1.61-3.57,3.57-3.57l7.68,0c1.96,0,3.57,1.6,3.57,3.57v55.13c0,1.96-1.61,3.56-3.57,3.56 h-7.68c-1.96,0-3.57-1.6-3.57-3.56v-0.06c-0.25,0.11-0.53,0.17-0.82,0.17c-0.3,0-0.58-0.07-0.83-0.18 c-5.74-1.96-10.66-5.66-14.13-10.48c-1.82-2.52-3.24-5.34-4.17-8.37l-3.12,0V59.15c0-16.27,6.65-31.05,17.37-41.77 C28.09,6.66,42.88,0,59.14,0c16.27,0,31.06,6.66,41.77,17.37c10.72,10.72,17.37,25.5,17.37,41.77v41.25h-2.27 C115.1,103.39,113.68,106.23,111.85,108.77L111.85,108.77L111.85,108.77z" />
                    </svg>
                    {{ props.files.audio_files.length }} Audio Files
                </span>
            </div>
            <div v-else>
                <span class="file-header no-click">
                    <svg viewBox="0 0 123 123">
                        <path
                            d="M111.85,108.77c-3.47,4.82-8.39,8.52-14.13,10.48c-0.26,0.12-0.55,0.18-0.84,0.18c-0.28,0-0.56-0.06-0.82-0.17v0.06 c0,1.96-1.6,3.56-3.57,3.56l-7.68,0c-1.96,0-3.57-1.6-3.57-3.56l0-55.13c0-1.96,1.6-3.57,3.57-3.57h7.68c1.96,0,3.57,1.6,3.57,3.57 v0.34c0.26-0.12,0.54-0.18,0.82-0.18c0.22,0,0.44,0.04,0.64,0.1l0,0.01c4.36,1.45,8.26,3.92,11.42,7.11V59.15 c0-14.89-4.99-27.63-13.81-36.6l-3.91,5.83c-7.95-8.75-19.4-14.27-32.08-14.27c-12.76,0-24.29,5.59-32.24,14.45l-4.73-5.78 C13.47,31.65,8.54,44.21,8.54,59.15V73.4c3.4-4.08,7.92-7.22,13.07-8.93l0-0.01c0.21-0.07,0.43-0.11,0.64-0.11 c0.28,0,0.57,0.06,0.82,0.17v-0.34c0-1.96,1.61-3.57,3.57-3.57l7.68,0c1.96,0,3.57,1.6,3.57,3.57v55.13c0,1.96-1.61,3.56-3.57,3.56 h-7.68c-1.96,0-3.57-1.6-3.57-3.56v-0.06c-0.25,0.11-0.53,0.17-0.82,0.17c-0.3,0-0.58-0.07-0.83-0.18 c-5.74-1.96-10.66-5.66-14.13-10.48c-1.82-2.52-3.24-5.34-4.17-8.37l-3.12,0V59.15c0-16.27,6.65-31.05,17.37-41.77 C28.09,6.66,42.88,0,59.14,0c16.27,0,31.06,6.66,41.77,17.37c10.72,10.72,17.37,25.5,17.37,41.77v41.25h-2.27 C115.1,103.39,113.68,106.23,111.85,108.77L111.85,108.77L111.85,108.77z" />
                    </svg>
                    {{ getBaseName(props.files.audio_files[0]) }}
                </span>
            </div>
            <div v-if="showAudioFiles">
                <div v-if="props.files.audio_files && props.files.audio_files.length > 0" class="file-row"
                    v-for="file in props.files.audio_files" :key="file">
                    {{ getBaseName(file) }}
                </div>
            </div>
        </div>

        <!-- Other Files -->
        <div v-if="props.files.other_files && props.files.other_files.length > 0">
            <div v-if="props.files.other_files.length > 1">
                <span class="file-header" @click="showOtherFiles = !showOtherFiles">
                    <svg viewBox="0 0 123 123">
                        <path
                            d="M5.32,14.64h20.51V5.32v0h0.01c0-1.47,0.6-2.8,1.56-3.76c0.95-0.95,2.28-1.55,3.75-1.55V0h0h39.61h1.22l0.88,0.88 l31.29,31.41l0.87,2.09v69.2v0h-0.01c0,1.47-0.59,2.8-1.55,3.76h-0.01c-0.95,0.96-2.28,1.55-3.75,1.55v0.01h0H79.19v8.65v0h-0.01 c0,1.47-0.59,2.8-1.55,3.76h-0.01c-0.96,0.95-2.28,1.55-3.75,1.55v0.01h0H5.32h0v-0.01c-1.47,0-2.8-0.6-3.76-1.56 c-0.95-0.96-1.55-2.28-1.55-3.75H0v0V19.97v0h0.01c0-1.47,0.6-2.8,1.56-3.76c0.95-0.95,2.28-1.55,3.75-1.55L5.32,14.64L5.32,14.64 L5.32,14.64z M31.76,14.64h13.17h1.22l0.88,0.88l31.29,31.41l0.87,2.09v53.95h19.89V36.24H74.73h0v0c-1.78,0-3.39-0.74-4.56-1.94 c-1.17-1.19-1.9-2.84-1.9-4.65h0v0V5.94H31.76V14.64L31.76,14.64z M68.39,2.97h2.37l31.29,31.41v1.74H74.73 c-3.49,0-6.35-2.92-6.35-6.48V2.97L68.39,2.97z M73.26,50.88H48.91h0v0c-1.78,0-3.39-0.74-4.56-1.94c-1.17-1.19-1.9-2.84-1.9-4.65 h0v0V20.58H25.83H5.94v96.36h67.32v-8.04v-2.97V50.88L73.26,50.88z" />
                    </svg>
                    {{ props.files.other_files.length }} Other Files
                </span>
            </div>
            <div v-else>
                <span class="file-header no-click">
                    <svg viewBox="0 0 123 123">
                        <path
                            d="M5.32,14.64h20.51V5.32v0h0.01c0-1.47,0.6-2.8,1.56-3.76c0.95-0.95,2.28-1.55,3.75-1.55V0h0h39.61h1.22l0.88,0.88 l31.29,31.41l0.87,2.09v69.2v0h-0.01c0,1.47-0.59,2.8-1.55,3.76h-0.01c-0.95,0.96-2.28,1.55-3.75,1.55v0.01h0H79.19v8.65v0h-0.01 c0,1.47-0.59,2.8-1.55,3.76h-0.01c-0.96,0.95-2.28,1.55-3.75,1.55v0.01h0H5.32h0v-0.01c-1.47,0-2.8-0.6-3.76-1.56 c-0.95-0.96-1.55-2.28-1.55-3.75H0v0V19.97v0h0.01c0-1.47,0.6-2.8,1.56-3.76c0.95-0.95,2.28-1.55,3.75-1.55L5.32,14.64L5.32,14.64 L5.32,14.64z M31.76,14.64h13.17h1.22l0.88,0.88l31.29,31.41l0.87,2.09v53.95h19.89V36.24H74.73h0v0c-1.78,0-3.39-0.74-4.56-1.94 c-1.17-1.19-1.9-2.84-1.9-4.65h0v0V5.94H31.76V14.64L31.76,14.64z M68.39,2.97h2.37l31.29,31.41v1.74H74.73 c-3.49,0-6.35-2.92-6.35-6.48V2.97L68.39,2.97z M73.26,50.88H48.91h0v0c-1.78,0-3.39-0.74-4.56-1.94c-1.17-1.19-1.9-2.84-1.9-4.65 h0v0V20.58H25.83H5.94v96.36h67.32v-8.04v-2.97V50.88L73.26,50.88z" />
                    </svg>
                    {{ getBaseName(props.files.other_files[0]) }}
                </span>
            </div>
            <div v-if="showOtherFiles">
                <div v-if="props.files.audio_files && props.files.audio_files.length > 0" class="file-row"
                    v-for="file in props.files.audio_files" :key="file">
                    {{ getBaseName(file) }}
                </div>
            </div>
        </div>

    </div>
</template>

<style scoped>
.files-view {
    width: 100%;

    display: flex;
    flex-direction: column;

    --metadata-color: rgba(233, 150, 226, 0.2);
    --book-files-color: rgba(143, 218, 236, 0.2);
    --audio-files-color: rgba(236, 143, 143, 0.2);
}

.files-view>div {
    flex-grow: 1;
    border-bottom: 1px solid var(--color-border);
    align-content: center;
}

.files-view svg {
    max-width: 20px;
    max-height: 20px;
    fill: var(--color-text);
}

.file-header {
    display: flex;
    justify-content: space-between;
    padding: 0.5rem;
    align-items: center;
    cursor: pointer;
}

.file-header:hover {
    text-decoration: underline;
}

.file-header.metadata {
    background: var(--metadata-color);
    cursor: default;
    text-decoration: none;
}

.file-header.no-click {
    cursor: default;
    text-decoration: none;
}

.file-list {
    flex: 1;
    padding: 1rem;
}

.file-row {
    padding: 0.4rem 0;
    border-bottom: 1px solid var(--color-border);
    font-size: 0.9rem;
    margin-left: 1rem;
}

.file-row:last-child {
    border-bottom: none;
}
</style>