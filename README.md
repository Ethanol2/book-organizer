# <img src="frontend/src/assets/book-organizer-logo.png" width="50"> book-organizer

An app written in Go and designed to run in Docker that helps organize a book library.

## Description and Motivation

This app will scan a downloads folder for new book directories, let users associate each download with a book entry, and then move the files into a structured library folder. It’s heavily inspired by the *arr* apps (RIP Readarr).

I wanted to make this app because I have a big audiobook library and I wanted a tool to manage it. I know alternatives exist, plus a whole host of projects inspired by Readarr like this one, so this is also a learning opportunity for me.

---

## Quick Start

At the moment I haven't moved to get the project onto docker.

Clone or download the repo to your local machine. I'm not sure if it works using github codespaces, but let me know if you try it and it does.

- Backend
  - Setup your `.env` file. Use the `.env (example)` as a base.
      - `DB_PATH`: The app will create the file automatically at the set location
      - `FRONTEND_PATH`: Leave default. Not currently used
      - `DOWNLOADS_PATH`: The app will search this directory for new downloads.
      - `LIBRARY_PATH`: Directory where books are moved once they're associated with a download.
      - `PORT`: Leave defaulf if unsure. The frontend is setup to use this port during development. 
      - `GOOGLE_BOOKS_API_KEY`: Fill if you want google books metadata fetching. This involves figuring out google's api keys with your own google account.

    The directories must exist

- Frontend
  - Install npm
  
  - Install Node.js
    https://nodejs.org/en/download

  - Install frontend dependencies
    ```bash
    cd frontend
    npm install
    ```

## Usage ⚙️

- **Start the app locally:**
  - Backend
    ```bash
    go run main.go
    ```
    - Flags available at startup:
      - `-r` — reset (remove) the DB file before starting
      - `-t` — insert test data (implies reset)
  - Frontend
    ```bash
    cd frontend
    npm install
    npm run dev
    ```
    The frontend will be available at the local development server (typically `http://localhost:5173`)

### Frontend Development

Build and production:
```bash
npm run build      # Build for production
npm run preview    # Preview production build locally
npm run type-check # Type check TypeScript code
npm run lint       # Lint and format code
```
---

## Contributing

If you'd like to contribute, please fork the repository and open a pull request to the `main` branch.
 
## Technologies

- **Language**: Go (backend)
- **Frontend**: Vue 3 with TypeScript, Vite bundler, Pinia state management
- **Database**: SQLite
- **Runtime**: Docker *(planned)*

### Frontend Stack
- Vue 3 (Composition API)
- TypeScript for type safety
- Vite for fast development and optimized builds
- Vue Router for navigation
- Pinia for state management
- ESLint + Prettier for code quality

### Backend Stack
- Go standard library with custom middleware
- SQLite3 driver
- RESTful API design

## Features

### New File Scanning

The app periodically scans a configurable downloads folder for new book directories. It detects:

- The folder name
- File types (audio files, ebook files, cover art, etc.)

From the UI, users can view and manage this pending list, associating downloads with library entries.

### Book Library

Books are stored in a SQLite database. Once a pending download is associated with a book, the app moves its files into the library folder using a fixed structure:

`Author/Series/Book Title/`

The system can fetch metadata from Google Books and OpenLibrary.

### Frontend

A responsive Vue 3 single-page application with the following pages:

- **Library** — Browse all books in your collection with filtering and advanced search
- **Downloads** — View pending imports and associate them with books
- **Add Book** — Search metadata sources and create new book entries
- **Book Details** — View and edit book information, metadata, and associated files
- **About** — Application information

### Backend

Go-based RESTful API handling all data management, metadata search, file organization, and database operations.

## Planned Features

### Audible

In my experience, Audible is a vastly better source of metadata than OpenLibrary and Google Books. The complication is that the endpoints aren't officially accessible, so no official documentation (as far as I know).

### Login

Add a login system to protect libraries that are publically accessible

### Library Scanning

