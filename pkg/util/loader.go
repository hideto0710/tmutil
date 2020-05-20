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

package util

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/hideto0710/torchstand/pkg/types"
)

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) Load(reader *zip.ReadCloser) (*BuiltBytes, *types.Manifest, error) {
	manifest, err := l.loadManifest(reader.File)
	if err != nil {
		return nil, nil, err
	}
	result := &BuiltBytes{}
	result.Config, err = json.Marshal(manifest)
	result.PyTorchModel, err = l.loadPyTorchModel(reader.File, manifest.Model.SerializedFile)
	result.Contents, err = l.loadContents(reader.File, manifest.Model.SerializedFile)

	return result, manifest, err
}

func (l *Loader) loadManifest(files []*zip.File) (*types.Manifest, error) {
	var result *types.Manifest
	for _, f := range files {
		if f.Name == MarFilePath {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			if err := json.NewDecoder(rc).Decode(&result); err != nil {
				return nil, err
			}
		}
	}
	if result == nil {
		return nil, fmt.Errorf("invalid TorchServe model archive, %s not found", MarFilePath)
	}
	return result, nil
}

func (l *Loader) loadPyTorchModel(files []*zip.File, modelFilename string) ([]byte, error) {
	for _, f := range files {
		if f.Name == MarFilePath {
			continue
		}
		if f.Name == modelFilename {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			fileBytes, err := ioutil.ReadAll(rc)
			if err := rc.Close(); err != nil {
				return nil, err
			}
			return fileBytes, nil
		}
	}
	return nil, nil
}

func (l *Loader) loadContents(files []*zip.File, modelFilename string) ([]byte, error) {
	var contentBuffer bytes.Buffer
	tw := tar.NewWriter(&contentBuffer)
	defer tw.Close()

	for _, f := range files {
		if f.Name == MarFilePath {
			continue
		}
		if f.Name == modelFilename {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		fileBytes, err := ioutil.ReadAll(rc)
		info := f.FileInfo()
		mode, err := strconv.ParseInt(fmt.Sprintf("%o", info.Mode().Perm()), 10, 64)
		if err != nil {
			return nil, err
		}
		hdr := &tar.Header{
			Name: info.Name(),
			Mode: mode,
			Size: info.Size(),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := tw.Write(fileBytes); err != nil {
			return nil, err
		}
		if err := rc.Close(); err != nil {
			return nil, err
		}
	}
	return contentBuffer.Bytes(), nil
}
