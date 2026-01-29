# JSON-Excel Editor

A modern web application that transforms structured JSON into a spreadsheet-like interface, allowing you to edit data visually with support for nested tables (Table-inside-Table). It ensures data integrity through strict validation enforcing a maximum number of keys per object and 1 level of nesting, includes a premium Glassmorphism UI with full Upload and Download capabilities for seamless file management. It supports multiple users with unique session IDs, storing separate files per user.

## Features

- **Visual JSON Editing**: Edit JSON data in a spreadsheet-like interface.
- **Nested Tables**: Support for Table-inside-Table (1 level of nesting) with add/delete rows in nested tables.
- **Create from Scratch**: Start with an empty JSON and add columns (primitive or nested) and rows.
- **Advanced Editor**: 3-level deep nesting for complex JSON structures (/advanced).
- **Undo Functionality**: One-step undo for accidental changes.
- **File Access**: Access saved files via URL parameters (?id=<fileid>) for resuming work.
- **Validation**: Enforces maximum keys per object (configurable) and nesting levels.
- **Glassmorphism UI**: Modern, premium user interface.
- **File Management**: Upload and download JSON files with size limits and normalization.
- **Multi-User Support**: Generates unique User ID per session/browser, stores separate files in data/ folder.
- **Concurrency Safe**: Uses RWMutex to prevent file corruption during simultaneous saves.
- **Automatic Cleanup**: Background process deletes files unmodified for 24+ hours every hour.
- **Logging**: Structured logging with rotation to prevent log files from growing indefinitely.

## Setup

1. Clone the repository.
2. Copy `.env-example` to `.env` and configure as needed.
3. Run `go mod tidy` to install dependencies.
4. Run `go run main.go handlers.go` to start the server.
5. Open `http://localhost:<PORT>` in your browser (defaults to demo).
6. Access saved files via `http://localhost:<PORT>?id=<fileid>`.
7. For advanced 3-level nesting editor, go to `http://localhost:<PORT>/advanced`.

## Configuration

- `PORT`: Server port (default 8080)
- `MAX_KEYS`: Maximum keys per object (default 20)
- `MAX_UPLOAD_SIZE_MB`: Maximum upload file size in MB (default 1)

## API Endpoints

- `GET /api/data?id=<id>`: Retrieve data for the given ID.
- `POST /api/data?id=<id>`: Save data for the given ID.
- `POST /api/upload`: Upload a JSON file and get a new ID.
- `GET /api/download?id=<id>`: Download data for the given ID.

## Technologies

- Go (Backend)
- HTML/CSS/JavaScript (Frontend)