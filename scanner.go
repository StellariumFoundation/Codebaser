package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Code file extensions that get syntax-highlighted code blocks.
var codeExtensions = map[string]string{
	".js":      "javascript",
	".mjs":     "javascript",
	".cjs":     "javascript",
	".go":      "go",
	".jsx":     "jsx",
	".ts":      "typescript",
	".tsx":     "tsx",
	".svelte":  "svelte",
	".vue":     "vue",
	".css":     "css",
	".scss":    "scss",
	".sass":    "sass",
	".less":    "less",
	".html":    "html",
	".htm":     "html",
	".xml":     "xml",
	".svg":     "xml",
	".json":    "json",
	".jsonc":   "json",
	".yaml":    "yaml",
	".yml":     "yaml",
	".toml":    "toml",
	".py":      "python",
	".pyw":     "python",
	".rb":      "ruby",
	".erb":     "erb",
	".rs":      "rust",
	".java":    "java",
	".kt":      "kotlin",
	".kts":     "kotlin",
	".swift":   "swift",
	".c":       "c",
	".h":       "c",
	".cpp":     "cpp",
	".hpp":     "cpp",
	".cxx":     "cpp",
	".cc":      "cpp",
	".cs":      "csharp",
	".fs":      "fsharp",
	".php":     "php",
	".pl":      "perl",
	".pm":      "perl",
	".r":       "r",
	".sh":      "bash",
	".bash":    "bash",
	".zsh":     "bash",
	".ps1":     "powershell",
	".bat":     "batch",
	".cmd":     "batch",
	".sql":     "sql",
	".graphql": "graphql",
	".gql":     "graphql",
	".dart":    "dart",
	".lua":     "lua",
	".scala":   "scala",
	".zig":     "zig",
	".nim":     "nim",
	".ex":      "elixir",
	".exs":     "elixir",
	".ml":      "ocaml",
	".mli":     "ocaml",
	".clj":     "clojure",
	".cljs":    "clojure",
	".hs":      "haskell",
	".lhs":     "haskell",
	".cr":      "crystal",
	".tex":     "latex",
	".md":      "markdown",
	".mdx":     "mdx",
	".rmd":     "rmarkdown",
	".dockerfile": "dockerfile",
	".makefile":   "makefile",
	".cmake":   "cmake",
	".tf":      "terraform",
	".sqlite":  "sql",
	".prisma":  "prisma",
}

// Files larger than this (in bytes) are treated as binary instead of text.
const maxTextFileSize = 10 * 1024 * 1024 // 10 MB

// ProgressFunc is called by ScanDirectoryTree to report progress.
// line is the log line to append (e.g. "📁 src/", "  foo.go").
type ProgressFunc func(line string)

// ScanOptions controls which files to include and reports progress.
type ScanOptions struct {
	// Filter, if non-nil, is called for each entry. Return true to include.
	// relPath is relative to rootDir with forward slashes.
	Filter func(relPath string, isDir bool) bool
	// Progress, if non-nil, is called for each entry processed.
	Progress ProgressFunc
}

// ScanDirectoryTree walks rootDir recursively and writes a markdown dump to w.
func ScanDirectoryTree(w io.Writer, rootDir string, opts *ScanOptions) error {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	// Write header
	fmt.Fprintf(w, "# Codebase Dump\n\n")
	fmt.Fprintf(w, "**Source:** `%s`  \n", absRoot)
	fmt.Fprintf(w, "\n---\n\n")

	// Load all .gitignore files in the tree.
	gitignores := loadAllGitignores(absRoot)

	if opts != nil && opts.Progress != nil {
		opts.Progress("Scanning directory tree ...")
	}

	scanDirectory(w, absRoot, absRoot, gitignores, opts)

	return nil
}

// ---------------------------------------------------------------------------
// Directory scanning
// ---------------------------------------------------------------------------

