# Archiver

A lightweight CLI tool written in Go to archive development projects while automatically excluding heavy or unnecessary directories.

---

## Features

* Supports **ZIP** and **TAR.GZ** formats.
* Automatically names the archive based on the source folder and outputs it to the current working directory.
* Excludes the `.git` directory by default.
* Supports custom ignore directories via CLI flags or a global `~/.archiverignore` file.
* Automatically generates a **SHA-256 checksum** file alongside the archive.
* Interactive terminal UI with a real-time progress spinner.

---

## Installation

Clone the repository, download the dependencies, and build the binary:

```bash
git clone https://github.com/Treast/go-archiver
cd go-archiver
go build -o archiver main.go
```

## Usage

```bash
./archiver [source_directory] [flags]
```

## Examples

Create a standard ZIP backup of a project:

```bash
./archiver ./my-project
```

Create a TAR.GZ backup with maximum compression (level 9):

```bash
./archiver ./my-project -f tar.gz -l 9
```

Exclude specific directories:

```bash
./archiver ./my-project -i node_modules -i vendor
```

Include the .git directory in the archive:

```bash
./archiver ./my-project --git
```

## Flags

- `-f, --format string`: Archive format: 'zip' or 'tar.gz' (default "zip")
- `-l, --level int`: Compression level: from 1 (fast) to 9 (best) (default 7)
- `--i, --ignore strings`: Additional directories to ignore (can be repeated)
- `--git`: Include the .git directory in the archive