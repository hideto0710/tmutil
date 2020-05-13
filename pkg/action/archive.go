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
	"encoding/json"
	"fmt"
	"github.com/hideto0710/torchstand/pkg/model"
	"io"
	"io/ioutil"
	"os"
)

const (
	manifestFileName = "MANIFEST.json"
	marInf           = "MAR-INF"
)

type ArchiveOpts struct {
	Tag string
}

type Archive struct {
	cfg *Configuration
}

func NewArchive(cfg *Configuration) *Archive {
	return &Archive{
		cfg: cfg,
	}
}

func (a *Archive) Run(m *model.Model, opts *ArchiveOpts, writer io.Writer) error {
	dir, err := ioutil.TempDir(os.TempDir(), "torchstand-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	fs, err := copyFiles(m, dir)
	if err != nil {
		return err
	}
	archiveFilename := fmt.Sprintf("%s.zip", dir)
	if err := createZip(archiveFilename, func(w *zip.Writer) error {
		for _, f := range fs {
			if err := addFileToZip(w, f); err != nil {
				return err
			}
		}
		f, err := w.Create(fmt.Sprintf("%s/%s", marInf, manifestFileName))
		if err != nil {
			return err
		}
		manifestBytes, err := json.Marshal(m.Manifest())
		if err != nil {
			return err
		}
		_, err = f.Write(manifestBytes)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	content, err := ioutil.ReadFile(archiveFilename)
	if err != nil {
		return err
	}
	return NewImport(a.cfg).Run(opts.Tag, content, m, writer)
}

func copyFiles(m *model.Model, dir string) ([]string, error) {
	var files []string
	mf, err := copyFile(m.ModelFile, dir)
	if err != nil {
		return files, err
	}
	files = append(files, mf)

	sf, err := copyFile(m.SerializedFile, dir)
	if err != nil {
		return files, err
	}
	files = append(files, sf)

	for _, f := range m.ExtraFiles {
		ef, err := copyFile(f, dir)
		if err != nil {
			return files, err
		}
		files = append(files, ef)
	}
	if m.SourceVocab != "" {
		sv, err := copyFile(m.SourceVocab, dir)
		if err != nil {
			return files, err
		}
		files = append(files, sv)
	}
	if m.IsCustomHandler() {
		h, err := copyFile(m.Handler, dir)
		if err != nil {
			return files, err
		}
		files = append(files, h)
	}
	return files, nil
}
