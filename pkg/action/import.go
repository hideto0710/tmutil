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
	"archive/tar"
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	orascontext "github.com/deislabs/oras/pkg/context"
	"github.com/gosuri/uitable"
	"github.com/hideto0710/torchstand/pkg/types"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"go.uber.org/zap"
)

type Import struct {
	cfg *Configuration
}

type ImportOpt struct {
	ModelName string
}

func NewImport(cfg *Configuration) *Import {
	return &Import{
		cfg: cfg,
	}
}

func (p *Import) Run(argRef string, filePath string, writer io.Writer) error {
	ctx := orascontext.Background()
	logger := zap.L().Named("import")
	store := p.cfg.OCIStore

	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	var torchServeManifest *types.Manifest
	for _, f := range zipReader.File {
		if f.Name == marFilePath {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			if err := json.NewDecoder(rc).Decode(&torchServeManifest); err != nil {
				return err
			}
		}
	}
	if torchServeManifest == nil {
		return fmt.Errorf("invalid TorchServe model archive, %s not found", marFilePath)
	}
	pytorchModelBytes, contentBytes, err := loadZip(zipReader.File, torchServeManifest.Model.SerializedFile)
	if err != nil {
		return err
	}

	// store config
	configBytes, err := json.Marshal(torchServeManifest)
	if err != nil {
		return err
	}
	configDesc, err := p.cfg.StoreBlob(ctx, "", TorchServeModelConfigMediaType, configBytes)
	if err != nil {
		return err
	}
	logger.Debug("stored",
		zap.String("mediaType", configDesc.MediaType),
		zap.String("digest", configDesc.Digest.String()))

	// store pytorch model
	pytorcchModelDesc, err := p.cfg.StoreBlob(ctx, torchServeManifest.Model.SerializedFile, PyTorchModelMediaType, pytorchModelBytes)
	if err != nil {
		return err
	}
	logger.Debug("stored",
		zap.String("mediaType", pytorcchModelDesc.MediaType),
		zap.String("digest", pytorcchModelDesc.Digest.String()))

	// store content
	contentDesc, err := p.cfg.StoreBlob(ctx, torchServeManifest.Model.ModelName, TorchServeModelContentLayerMediaType, contentBytes)
	if err != nil {
		return err
	}
	logger.Debug("stored",
		zap.String("mediaType", contentDesc.MediaType),
		zap.String("digest", contentDesc.Digest.String()))

	// store manifest
	manifest := ocispec.Manifest{
		Versioned: specs.Versioned{SchemaVersion: 2},
		Config:    *configDesc,
		Layers:    []ocispec.Descriptor{*pytorcchModelDesc, *contentDesc},
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	manifestDesc, err := p.cfg.StoreBlob(ctx, "", ocispec.MediaTypeImageManifest, manifestBytes)
	if err != nil {
		return err
	}
	logger.Debug("stored",
		zap.String("mediaType", manifestDesc.MediaType),
		zap.String("digest", manifestDesc.Digest.String()))

	if err := store.LoadIndex(); err != nil {
		return err
	}
	store.AddReference(argRef, *manifestDesc)
	if err := store.SaveIndex(); err != nil {
		return err
	}

	table := uitable.New()
	table.Wrap = true
	table.AddRow("Ref:", argRef)
	table.AddRow("Digest:", manifestDesc.Digest.Hex())
	table.AddRow("Model Digest:", pytorcchModelDesc.Digest.Hex())
	table.AddRow("Size:",
		byteCountBinary(pytorcchModelDesc.Size+contentDesc.Size))
	table.AddRow()
	_, err = writer.Write(table.Bytes())
	return err
}

func loadZip(files []*zip.File, modelFileName string) ([]byte, []byte, error) {
	var contentBuffer bytes.Buffer
	var pytorchModelBytes []byte

	// gzw := gzip.NewWriter(&contentBuffer)
	// defer gzw.Close()

	tw := tar.NewWriter(&contentBuffer)
	defer tw.Close()

	for _, f := range files {
		if f.Name == marFilePath {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, nil, err
		}
		fileBytes, err := ioutil.ReadAll(rc)
		if f.Name == modelFileName {
			pytorchModelBytes = fileBytes
			if err := rc.Close(); err != nil {
				return nil, nil, err
			}
			continue
		}

		info := f.FileInfo()
		mode, err := strconv.ParseInt(fmt.Sprintf("%o", info.Mode().Perm()), 10, 64)
		if err != nil {
			return nil, nil, err
		}
		hdr := &tar.Header{
			Name: info.Name(),
			Mode: mode,
			Size: info.Size(),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, nil, err
		}
		if _, err := tw.Write(fileBytes); err != nil {
			return nil, nil, err
		}
		if err := rc.Close(); err != nil {
			return nil, nil, err
		}
	}
	return pytorchModelBytes, contentBuffer.Bytes(), nil
}
