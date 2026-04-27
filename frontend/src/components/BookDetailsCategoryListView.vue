<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import type { Category } from '@/types/book';

interface Props {
  list: Category[] | string[];
  title: string;
  fieldName: string;
  listLimit: number;
}

const showAll = ref(false);
const shortList = computed(() => props.list?.slice(0, props.listLimit));
const listLength = computed(() => props.list?.length ?? 0);
const props = defineProps<Props>();

const emit = defineEmits<{
  itemClick: [query: Record<string, string>];
}>();

function handleItemClick(itemName: string) {
  emit('itemClick', { [props.fieldName]: itemName });
}

function toggleShowAll() {
  showAll.value = !showAll.value;
}

onMounted(() => {
  if (listLength.value < props.listLimit) {
    showAll.value = true;
  }
})

</script>

<template>
  <div class="category-list-view">
    <dt>
      {{ title }}
    </dt>
    <div v-if="listLength > 0">
        <div v-if="showAll" class="search-category-button">
          <dd v-for="(item, index) in list" :key="typeof item === 'string' ? item : item.name">
            <a @click="handleItemClick(typeof item === 'string' ? item : item.name)">
              {{ typeof item === 'string' ? item : item.name }}{{ index < list.length - 1 ? ', ' : '' }}
            </a>
          </dd>
          <a v-if="listLength > listLimit"
             class="expand-category-button"
             @click="toggleShowAll">
             (hide...)
          </a>
        </div>
        <div v-else class="search-category-button">
          <dd v-for="(item, index) in shortList" :key="typeof item === 'string' ? item : item.name">
            <a @click="handleItemClick(typeof item === 'string' ? item : item.name)">
              {{ typeof item === 'string' ? item : item.name }}{{ index < listLimit - 1 ? ', ' : '' }}
            </a>
          </dd>
          <a v-if="listLength > listLimit"
             class="expand-category-button"
             @click="toggleShowAll">
             (show {{ list.length - listLimit }} more...)
          </a>
        </div>
    </div>
  </div>
</template>

<style scoped>

.category-list-view {
    color: var(--color-text);
}

.category-list-view dt {
  font-size: 0.85rem;
  color: var(--color-text-subtle);
  font-weight: 600;    
}
.category-list-view dd {
  margin: 0;
  font-size: 0.9rem;
}

.search-category-button {
  color: var(--color-text);
  cursor: pointer;
  display: flex;
  flex-wrap: wrap;
}

.expand-category-button {
  color: var(--color-text);
  font-style: italic;
}
</style>
