<script setup lang="ts">
// Modal component for adding/editing books with form fields and submission

interface ModalProps {
  show: boolean
  title: string
  subtitle: string
  description: string
  year: string
  isbn: string
  publisher: string
  authors: string
  genres: string
  cover: string
}

const props = withDefaults(defineProps<ModalProps>(), {
  show: false,
  title: '',
  subtitle: '',
  description: '',
  year: '',
  isbn: '',
  publisher: '',
  authors: '',
  genres: '',
  cover: '',
})

// Emits for closing modal and submitting book data
const emit = defineEmits<{
  close: []
  'update:title': [value: string]
  'update:subtitle': [value: string]
  'update:description': [value: string]
  'update:year': [value: string]
  'update:isbn': [value: string]
  'update:publisher': [value: string]
  'update:authors': [value: string]
  'update:genres': [value: string]
  'update:cover': [value: string]
  'add-book': [bookData: {
    title: string | null
    subtitle: string | null
    description: string | null
    year: number | null
    isbn: string | null
    publisher: string | null
    authors: Array<{ name: string }> | null
    genres: Array<{ name: string }> | null
    cover: string | null
  }]
}>()

// Handle form submission
function handleSubmit() {
  const bookData = {
    title: props.title || null,
    subtitle: props.subtitle || null,
    description: props.description || null,
    year: props.year ? parseInt(props.year) : null,
    isbn: props.isbn || null,
    publisher: props.publisher || null,
    authors: props.authors ? props.authors.split(',').map(name => ({ name: name.trim() })) : null,
    genres: props.genres ? props.genres.split(',').map(name => ({ name: name.trim() })) : null,
    cover: props.cover || null,
  }
  emit('add-book', bookData)
}

// Close modal when clicking overlay
function handleOverlayClick() {
  emit('close')
}

// Prevent closing when clicking inside modal
function handleModalClick(e: Event) {
  e.stopPropagation()
}
</script>

<template>
  <!-- Modal overlay - closes when clicked -->
  <div v-if="show" class="modal-overlay" @click="handleOverlayClick">
    <!-- Modal dialog - prevents closing from interior clicks -->
    <div class="modal" @click="handleModalClick">
      <h3>Add Book</h3>
      <form @submit.prevent="handleSubmit">
        <!-- Book metadata form fields -->
        <label>Title: <input :value="title" @input="emit('update:title', ($event.target as HTMLInputElement).value)" type="text" required /></label>
        <label>Subtitle: <input :value="subtitle" @input="emit('update:subtitle', ($event.target as HTMLInputElement).value)" type="text" /></label>
        <label>Description: <textarea :value="description" @input="emit('update:description', ($event.target as HTMLTextAreaElement).value)"></textarea></label>
        <label>Year: <input :value="year" @input="emit('update:year', ($event.target as HTMLInputElement).value)" type="number" /></label>
        <label>ISBN: <input :value="isbn" @input="emit('update:isbn', ($event.target as HTMLInputElement).value)" type="text" /></label>
        <label>Publisher: <input :value="publisher" @input="emit('update:publisher', ($event.target as HTMLInputElement).value)" type="text" /></label>
        <label>Authors (comma-separated): <input :value="authors" @input="emit('update:authors', ($event.target as HTMLInputElement).value)" type="text" /></label>
        <label>Genres (comma-separated): <input :value="genres" @input="emit('update:genres', ($event.target as HTMLInputElement).value)" type="text" /></label>
        <label>Cover URL: <input :value="cover" @input="emit('update:cover', ($event.target as HTMLInputElement).value)" type="text" /></label>

        <!-- Action buttons -->
        <div class="modal-buttons">
          <button type="button" @click="handleOverlayClick">Cancel</button>
          <button type="submit">Add Book</button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
/* Modal overlay background */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

/* Modal dialog container */
.modal {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  max-width: 600px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}

.modal h3 {
  margin-top: 0;
  margin-bottom: 1rem;
}

/* Form labels and inputs */
.modal label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.modal input,
.modal textarea {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  margin-bottom: 1rem;
  box-sizing: border-box;
  font-family: inherit;
  font-size: 0.95rem;
}

.modal textarea {
  height: 100px;
  resize: vertical;
}

/* Modal action buttons */
.modal-buttons {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}

.modal-buttons button {
  padding: 0.5rem 1rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  background: #f9f9f9;
  cursor: pointer;
  font-size: 0.95rem;
  transition: background-color 0.2s;
}

.modal-buttons button:hover {
  background: #e9e9e9;
}

/* Submit button styling */
.modal-buttons button[type="submit"] {
  background: #4CAF50;
  color: white;
  border-color: #45a049;
}

.modal-buttons button[type="submit"]:hover {
  background: #45a049;
}
</style>
