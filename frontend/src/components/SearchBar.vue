<script setup lang="ts">
import { onMounted, ref } from 'vue';

const searchTerm = ref('');
const showAdvanced = ref(false);

const props = defineProps({
    placeholder: {
        type: String
    },
    showAdvanced: {
        type: Boolean
    },
    searchTerm: {
        type: String
    }
})

const emit = defineEmits(['searchTerm', 'showAdvanced', 'reset']);

const searchBooks = () => {
    emit('searchTerm', searchTerm.value);
};

const reset = () => {
    searchTerm.value = '';
    emit('reset');
}

const toggleShowAdvanced = () => {
    showAdvanced.value = !showAdvanced.value;
    emit('showAdvanced', showAdvanced.value);
}

onMounted(() => {
    searchTerm.value = props.searchTerm || '';
    showAdvanced.value = props.showAdvanced || false;
})

</script>

<template>
    <div class="search-bar">
        <input class="text-input" v-model="searchTerm" type="search" :placeholder="props.placeholder || 'Search books'"
            aria-label="Search books" @keyup.enter="searchBooks" />
        <button class="search-button" type="button" @click="searchBooks()">Search</button>
        <div class="secondary-controls">
            <button class="toggle-button" type="button" @click="toggleShowAdvanced()">
                {{ showAdvanced ? 'Hide' : 'Advanced' }}
            </button>
            <button class="toggle-button" type="button" @click="reset()">Reset</button>
        </div>
    </div>
</template>

<style scoped>

.search-bar {
    display: flex;
    gap: 0.75rem;
}

.search-button {
    max-width: 100px;
}

.secondary-controls {
    min-height: 42px;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    gap: 0.75rem;
}

@media (max-width: 768px) {
    
    .search-bar {
        flex-direction: column;
    }
    
    .search-button {
        max-width: unset;
    }
}

</style>