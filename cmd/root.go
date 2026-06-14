package cmd

import (
	"crypto/sha256"
	"fmt"
	"go-archiver/pkg/backup"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	format      string
	level       int
	extraIgnore []string
	git         bool
	recursive   bool
)

var rootCmd = &cobra.Command{
	Use:   "archiver [folder]",
	Short: "Archiver compress and save your project",
	Long:  `A quick CLI tool for archiving your projects while excluding unnecessary folders.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceDir := args[0]
		ignoreList := backup.LoadIgnoreRules(extraIgnore, git)

		pterm.DefaultHeader.WithFullWidth().Println("⚡ Archiver")
		fmt.Println()

		archiveFolder := func(dir string) error {
			cleanedSource := filepath.Clean(dir)
			folderName := filepath.Base(cleanedSource)

			if folderName == "." || folderName == "/" {
				return fmt.Errorf("unable to determine a valid archive name for the folder '%s', please specify a more precise path", sourceDir)
			}

			ext := ".zip"
			if strings.ToLower(format) == "tar.gz" || strings.ToLower(format) == "tgz" {
				ext = ".tar.gz"
			}

			destArchive := folderName + ext

			pterm.Info.Printf("Source      : %s\n", pterm.Cyan(dir))
			pterm.Info.Printf("Destination : %s\n", pterm.Cyan(destArchive))
			pterm.Info.Printf("Format      : %s (Level %d)\n", pterm.Magenta(format), level)
			fmt.Println()

			outFile, err := os.Create(destArchive)
			if err != nil {
				return fmt.Errorf("unable to create archive : %w", err)
			}
			defer outFile.Close()

			hash := sha256.New()
			teeWriter := io.MultiWriter(outFile, hash)

			spinner, _ := pterm.DefaultSpinner.Start("Archiving initialization...")
			onProgress := func(currentFile string) {
				if len(currentFile) > 40 {
					currentFile = "..." + currentFile[len(currentFile)-37:]
				}
				spinner.UpdateText(fmt.Sprintf("Adding : %s", pterm.Gray(currentFile)))
			}

			fmt.Printf("Creating archive [%s] (Compression level %d)...\n", format, level)

			switch strings.ToLower(format) {
			case "zip":
				err = backup.CreateZip(dir, teeWriter, level, ignoreList, onProgress)
			case "tar.gz", "tgz":
				err = backup.CreateTarGz(dir, teeWriter, level, ignoreList, onProgress)
			default:
				spinner.Fail("Unsupported format")
				return fmt.Errorf("unsupported format : %s", format)
			}

			if err != nil {
				spinner.Fail("Error during compression")
				return fmt.Errorf("error during compression : %w", err)
			}

			spinner.Success("Files successfully compressed!")

			checksumPath := destArchive + ".sha256"
			checksumContent := fmt.Sprintf("%x  %s\n", hash.Sum(nil), destArchive)

			err = os.WriteFile(checksumPath, []byte(checksumContent), 0644)
			if err != nil {
				return fmt.Errorf("unable to save checksum file : %w", err)
			}

			fmt.Println()
			pterm.DefaultBox.
				WithTitle(pterm.Green("✨ SUCCESSFUL ARCHIVE ✨")).
				WithTitleBottomCenter().
				WithBoxStyle(pterm.NewStyle(pterm.FgGreen)).
				Printf("Archive created  : %s\nChecksum SHA256: %s", pterm.Bold.Sprint(destArchive), pterm.Bold.Sprint(checksumPath))
			fmt.Println()

			return nil
		}

		if recursive {
			entries, err := os.ReadDir(sourceDir)
			if err != nil {
				return fmt.Errorf("unable to read the root folder : %w", err)
			}

			var archivedCount int
			for _, entry := range entries {
				if entry.IsDir() {
					if ignoreList[entry.Name()] {
						continue
					}

					targetPath := filepath.Join(sourceDir, entry.Name())
					if err := archiveFolder(targetPath); err != nil {
						pterm.Error.Printf("Error processing %s: %v\n", entry.Name(), err)
						return err
					}
					archivedCount++
				}
			}

			if archivedCount == 0 {
				pterm.Warning.Println("No direct subfolders found to archive.")
			} else {
				pterm.Success.Printf("Batch processing complete. %d files have been archived.\n", archivedCount)
			}
			return nil
		}

		return archiveFolder(sourceDir)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&format, "format", "f", "zip", "Archive format: 'zip' or 'tar.gz'")
	rootCmd.Flags().IntVarP(&level, "level", "l", 7, "Compression ratio: from 1 to 9")
	rootCmd.Flags().StringSliceVarP(&extraIgnore, "ignore", "i", []string{}, "Files to ignore (e.g., -i node_modules -i vendor)")
	rootCmd.Flags().BoolVar(&git, "git", false, "Include the .git folder in the archive (ignored by default)")
	rootCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Apply the command to all direct subfolders")
}