func scanDirectory(w io.Writer, rootDir, currentDir string, gitignores map[string]*GitIgnore, opts *ScanOptions) {
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		fmt.Fprintf(w, "*(error reading directory: %v)*\n\n", err)
		return
	}

	// Sort alphabetically (case-insensitive)
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	relPath, _ := filepath.Rel(rootDir, currentDir)
	relPath = filepath.ToSlash(relPath)

	for _, entry := range entries {
		name := entry.Name()

		// Always skip .git
		if name == ".git" {
			continue
		}

		fullPath := filepath.Join(currentDir, name)
		entryRel := relPath + "/" + name
		if relPath == "." || relPath == "" {
			entryRel = name
		}

		// Check .gitignore
		if isIgnored(fullPath, rootDir, gitignores) {
			continue
		}

		// Check user filter (if provided)
		if opts != nil && opts.Filter != nil && !opts.Filter(entryRel, entry.IsDir()) {
			continue
		}

		if entry.IsDir() {
			if opts != nil && opts.Progress != nil {
				opts.Progress(entryRel + "/")
			}
			fmt.Fprintf(w, "%s/\n\n", entryRel)
			scanDirectory(w, rootDir, fullPath, gitignores, opts)
		} else {
			ext := strings.ToLower(filepath.Ext(name))

			if opts != nil && opts.Progress != nil {
				opts.Progress("  " + name)
			}

			if langID, ok := codeExtensions[ext]; ok {
				// Code file – syntax-highlighted code block
				fmt.Fprintf(w, "%s\n", entryRel)
				fmt.Fprintf(w, "```%s\n", langID)
				dumpFile(w, fullPath)
				fmt.Fprintf(w, "```\n\n")
			} else if canReadAsText(fullPath) {
				// Text file – plain code block
				fmt.Fprintf(w, "%s\n", entryRel)
				fmt.Fprintf(w, "```\n")
				dumpFile(w, fullPath)
				fmt.Fprintf(w, "```\n\n")
			} else {
				// Binary / too-large file – just list the name
				info, err := entry.Info()
				size := ""
				reason := "binary"
				if err == nil {
					size = fmt.Sprintf(" (%s)", formatSize(info.Size()))
					if info.Size() > maxTextFileSize {
						reason = "file too large"
					}
				}
				fmt.Fprintf(w, "%s (%s%s)\n\n", entryRel, reason, size)
			}
		}
	}
}

func dumpFile(w io.Writer, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(w, "[error reading file: %v]", err)
		return
	}
	_, _ = w.Write(data)
	if len(data) > 0 && data[len(data)-1] != '\n' {
		_, _ = w.Write([]byte{'\n'})
	}
}

// ---------------------------------------------------------------------------
// Text / binary detection
// ---------------------------------------------------------------------------

func canReadAsText(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.Size() > maxTextFileSize {
		return false
	}
	return isTextFile(path)
}

func isTextFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 8192)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return false
	}
	buf = buf[:n]
	return !bytes.Contains(buf, []byte{0})
}

func formatSize(size int64) string {
	switch {
	case size < 1024:
		return fmt.Sprintf("%d B", size)
	case size < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
}

// ---------------------------------------------------------------------------
// .gitignore handling
// ---------------------------------------------------------------------------

type GitIgnore struct {
	patterns []gitIgnorePattern
}

type gitIgnorePattern struct {
	pattern  string
	negate   bool
	dirOnly  bool
	anchored bool
}

func loadAllGitignores(rootDir string) map[string]*GitIgnore {
	gitignores := make(map[string]*GitIgnore)

	_ = filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if !d.IsDir() && d.Name() == ".gitignore" {
			if gi, err := LoadGitIgnore(path); err == nil {
				gitignores[filepath.Dir(path)] = gi
			}
		}
		return nil
	})

	return gitignores
}

