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

More details (API endpoints, schema design, configuration options) will be added as development progresses.
