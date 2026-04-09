# book-organizer

An app written in Go and designed to run in Docker that helps organize a book library.

**Currently on hold**

This project benefits me, so I definitely want to continue work and finish it. Due to life circumstances and minor burnout, I'm shelving the project for now. Might come back to this in a week, might be 6 months.

The backend portion is pretty complete. I'm sure as the frontend continues development I'll have to make changes.

The frontend development is where I'm technically most weak. What is currently developed is a combination of my learning and assists from ChatGPT. I'm hopeful that I can power through the learning curve and create something useable.

## Description and Motivation

This app will scan a downloads folder for new book directories, let users associate each download with a book entry, and then move the files into a structured library folder. It’s heavily inspired by the *arr* apps (RIP Readarr).

I wanted to make this app because I have a big audiobook library and I wanted a tool to manage it. I know alternatives exist, plus a whole host of projects inspired by Readarr like this one, so this is also a learning opportunity for me.

---

## Quick start

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
    npm run dev
    ```
---

## Contributing

If you'd like to contribute, please fork the repository and open a pull request to the `main` branch.
 
## Technologies

- **Language**: Go
- **Frontend**: Vue *(in development)*
- **Database**: SQLite
- **Runtime**: Docker *(planned)*

## Features

### New File Scanning

The app periodically scans a configurable downloads folder for new book directories.

- The folder name
- Detected file types (audio files, ebook files, cover art, etc.)

From the UI, are able to view and will be able to manage this pending list.

### Book Library

Books are stored in a `books` table in SQLite. Once a pending download is associated with a book, the app will move its files into the library folder using a fixed structure, for example:

`Author/Series/Book Title/`

Currently the backend can make queries to Google Books and OpenLibrary.

### Frontend and Backend

- **Backend**: Go, exposing a RESTful HTTP API.
- **Frontend**: A Vue single-page application that interacts with the API to:
  - List pending downloads
  - Search and select metadata
  - Create and edit book entries
  - Trigger imports/moves

## Planned Features

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

---

### Metadata 🔎

- **GET /api/metadata/openlibrary**
  - **Description:** Search OpenLibrary for metadata matching the provided query parameters
  - **Request JSON:** `SearchParams` (all fields are optional)
  - **Response:** 200 OK — `SearchResults`

- **GET /api/metadata/googlebooks**
  - **Description:** Search Google Books for metadata (requires `GOOGLE_BOOKS_API_KEY` in env)
  - **Request JSON:** `SearchParams` (all fields are optional)
  - **Response:** 200 OK — `SearchResults`

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

### Categories

- Add a category value:
  ```bash
  curl -X POST http://localhost:8080/api/categories/series \
    -H "Content-Type: application/json" \
    -d '{"value":"My Series"}' | jq .
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
