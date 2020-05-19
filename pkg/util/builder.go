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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/hideto0710/torchstand/pkg/types"
)

type Builder struct {
	modelFile *types.TorchServeModelfile
}

type BuiltBytes struct {
	Config       []byte
	PyTorchModel []byte
	Contents     []byte
}

func NewBuilder(modelFile *types.TorchServeModelfile) *Builder {
	return &Builder{modelFile: modelFile}
}

func (b *Builder) Build() (*BuiltBytes, error) {
	result := &BuiltBytes{}
	var err error
	result.Config, err = b.buildConfig()
	result.PyTorchModel, err = b.buildPyTorchModel()
	result.Contents, err = b.buildContents()
	return result, err
}

func (b *Builder) buildConfig() ([]byte, error) {
	return json.Marshal(b.modelFile.Manifest())
}

func (b *Builder) buildPyTorchModel() ([]byte, error) {
	return ioutil.ReadFile(b.modelFile.SerializedFile)
}

func (b *Builder) buildContents() ([]byte, error) {
	var contentBuffer bytes.Buffer
	tw := tar.NewWriter(&contentBuffer)
	defer tw.Close()
	if err := writeToTar(b.modelFile.ModelFile, tw); err != nil {
		return nil, err
	}
	for _, name := range b.modelFile.ExtraFiles {
		if err := writeToTar(name, tw); err != nil {
			return nil, err
		}
	}
	if b.modelFile.SourceVocab != "" {
		if err := writeToTar(b.modelFile.SourceVocab, tw); err != nil {
			return nil, err
		}
	}
	if b.modelFile.IsCustomHandler() {
		if err := writeToTar(b.modelFile.Handler, tw); err != nil {
			return nil, err
		}
	}
	return contentBuffer.Bytes(), nil
}

func writeToTar(filename string, writer *tar.Writer) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	fileBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		return err
	}
	mode, err := strconv.ParseInt(fmt.Sprintf("%o", info.Mode().Perm()), 10, 64)
	if err != nil {
		return err
	}
	// FIXME: ignore tree structure, may overwritten.
	hdr := &tar.Header{
		Name: info.Name(),
		Mode: mode,
		Size: info.Size(),
	}
	if err := writer.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := writer.Write(fileBytes); err != nil {
		return err
	}
	return nil
}
