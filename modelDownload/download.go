package modelDownload

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func downloadFile(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching %s: %v\n", url, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Bad status downloading %s: %s\n", url, resp.Status)
		return fmt.Errorf("bad status downloading %s: %s", url, resp.Status)
	}

	out, err := os.Create(path)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", path, err)
		return err
	}
	defer out.Close()

	total := resp.ContentLength
	pr := &progressReader{
		Reader: resp.Body,
		Total:  total,
		Path:   path,
	}

	if _, err = io.Copy(out, pr); err != nil {
		fmt.Printf("\nError downloading %s: %v\n", path, err)
		return err
	}

	fmt.Printf("\nDownload of %s complete\n", path)
	return nil
}

type progressReader struct {
	Reader     io.Reader
	Total      int64
	Downloaded int64
	Path       string
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Downloaded += int64(n)

	if pr.Total > 0 {
		percent := float64(pr.Downloaded) * 100 / float64(pr.Total)
		fmt.Printf("\rDownloading %s: %.1f%%", pr.Path, percent)
	} else {
		fmt.Printf("\rDownloading %s: %d bytes", pr.Path, pr.Downloaded)
	}

	return n, err
}
