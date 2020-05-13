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
	"encoding/json"
	"github.com/gosuri/uitable"
	"io"

	orascontext "github.com/deislabs/oras/pkg/context"
	"github.com/hideto0710/torchstand/pkg/model"
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

func (p *Import) Run(ref string, contentBytes []byte, config *model.Model, writer io.Writer) error {
	ctx := orascontext.Background()
	logger := zap.L().Named("import")
	store := p.cfg.OCIStore
	// TODO: check exist or not

	// store config
	configBytes, err := json.Marshal(config)
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

	// store content
	contentDesc, err := p.cfg.StoreBlob(ctx, config.ModelName, TorchServeModelContentLayerMediaType, contentBytes)
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
		Layers:    []ocispec.Descriptor{*contentDesc},
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
	store.AddReference(ref, *manifestDesc)
	if err := store.SaveIndex(); err != nil {
		return err
	}

	table := uitable.New()
	table.Wrap = true
	table.AddRow("Ref:", ref)
	table.AddRow("Digest:", contentDesc.Digest.Hex())
	table.AddRow("Size:", byteCountBinary(contentDesc.Size))
	table.AddRow()
	_, err = writer.Write(table.Bytes())
	return err
}
