<script setup lang="ts">
import type { Book } from '@/types/book';
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();
const book = ref<Book | null>(null);
const loading = ref(true);
const error = ref<string | null>(null);

const coverSrc = computed(() => {
  if (!book.value) {
    return '';
  }
  return book.value.files?.cover ?? `/media/metadata/${book.value.id}.jpg`;
});

const formattedAuthors = computed(() => book.value?.authors.map(a => a.name).join(', ') ?? 'Unknown');
const formattedGenres = computed(() => book.value?.genres.map(g => g.name).join(', ') ?? 'Unknown');
const formattedSeries = computed(() => book.value?.series.map(s => `${s.name} #${s.index}`).join(', ') ?? 'None');
const formattedNarrators = computed(() => book.value?.narrators.map(n => n.name).join(', ') ?? 'None');
const formattedTags = computed(() => book.value?.tags.length ? book.value.tags.join(', ') : 'None');

const audioFiles = computed(() => book.value?.files?.audio_files ?? []);
const textFiles = computed(() => book.value?.files?.text_files ?? []);

function formatDate(value: string | undefined | null) {
  if (!value) {
    return 'Unknown';
  }
  return new Date(value).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

onMounted(async () => {
  loading.value = true;
  error.value = null;

  try {
    const resp = await fetch(`/api/books/${route.params.id}`);
    if (!resp.ok) {
      throw new Error(`HTTP error with status: ${resp.status}`);
    }

    book.value = await resp.json();
  } catch (err) {
    console.error('Error fetching book details:', err);
    error.value = 'Unable to load book details.';
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <div class="book-details">
    <div v-if="loading" class="status-message">Loading book details…</div>
    <div v-else-if="error" class="status-message status-error">{{ error }}</div>
    <div v-else-if="book" class="details-shell">
      <section class="details-hero">
        <div class="hero-cover">
          <img :src="coverSrc" :alt="book.title" />
        </div>
        <div class="hero-info">
          <div class="hero-label">Book Details</div>
          <h1>{{ book.title }}</h1>
          <p v-if="book.subtitle" class="hero-subtitle">{{ book.subtitle }}</p>

          <div class="hero-stats">
            <div>
              <span class="stat-label">Year</span>
              <strong>{{ book.year || 'Unknown' }}</strong>
            </div>
            <div>
              <span class="stat-label">Publisher</span>
              <strong>{{ book.publisher || 'Unknown' }}</strong>
            </div>
            <div>
              <span class="stat-label">ISBN</span>
              <strong>{{ book.isbn || 'N/A' }}</strong>
            </div>
            <div>
              <span class="stat-label">ASIN</span>
              <strong>{{ book.asin || 'N/A' }}</strong>
            </div>
          </div>

          <div class="hero-tags" v-if="book.tags?.length">
            <span class="chip" v-for="tag in book.tags" :key="tag">{{ tag }}</span>
          </div>
        </div>
      </section>

      <div class="details-grid">
        <aside class="details-main">
          <section class="section-card summary-card">
            <div class="section-heading">
              <h2>Summary</h2>
            </div>
            <p v-if="book.description" class="description">{{ book.description }}</p>
            <p v-else class="description muted">No description available for this book.</p>
          </section>

          <section class="section-card relationships-card">
            <div class="section-heading">
              <h2>Relationships</h2>
            </div>
            <dl class="info-grid">
              <div>
                <dt>Authors</dt>
                <dd>{{ formattedAuthors }}</dd>
              </div>
              <div>
                <dt>Genres</dt>
                <dd>{{ formattedGenres }}</dd>
              </div>
              <div>
                <dt>Series</dt>
                <dd>{{ formattedSeries }}</dd>
              </div>
              <div>
                <dt>Narrators</dt>
                <dd>{{ formattedNarrators }}</dd>
              </div>
              <div>
                <dt>Tags</dt>
                <dd>{{ formattedTags }}</dd>
              </div>
            </dl>
          </section>
        </aside>

        <aside class="details-side">
          <section class="section-card metadata-card">
            <div class="section-heading">
              <h2>Metadata</h2>
            </div>
            <dl class="info-grid">
              <div>
                <dt>Created</dt>
                <dd>{{ formatDate(book.created_at) }}</dd>
              </div>
              <div>
                <dt>Updated</dt>
                <dd>{{ formatDate(book.updated_at) }}</dd>
              </div>
              <div>
                <dt>Cover path</dt>
                <dd>{{ coverSrc }}</dd>
              </div>
              <div>
                <dt>Root folder</dt>
                <dd>{{ book.files?.root || 'Unknown' }}</dd>
              </div>
            </dl>
          </section>

          <section class="section-card files-card">
            <div class="section-heading">
              <h2>Files</h2>
            </div>
            <div class="file-overview">
              <div class="file-metric">
                <span class="metric-label">Audio files</span>
                <strong>{{ audioFiles.length }}</strong>
              </div>
              <div class="file-metric">
                <span class="metric-label">Text files</span>
                <strong>{{ textFiles.length }}</strong>
              </div>
            </div>

            <div class="file-list-group" v-if="audioFiles.length">
              <h3>Audio file names</h3>
              <ul class="file-list">
                <li v-for="path in audioFiles" :key="path">{{ path }}</li>
              </ul>
            </div>

            <div class="file-list-group" v-if="textFiles.length">
              <h3>Text file names</h3>
              <ul class="file-list">
                <li v-for="path in textFiles" :key="path">{{ path }}</li>
              </ul>
            </div>
          </section>
        </aside>
      </div>
    </div>
    <div v-else class="status-message">Book not found.</div>
  </div>
</template>

<style scoped>
.book-details {
  width: 100%;
  padding: 1.5rem;
  overflow-y: auto;
}

.status-message {
  padding: 2rem 1.5rem;
  border-radius: 14px;
  background: var(--color-background-soft);
  color: var(--color-text);
  font-weight: 600;
}

.status-error {
  border: 1px solid rgba(220, 38, 38, 0.2);
  color: rgb(220, 38, 38);
}

.details-shell {
  display: grid;
  gap: 1.5rem;
}

.details-hero {
  display: grid;
  grid-template-columns: minmax(240px, 280px) 1fr;
  gap: 1.5rem;
  padding: 1.5rem;
  border-radius: 20px;
  background: radial-gradient(circle at top left, rgba(31, 96, 103, 0.12), transparent 40%),
    var(--color-background-soft);
  border: 1px solid var(--color-border);
}

.hero-cover {
  width: 100%;
  aspect-ratio: 5 / 7;
  overflow: hidden;
  border-radius: 18px;
  background: var(--color-background);
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.08);
}

.hero-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.hero-info {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 1.25rem;
}

.hero-label {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.9rem;
  border-radius: 999px;
  background: rgba(120, 138, 163, 0.12);
  color: var(--color-text);
  font-size: 0.85rem;
  font-weight: 700;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.hero-info h1 {
  margin: 0;
  font-size: clamp(2rem, 2.5vw, 3rem);
  line-height: 1.05;
}

.hero-subtitle {
  font-size: 1rem;
  color: rgba(60, 60, 60, 0.8);
  max-width: 72ch;
}

.hero-stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.hero-stats div {
  padding: 1rem;
  border-radius: 14px;
  background: var(--color-background);
  border: 1px solid var(--color-border);
}

.stat-label {
  display: block;
  color: rgba(60, 60, 60, 0.7);
  font-size: 0.85rem;
  margin-bottom: 0.35rem;
}

.hero-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.chip {
  display: inline-flex;
  padding: 0.45rem 0.8rem;
  border-radius: 999px;
  background: rgba(54, 140, 211, 0.1);
  color: var(--color-text);
  font-size: 0.85rem;
}

.details-grid {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(280px, 1fr);
  gap: 1.5rem;
}

.details-main,
.details-side {
  display: grid;
  gap: 1.5rem;
}

.section-card {
  padding: 1.35rem 1.5rem;
  border-radius: 20px;
  background: var(--color-background-soft);
  border: 1px solid var(--color-border);
}

.section-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.section-heading h2 {
  margin: 0;
  font-size: 1.1rem;
}

.description {
  margin: 0;
  color: rgba(60, 60, 60, 0.85);
  line-height: 1.75;
}

.muted {
  color: rgba(60, 60, 60, 0.55);
}

.info-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: minmax(110px, 160px) minmax(0, 1fr);
}

.info-grid dt {
  font-size: 0.85rem;
  color: rgba(60, 60, 60, 0.7);
}

.info-grid dd {
  margin: 0;
  font-size: 0.95rem;
  color: rgba(20, 20, 20, 0.95);
}

.file-overview {
  display: grid;
  gap: 0.85rem;
}

.file-metric {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.95rem 1rem;
  border-radius: 14px;
  background: rgba(54, 140, 211, 0.08);
}

.file-metric .metric-label {
  color: rgba(60, 60, 60, 0.75);
  font-size: 0.9rem;
}

.file-list-group h3 {
  margin: 1rem 0 0.6rem;
  font-size: 0.95rem;
  color: rgba(60, 60, 60, 0.8);
}

.file-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 0.35rem;
}

.file-list li {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
  font-size: 0.88rem;
  padding: 0.65rem 0.9rem;
  border-radius: 12px;
  background: var(--color-background);
  border: 1px solid var(--color-border);
  color: rgba(30, 30, 30, 0.9);
}

@media (max-width: 960px) {
  .details-hero,
  .details-grid {
    grid-template-columns: 1fr;
  }

  .hero-stats {
    grid-template-columns: 1fr;
  }
}
</style>