Add a scanning system for adding books already in a library



### qBittorrent Integration (Nice-to-have)

After the core library flow is working, the app will integrate with qBittorrent to track labeled downloads. For completed torrents with a specific label (e.g. `books`), the app will automatically add them to the pending list. The frontend will also surface whether qBittorrent is reachable.


## API Reference ✅

Below are the current HTTP endpoints and the JSON structures they expect and return.

---

### Downloads 📥

- **GET /api/downloads**
  - **Description:** List pending downloads (items in the downloads folder that have not yet been associated with a book)
  - **Response:** 200 OK — array of `Download` objects

- **GET /api/downloads/{id}**
  - **Description:** Get a single pending download by UUID
  - **Response:** 200 OK — single `Download` object

- **GET /api/downloads/{id}/cover**
  - **Description:** Serve the cover image file associated with a download (if present)
  - **Response:** 200 OK — binary image (jpeg/png/webp/gif)

- **POST /api/downloads/{id}/associate**
  - **Description:** Associate a pending download with an existing book and move files into the library
  - **Request JSON:**
    ```json
    {
      "book_id": "<uuid>"
    }
    ```
  - **Response:** 200 OK — the updated `Book` object after association

---

### Categories 🏷️

- **POST /api/categories/{categoryType}**
  - **Description:** Add a category value to one of `series`, `genres`, `narrators`, or `authors`
  - **Request JSON:**
    ```json
    {
      "value": "<string>"
    }
    ```
  - **Response:** 200 OK — the created `Category` object

- **GET /api/categories/{categoryType}**
  - **Description:** Get all values for a given category type
  - **Response:** 200 OK — object with `values: Category[]`
    ```json
    {
      "values": [ /* array of Category */ ]
    }
    ```

---

### Books 📚

- **GET /api/books**
  - **Description:** List all books
  - **Response:** 200 OK — array of `Book` objects

- **GET /api/books/{id}**
  - **Description:** Get a single book by UUID
  - **Response:** 200 OK — single `Book` object

- **GET /api/books/{id}/cover**
  - **Description:** Serve the cover image file associated with a book (if present)
  - **Response:** 200 OK — binary image (jpeg/png/webp/gif)

- **POST /api/books**
  - **Description:** Create a new book
  - **Request JSON (BookParams)** — all fields are optional, include only what you want to set:
    ```json
    {
      "title": "<string>",
      "description": "<string>",
      "year": 2024,
      "isbn": "<string>",
      "asin": "<string>",
      "tags": ["tag1", "tag2"],
      "publisher": "<string>",
      "series": [{ "name": "Series Name" }],
      "authors": [{ "name": "Author Name" }],
      "genres": [{ "name": "Genre" }],
      "narrators": [{ "name": "Narrator Name" }]
    }
    ```
  - **Response:** 200 OK — created `Book` object

- **PATCH /api/books/{id}**
  - **Description:** Update fields on an existing book (same schema as `BookParams`; send only fields to change)
  - **Request JSON:** same as `POST /api/books`
  - **Response:** 200 OK — updated `Book` object

- **PATCH /api/books/{id}/cover**
  - **Description:** Upload a new cover image for a book
  - **Request:** Raw binary image in the request body with `Content-Type: image/jpeg|png|webp|gif`
  - **Response:** 200 OK — updated `Book` object

- **DELETE /api/books/{id}**
  - **Description:** Delete a book from the library
  - **Response:** 200 OK — confirmation message

---

### Metadata 🔎

