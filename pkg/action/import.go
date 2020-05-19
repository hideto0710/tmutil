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
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"io"

	orascontext "github.com/deislabs/oras/pkg/context"
	"github.com/gosuri/uitable"
	"github.com/hideto0710/torchstand/pkg/types"
	"github.com/hideto0710/torchstand/pkg/util"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type BuiltDescriptor struct {
	Config       *ocispec.Descriptor
	Manifest     *ocispec.Descriptor
	PyTorchModel *ocispec.Descriptor
	Content      *ocispec.Descriptor
}

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
	store := p.cfg.OCIStore

	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	bb, torchServeManifest, err := util.NewLoader().Load(zipReader)
	if err != nil {
		return err
	}

	descs, err := storeAll(ctx, bb, torchServeManifest, p.cfg.StoreBlob)
	if err != nil {
		return err
	}

	if err := store.LoadIndex(); err != nil {
		return err
	}
	store.AddReference(argRef, *descs.Manifest)
	if err := store.SaveIndex(); err != nil {
		return err
	}

	table := uitable.New()
	table.Wrap = true
	table.AddRow("Ref:", argRef)
	table.AddRow("Digest:", descs.Manifest.Digest.Hex())
	table.AddRow("Model Digest:", descs.PyTorchModel.Digest.Hex())
	table.AddRow("Size:",
		byteCountBinary(descs.PyTorchModel.Size+descs.Content.Size))
	table.AddRow()
	_, err = writer.Write(table.Bytes())
	return err
}

func storeAll(
	ctx context.Context,
	builtBytes *util.BuiltBytes,
	torchServeManifest *types.Manifest,
	store func(ctx context.Context, name string, mediaType string, bytes []byte) (*ocispec.Descriptor, error),
) (*BuiltDescriptor, error) {

	logger := zap.L().Named("store")

	var err error
	result := &BuiltDescriptor{}

	// store config
	result.Config, err = store(ctx, "", TorchServeModelConfigMediaType, builtBytes.Config)
	if err != nil {
		return nil, err
	}
	logger.Debug("stored",
		zap.String("mediaType", result.Config.MediaType),
		zap.String("digest", result.Config.Digest.String()))

	// store pytorch model
	result.PyTorchModel, err = store(ctx, torchServeManifest.Model.SerializedFile, PyTorchModelMediaType, builtBytes.PyTorchModel)
	if err != nil {
		return nil, err
	}
	logger.Debug("stored",
		zap.String("mediaType", result.PyTorchModel.MediaType),
		zap.String("digest", result.PyTorchModel.Digest.String()))

	// store content
	result.Content, err = store(ctx, torchServeManifest.Model.ModelName, TorchServeModelContentLayerMediaType, builtBytes.Contents)
	if err != nil {
		return nil, err
	}
	logger.Debug("stored",
		zap.String("mediaType", result.Content.MediaType),
		zap.String("digest", result.Content.Digest.String()))

	// store manifest
	manifest := ocispec.Manifest{
		Versioned: specs.Versioned{SchemaVersion: 2},
		Config:    *result.Config,
		Layers:    []ocispec.Descriptor{*result.PyTorchModel, *result.Content},
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	result.Manifest, err = store(ctx, "", ocispec.MediaTypeImageManifest, manifestBytes)
	if err != nil {
		return nil, err
	}
	logger.Debug("stored",
		zap.String("mediaType", result.Manifest.MediaType),
		zap.String("digest", result.Manifest.Digest.String()))

	return result, nil
}
