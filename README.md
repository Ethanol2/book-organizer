# book-organizer

An app written in Go and designed to run in Docker that helps organize a book library.

## Description

This app will scan a downloads folder for new book directories, let users associate each download with a book entry, and then move the files into a structured library folder. It’s heavily inspired by the *arr* apps (RIP Readarr).
 
## Planned Features

### New File Scanning

The app will periodically scan a configurable downloads folder for new book directories. When a new directory is found, it will add a record to a `pending` table in SQLite, including:

- The folder name
- Detected file types (audio files, ebook files, cover art, etc.)

From the UI, users will be able to view and manage this pending list.

### Book Library

Books will be stored in a `books` table in SQLite. Once a pending directory is associated with a book, the app will move its files into the library folder using a fixed structure, for example:

`Author/Series/Book Title/`

The app will integrate with at least one metadata source to help populate book details (title, author, series, ISBN, etc.). Likely candidates are Audible, Open Library, or Google Books.

### Frontend and Backend

- **Backend**: Go, exposing a RESTful HTTP API.
- **Frontend**: A Vue single-page application that interacts with the API to:
  - List pending downloads
  - Search and select metadata
  - Create and edit book entries
  - Trigger imports/moves

### qBittorrent Integration (Nice-to-have)

After the core library flow is working, the app will integrate with qBittorrent to track labeled downloads. For completed torrents with a specific label (e.g. `books`), the app will automatically add them to the pending list. The frontend will also surface whether qBittorrent is reachable.

## Technologies

- **Language**: Go
- **Frontend**: Vue
- **Database**: SQLite
- **Runtime**: Docker

## API Reference ✅

Below are the current HTTP endpoints and the JSON structures they expect and return.

---

### Downloads 📥

- **GET /api/downloads**
  - **Description:** List pending downloads
  - **Response:** 200 OK — array of `Download` objects

- **GET /api/downloads/{id}**
  - **Description:** Get a single download by UUID
  - **Response:** 200 OK — single `Download` object

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

---

## Usage ⚙️

- **Start the app locally:**
  ```bash
  go run main.go
  ```
  - Flags available at startup:
    - `-r` — reset (remove) the DB file before starting
    - `-t` — insert test data (implies reset)

---

If you'd like, I can also add example curl commands or small example responses to each endpoint. 💡
