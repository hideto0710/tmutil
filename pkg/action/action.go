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
	"context"
	"encoding/json"
	"fmt"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/remotes"
	orascontent "github.com/deislabs/oras/pkg/content"
	"github.com/hideto0710/torchstand/pkg/model"
	"github.com/hideto0710/torchstand/pkg/path"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type Configuration struct {
	Resolver remotes.Resolver
	OCIStore *orascontent.OCIStore
	Path     *path.TorchstandPath
}

func (cfg *Configuration) StoreBlob(ctx context.Context, name string, mediaType string, bytes []byte) (*ocispec.Descriptor, error) {
	store := cfg.OCIStore
	writer, err := store.Writer(ctx, content.WithRef(digest.FromBytes(bytes).Hex()))
	if err != nil {
		return nil, err
	}
	_, err = writer.Write(bytes)
	if err != nil {
		return nil, err
	}
	if err := writer.Commit(ctx, 0, writer.Digest()); err != nil {
		if !errdefs.IsAlreadyExists(err) {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	var annotations map[string]string
	if name != "" {
		annotations = map[string]string{
			ocispec.AnnotationTitle: name,
		}
	}
	return &ocispec.Descriptor{
		MediaType:   mediaType,
		Digest:      writer.Digest(),
		Size:        int64(len(bytes)),
		Annotations: annotations,
	}, nil
}

func (cfg *Configuration) FetchReference(ctx context.Context, argRef string) (*model.Ref, error) {
	store := cfg.OCIStore
	if err := store.LoadIndex(); err != nil {
		return nil, err
	}
	for ref, manifestDesc := range store.ListReferences() {
		if ref == argRef {
			return cfg.SummarizeModel(ctx, ref, manifestDesc)
		}
	}
	return &model.Ref{
		Name:   argRef,
		Exists: false,
	}, nil
}

func (cfg *Configuration) SummarizeModel(ctx context.Context, ref string, manifestDesc ocispec.Descriptor) (*model.Ref, error) {
	store := cfg.OCIStore
	result := &model.Ref{Name: ref}
	result.Manifest = manifestDesc
	result.Exists = true
	reader, err := store.ReaderAt(ctx, manifestDesc)
	if err != nil {
		return result, err
	}
	manifestBytes := make([]byte, manifestDesc.Size)
	_, err = reader.ReadAt(manifestBytes, 0)
	if err != nil {
		return result, err
	}
	var manifest ocispec.Manifest
	if err = json.Unmarshal(manifestBytes, &manifest); err != nil {
		return result, err
	}
	result.Config = manifest.Config
	numLayers := len(manifest.Layers)
	if numLayers != 1 {
		return result, fmt.Errorf("manifest does not contain exactly 1 layer (total: %d)", numLayers)
	}
	var contentDesc *ocispec.Descriptor
	for _, layer := range manifest.Layers {
		if layer.MediaType == TorchServeModelContentLayerMediaType {
			contentDesc = &layer
		}
	}
	if contentDesc == nil {
		return result, fmt.Errorf("manifest does not contain a layer with mediatype %s", TorchServeModelContentLayerMediaType)
	}
	if contentDesc.Size == 0 {
		return result, fmt.Errorf("manifest layer with mediatype %s is of size 0", TorchServeModelContentLayerMediaType)
	}
	result.Content = []ocispec.Descriptor{*contentDesc}
	info, err := store.Info(ctx, contentDesc.Digest)
	if err != nil {
		return result, err
	}
	result.Size = info.Size
	result.Digest = info.Digest
	result.CreatedAt = info.CreatedAt
	return result, nil
}
