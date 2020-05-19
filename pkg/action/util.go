/*
Copyright © 2020 HIDETO INAMURA <h.inamura0710@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package action

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/go-units"
)

// see https://github.com/helm/helm/blob/v3.2.1/internal/experimental/registry/util.go
// byteCountBinary produces a human-readable file size
func byteCountBinary(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

// shortDigest returns first 7 characters of a sha256 digest
func shortDigest(digest string) string {
	if len(digest) == 64 {
		return digest[:7]
	}
	return digest
}

// timeAgo returns a human-readable timestamp representing time that has passed
func timeAgo(t time.Time) string {
	return units.HumanDuration(time.Now().UTC().Sub(t))
}

func copyFile(src string, destDir string) (string, error) {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return "", err
	}
	target := fmt.Sprintf("%s/%s", destDir, filepath.Base(src))
	err = ioutil.WriteFile(target, input, 0644)
	return target, err
}

func createZip(filename string, fn func(w *zip.Writer) error) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()
	w := zip.NewWriter(newZipFile)
	defer w.Close()
	return fn(w)
}

func addFileToZip(w *zip.Writer, src string, dest string) error {
	fileToZip, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fileToZip.Close()
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	if dest != "" {
		header.Name = dest
	}
	header.Method = zip.Deflate
	writer, err := w.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}
