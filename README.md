# Codebaser

A desktop GUI tool that recursively scans a directory tree and exports the entire codebase into a single Markdown file. Built with [Gio](https://gioui.org/) (Go GUI framework).

## Features

- **Visual file tree** — Browse the selected directory and see every file/folder listed with checkboxes
- **Selective export** — Uncheck any file or directory to exclude it from the dump. Unchecking a folder automatically unchecks all its children
- **Respects `.gitignore`** — Files matching `.gitignore` rules are hidden from the tree and skipped during export
- **Syntax-highlighted code blocks** — 70+ language extensions detected (`.js`, `.py`, `.go`, `.rs`, `.tsx`, `.css`, etc.), rendered with the correct markdown fenced code identifier
- **Binary/text detection** — Text files have their contents included; binary files and files >10 MB are listed by name only
- **Native folder picker** — Modern Windows folder selection dialog (address bar, Quick Access, search)
- **Dark theme** — Easy on the eyes

## Requirements

- **Windows** (macOS/Linux fallbacks for the folder picker are included but untested)
- **Go 1.21+** to build from source

## Build

```
cd codebaser
go build -ldflags="-H=windowsgui" -o codebaser.exe .
```

Or double-click `build.bat`.

Run the resulting `codebaser.exe` — no console window will appear.

## Usage

1. **Select a directory** — type a path or click **Browse** to use the native folder picker
2. **Choose files** — the file tree loads automatically. Uncheck any files or directories you want to exclude
3. **Set output path** — defaults to `codebase-dump.md` in the current directory. Click **Folder** to pick a different location
4. **Click Generate** — the scan runs asynchronously with a live progress log
5. **Done** — a success alert shows the output path. Click **Clear** to reset

## Output format

The generated Markdown file has this structure:

```markdown
# Codebase Dump

**Source:** `C:\Projects\my-app`

---

## 📁 `src/`

### `src/index.js`
\`\`\`javascript
console.log("Hello");
\`\`\`

### `src/data.json`
\`\`\`
{ "key": "value" }
\`\`\`

- `logo.png` _(binary (12.5 KB))_

## 📁 `docs/`

### `docs/readme.md`
\`\`\`
# Project docs
\`\`\`
```

## Supported languages (code highlighting)

JavaScript, TypeScript, JSX, TSX, Go, Python, Ruby, Rust, Java, Kotlin, Swift, C, C++, C#, F#, PHP, Dart, Lua, Scala, Zig, Nim, Elixir, OCaml, Clojure, Haskell, Crystal, R, Perl, Shell, PowerShell, Batch, SQL, GraphQL, HTML, CSS, SCSS, SASS, Less, Vue, Svelte, Markdown, MDX, LaTeX, JSON, YAML, TOML, XML, SVG, Terraform, Prisma, and more.

## How it works

1. **File tree** — the tool walks the selected directory (sorted alphabetically, respecting `.gitignore`) and builds a flat list of entries with indentation depth
2. **User filtering** — unchecked entries are collected into an exclusion set; when scanning, any path whose ancestor is in the exclusion set is skipped
3. **Scanning** — `ScanDirectoryTree` in `scanner.go` walks the directory, formats each file as a Markdown section, and writes to the output file

## Project structure

```
codebaser/
├── main.go       — Proton GUI (window, widgets, event handling)
├── scanner.go    — Core scanning engine (gitignore, glob matching, text detection)
├── build.bat     — Windows build script
├── go.mod        — Go module definition
├── go.sum        — Dependency checksums
└── README.md     — This file
```

## Dependencies

- [gioui.org](https://gioui.org/) — Pure-Go GUI framework
- [github.com/harry1453/go-common-file-dialog](https://github.com/harry1453/go-common-file-dialog) — Native Windows folder picker
