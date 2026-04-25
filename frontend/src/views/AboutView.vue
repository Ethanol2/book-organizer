<template>
  <div class="about">
    <div class="about-header">
      <h1>About Book Organizer</h1>
      <p class="subtitle">A personal book library management tool</p>
    </div>

    <section class="about-section">
      <h2>What is Book Organizer?</h2>
      <p>
        Book Organizer is an application designed to help manage a personal book library. It scans a downloads
        folder for new book directories, lets you associate each download with a book entry, and automatically
        moves files into a structured library folder. It's heavily inspired by the <em>*arr</em> apps like Sonarr
        and Radarr.
      </p>
      <p>
        The app combines a modern Vue 3 frontend with a robust Go backend to provide a seamless experience for
        organizing audiobooks, ebooks, and other written media.
      </p>
    </section>

    <section class="about-section">
      <h2>Key Features</h2>
      <ul class="features-list">
        <li><strong>Automatic File Scanning</strong> — Periodically detects new downloads and pending imports</li>
        <li><strong>Smart Organization</strong> — Moves files into structured folders: Author/Series/Title</li>
        <li><strong>Metadata Search</strong> — Fetches book information from OpenLibrary and Google Books</li>
        <li><strong>Advanced Library Management</strong> — Browse, filter, and search your entire book collection</li>
        <li><strong>Cover Art Support</strong> — View and upload cover images for your books</li>
      </ul>
    </section>

    <section class="about-section contributing-section">
      <h2>Contributing</h2>
      <p>
        Interested in contributing? The project is open source! Feel free to fork the repository and open a pull
        request to the <code>main</code> branch with your improvements.
      </p>
      <div class="contributing-links">
        <a href="https://github.com/Ethanol2/book-organizer" target="_blank" rel="noopener noreferrer" class="cta-button">
          Visit GitHub Repository
        </a>
      </div>
    </section>

    <section class="about-section">
      <h2>Technology Stack</h2>
      <div class="tech-stack">
        <div class="tech-column">
          <h3>Frontend</h3>
          <ul>
            <li>Vue 3</li>
            <li>TypeScript</li>
            <li>Vite</li>
            <li>Vue Router</li>
            <li>Pinia</li>
          </ul>
        </div>
        <div class="tech-column">
          <h3>Backend</h3>
          <ul>
            <li>Go</li>
            <li>SQLite</li>
            <li>RESTful API</li>
          </ul>
        </div>
      </div>
    </section>

    <section class="about-section">
      <h2>About the Creator</h2>
      <p>
        Book Organizer was created by me, <strong>Ethan Colucci</strong>, as both a practical tool for managing a
        personal audiobook library and a learning opportunity primarily focused on backend development. All the 
        backend go code has been written by me, while the frontend was built leaning heavily on AI agents.
      </p>
      <p>
        Connect with me:
      </p>
      <div class="social-links">
        <a href="https://ethanasc.ca" target="_blank" rel="noopener noreferrer" class="social-link">
          Personal Website
        </a>
        <a href="https://github.com/Ethanol2" target="_blank" rel="noopener noreferrer" class="social-link">
          GitHub Profile
        </a>
        <a href="https://www.linkedin.com/in/ethan-colucci/" target="_blank" rel="noopener noreferrer" class="social-link">
          LinkedIn
        </a>
      </div>
    </section>

    <section class="about-section footer-section">
      <p class="copyright">© 2026 Book Organizer. Built by Ethan Colucci.</p>
    </section>
  </div>
</template>

<script setup lang="ts">
// About page component - displays project information and creator details
</script>

<style scoped>
.about {
  width: 100%;
  padding: 0;
  padding-bottom: 10rem;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.about-header {
  color: var(--vt-c-text-light-1, #f4efe6);
  padding: 3rem 2rem;
  text-align: center;
}

.about-header h1 {
  margin: 0 0 0.5rem 0;
  font-size: 2.5rem;
}

.subtitle {
  margin: 0;
  font-size: 1.1rem;
  opacity: 0.9;
}

.about-section {
  padding: 2rem;
  border-bottom: 1px solid var(--color-border, #eee);
  flex: 1;
}

.about-section:last-child {
  border-bottom: none;
}

.about-section h2 {
  color: var(--color-text, #333);
  margin-top: 0;
  margin-bottom: 1rem;
  font-size: 1.8rem;
}

.about-section p {
  color: var(--color-text);
  line-height: 1.6;
  margin: 0.5rem 0;
}

.features-list {
  list-style: none;
  padding: 0;
  margin: 1rem 0;
}

.features-list li {
  padding: 0.75rem 0;
  padding-left: 1.5rem;
  color: var(--color-text);
  line-height: 1.6;
  position: relative;
}

.features-list li:before {
  content: '✓';
  position: absolute;
  left: 0;
  color: var(--color-primary-blue);
  font-weight: bold;
}

.contributing-section {
  background: linear-gradient(135deg, rgba(126, 143, 122, 0.08) 0%, rgba(168, 178, 154, 0.08) 100%);
  border: 2px solid var(--color-primary-blue);
  border-radius: 8px;
}

.contributing-links {
  margin-top: 1.5rem;
}

.cta-button {
  display: inline-block;
  padding: 0.75rem 2rem;
  background: var(--color-primary-blue);
  color: white;
  text-decoration: none;
  border-radius: 6px;
  transition: all 0.3s ease;
  font-weight: 600;
}

.cta-button:hover {
  background: var(--color-primary-blue-dark);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(126, 143, 122, 0.3);
}

.tech-stack {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  margin-top: 1rem;
}

.tech-column h3 {
  color: var(--color-primary-blue);
  margin-top: 0;
  margin-bottom: 1rem;
}

.tech-column ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.tech-column li {
  padding: 0.5rem 0;
  color: var(--color-text);
  padding-left: 1.5rem;
  position: relative;
}

.tech-column li:before {
  content: '•';
  position: absolute;
  left: 0;
  color: var(--color-primary-blue);
}

.social-links {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin-top: 1.5rem;
}

.social-link {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  background: var(--color-primary-blue);
  color: white;
  text-decoration: none;
  border-radius: 6px;
  transition: all 0.3s ease;
  font-weight: 500;
}

.social-link:hover {
  background: var(--color-primary-blue-dark);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(126, 143, 122, 0.4);
}

code {
  background: var(--color-background-soft);
  padding: 0.2rem 0.4rem;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  color: var(--color-text);
}

.footer-section {
  text-align: center;
  background: var(--color-background-soft);
  margin-top: auto;
}

.copyright {
  color: var(--color-gray-600);
  font-size: 0.9rem;
  margin: 0;
}

@media (max-width: 768px) {
  .about-header h1 {
    font-size: 2rem;
  }

  .about-section {
    padding: 1.5rem;
  }

  .tech-stack {
    grid-template-columns: 1fr;
  }

  .social-links {
    flex-direction: column;
  }

  .social-link {
    display: block;
    text-align: center;
  }

  .cta-button {
    display: block;
    text-align: center;
  }
}
</style>
