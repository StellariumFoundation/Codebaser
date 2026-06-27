package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	var (
		dirPath    string
		outputPath string
		excludes   []string
	)

	rootCmd := &cobra.Command{
		Use:   "codebaser [directory]",
		Short: "Export a codebase into a single Markdown file",
		Long:  "Codebaser recursively scans a directory tree and exports the entire codebase into a single Markdown file with syntax-highlighted code blocks.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				dirPath = args[0]
			}
			if dirPath == "" {
				dirPath = "."
			}
			if outputPath == "" {
				outputPath = "codebase-dump.md"
			}

			f, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer f.Close()

			var filter func(string, bool) bool
			if len(excludes) > 0 {
				excludeSet := make(map[string]bool)
				for _, e := range excludes {
					excludeSet[strings.TrimPrefix(e, "/")] = true
				}
				filter = func(relPath string, isDir bool) bool {
					parts := strings.Split(relPath, "/")
					for i := 1; i <= len(parts); i++ {
						prefix := strings.Join(parts[:i], "/")
						if excludeSet[prefix] {
							return false
						}
					}
					return true
				}
			}

			opts := &ScanOptions{
				Filter: filter,
				Progress: func(line string) {
					fmt.Println(line)
				},
			}

			fmt.Printf("Scanning %s ...\n", dirPath)
			if err := ScanDirectoryTree(f, dirPath, opts); err != nil {
				return fmt.Errorf("scan failed: %w", err)
			}

			absPath, _ := os.Executable()
			_ = absPath
			fmt.Printf("Done! Written to %s\n", outputPath)
			return nil
		},
	}

	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "codebase-dump.md", "Output file path")
	rootCmd.Flags().StringArrayVarP(&excludes, "exclude", "e", nil, "Paths to exclude (can be specified multiple times)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