- **GET /api/metadata/**
  - **Description:** Search for metadata from OpenLibrary or Google Books
  - **Query Params:** (all fields are optional)
    - `source` — metadata source to search (`openlibrary` or `googlebooks`; defaults to `openlibrary`)
    - `title` — book title
    - `author` — author name
    - `year` — publication year
    - `publisher` — publisher name
    - `isbn` — ISBN
    - `genre` — genre(s) (repeatable: `?genre=scifi&genre=fantasy`)
    - `language` — language code(s) (repeatable: `?language=eng&language=fr`)
  - **Response:** 200 OK — `SearchResults` object

- **GET /api/metadata/{id}**
  - **Description:** Get detailed metadata for a specific result from a metadata source
  - **URL Params:**
    - `id` — metadata provider ID (OpenLibrary work ID or Google Books volume ID)
  - **Query Params:**
    - `source` — metadata source (`openlibrary` or `googlebooks`)
  - **Response:** 200 OK — full `Book` object with metadata populated

---

### Media static file access 📂

These are served directly from the configured folders:

- **GET /media/downloads/{path...}** — files in the downloads directory
- **GET /media/library/{path...}** — files in the library directory
- **GET /media/metadata/{path...}** — stored metadata/cover assets

---

## Example Usage (curl) 🧪

### Downloads

- List pending downloads:
  ```bash
  curl -s http://localhost:8080/api/downloads | jq .
  ```

- Get a download by ID:
  ```bash
  curl -s http://localhost:8080/api/downloads/<DOWNLOAD_ID> | jq .
  ```

- Associate a download with an existing book:
  ```bash
  curl -X POST http://localhost:8080/api/downloads/<DOWNLOAD_ID>/associate \
    -H "Content-Type: application/json" \
    -d '{"book_id":"<BOOK_ID>"}' | jq .
  ```

### Books

- Create a book:
  ```bash
  curl -X POST http://localhost:8080/api/books \
    -H "Content-Type: application/json" \
    -d '{"title":"My Book","authors":[{"name":"Me"}]}' | jq .
  ```

- Update a book:
  ```bash
  curl -X PATCH http://localhost:8080/api/books/<BOOK_ID> \
    -H "Content-Type: application/json" \
    -d '{"description":"Updated summary"}' | jq .
  ```

- Upload a cover image:
  ```bash
  curl -X PATCH http://localhost:8080/api/books/<BOOK_ID>/cover \
    -H "Content-Type: image/jpeg" \
    --data-binary @cover.jpg | jq .
  ```

- Delete a book:
  ```bash
  curl -X DELETE http://localhost:8080/api/books/<BOOK_ID> | jq .
  ```

### Metadata

- Search OpenLibrary (JSON body):
  ```bash
  curl -X GET http://localhost:8080/api/metadata/openlibrary \
    -H "Content-Type: application/json" \
    -d '{"title":"The Martian"}' | jq .
  ```

- Search Google Books (requires GOOGLE_BOOKS_API_KEY):
  ```bash
  curl -X GET http://localhost:8080/api/metadata/googlebooks \
    -H "Content-Type: application/json" \
    -d '{"title":"The Martian"}' | jq .
  ```

---

## JSON Schemas (struct reference) 🔧

- `Book` (response)
  ```json
  {
    "id": "<uuid>",
    "title": "<string>",
    "description": "<string>",
    "year": <int|null>,
    "isbn": "<string>",
    "asin": "<string>",
    "tags": ["string"],
    "publisher": "<string>",
    "created_at": "<timestamp>",
    "updated_at": "<timestamp>",

    "series": [ { "id": <int|null>, "name": "<string>", "index": "<string>|null" } ],
    "authors": [ { "id": <int|null>, "name": "<string>" } ],
    "genres": [ { "id": <int|null>, "name": "<string>" } ],
    "narrators": [ { "id": <int|null>, "name": "<string>" } ],

    "files": {
      "audio_files": { "files": ["file1.m4b"] },
      "text_files": { "files": ["file.epub"] },
      "cover": "<relative_or_url_to_cover>"
    }
  }
  ```

- `Download` (response)
  ```json
  {
    "id": "<uuid>",
    "created_at": "<timestamp>",
    "directory_name": "<string>",
    "files": { /* same shape as Book.files */ }
  }
  ```

- `Category` (response)
  ```json
  {
    "id": <int|null>,
    "name": "<string>"
  }
  ```
