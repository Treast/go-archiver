package backup

import (
	"archive/tar"
	"archive/zip"
	"compress/flate"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func CreateZip(source string, w io.Writer, level int, ignore map[string]bool, progress func(string)) error {
	archive := zip.NewWriter(w)
	defer archive.Close()

	if level != flate.DefaultCompression {
		archive.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
			return flate.NewWriter(out, level)
		})
	}

	return walkAndArchive(source, ignore, progress, func(path, relPath string, d fs.DirEntry) error {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		zipFile, err := archive.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(zipFile, file)
		return err
	})
}

func CreateTarGz(source string, w io.Writer, level int, ignore map[string]bool, progress func(string)) error {
	gw, err := gzip.NewWriterLevel(w, level)
	if err != nil {
		return err
	}
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return walkAndArchive(source, ignore, progress, func(path, relPath string, d fs.DirEntry) error {
		info, err := d.Info()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)
		return err
	})
}

func walkAndArchive(source string, ignore map[string]bool, progress func(string), addToArchive func(path, relPath string, d fs.DirEntry) error) error {
	return filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Gestion des dossiers ignorés
		if d.IsDir() && ignore[d.Name()] {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		if progress != nil {
			progress(relPath)
		}

		return addToArchive(path, relPath, d)
	})
}