func isIgnored(path, rootDir string, gitignores map[string]*GitIgnore) bool {
	rel, err := filepath.Rel(rootDir, path)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)

	parts := strings.Split(rel, "/")

	for i := 0; i <= len(parts); i++ {
		parent := filepath.Join(rootDir, filepath.FromSlash(strings.Join(parts[:i], "/")))

		if gi, ok := gitignores[parent]; ok {
			var remaining string
			if i < len(parts) {
				remaining = strings.Join(parts[i:], "/")
			}

			if gi.Matches(remaining) {
				return true
			}
		}
	}

	return false
}

func LoadGitIgnore(path string) (*GitIgnore, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gi := &GitIgnore{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		negate := false
		if strings.HasPrefix(line, "!") {
			negate = true
			line = strings.TrimSpace(line[1:])
			if line == "" {
				continue
			}
		}

		dirOnly := strings.HasSuffix(line, "/")
		if dirOnly {
			line = strings.TrimSuffix(line, "/")
		}

		anchored := strings.HasPrefix(line, "/")
		if anchored {
			line = strings.TrimPrefix(line, "/")
		}

		gi.patterns = append(gi.patterns, gitIgnorePattern{
			pattern:  line,
			negate:   negate,
			dirOnly:  dirOnly,
			anchored: anchored,
		})
	}

	return gi, nil
}

func (gi *GitIgnore) Matches(relPath string) bool {
	relPath = filepath.ToSlash(relPath)
	matched := false

	for _, p := range gi.patterns {
		if p.dirOnly {
			parts := strings.Split(relPath, "/")
			for i, part := range parts {
				if matchGlob(part, p.pattern) {
					if p.anchored && i > 0 {
						break
					}
					goto patternMatched
				}
			}
			continue

		patternMatched:
			if p.negate {
				matched = false
			} else {
				matched = true
			}
			continue
		}

		if matchesGitIgnorePattern(relPath, p) {
			if p.negate {
				matched = false
			} else {
				matched = true
			}
		}
	}

	return matched
}

func matchesGitIgnorePattern(relPath string, p gitIgnorePattern) bool {
	if p.anchored {
		return matchGlob(relPath, p.pattern)
	}

	if matchGlob(relPath, p.pattern) {
		return true
	}

	for _, part := range strings.Split(relPath, "/") {
		if matchGlob(part, p.pattern) {
			return true
		}
	}

	return false
}

// ---------------------------------------------------------------------------
// Glob matching
// ---------------------------------------------------------------------------

func matchGlob(path, pattern string) bool {
	if pattern == "" {
		return path == ""
	}
	if pattern == "**" || pattern == "*" {
		return true
	}

	pathParts := strings.Split(path, "/")
	patternParts := tokenize(pattern)

	return matchGlobParts(pathParts, patternParts)
}

func tokenize(s string) []string {
	if s == "" {
		return []string{""}
	}
	return strings.Split(s, "/")
}

func matchGlobParts(path []string, pattern []string) bool {
	pp := 0

	for pi := 0; pi < len(path) || pp < len(pattern); pi++ {
		if pp >= len(pattern) {
			return false
		}

		if pattern[pp] == "**" {
			for i := pi; i <= len(path); i++ {
				if matchGlobPartsStrict(path[i:], pattern[pp+1:]) {
					return true
				}
			}
			return false
		}

		if pi >= len(path) {
			return false
		}

		if !matchSegment(path[pi], pattern[pp]) {
			return false
		}
		pp++
	}

	return true
}

func matchGlobPartsStrict(path []string, pattern []string) bool {
	if len(pattern) == 0 {
		return len(path) == 0
	}

	pp := 0
	for pi := 0; pi < len(path) || pp < len(pattern); pi++ {
		if pp >= len(pattern) {
			return false
		}

		if pattern[pp] == "**" {
			for i := pi; i <= len(path); i++ {
				if matchGlobPartsStrict(path[i:], pattern[pp+1:]) {
					return true
				}
			}
			return false
		}

		if pi >= len(path) {
			return false
		}

		if !matchSegment(path[pi], pattern[pp]) {
			return false
		}
		pp++
	}

	return true
}

func matchSegment(s, pat string) bool {
	matched, err := filepath.Match(pat, s)
	return err == nil && matched
}
