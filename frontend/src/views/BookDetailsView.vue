<script setup lang="ts">
import { useRoute } from 'vue-router';


const route = useRoute()
const bookTitle = route.params.title

</script>

<template>
    Book details
</template>