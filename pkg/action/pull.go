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
	"fmt"
	orascontext "github.com/deislabs/oras/pkg/context"
	"github.com/deislabs/oras/pkg/oras"
	"github.com/gosuri/uitable"
	"github.com/opencontainers/go-digest"
	"io"
)

type Pull struct {
	cfg *Configuration
}

func NewPull(cfg *Configuration) *Pull {
	return &Pull{
		cfg: cfg,
	}
}

func (p *Pull) Run(ref string, writer io.Writer) error {
	ctx := orascontext.Background()
	store := p.cfg.OCIStore

	pullOpts := []oras.PullOpt{
		oras.WithContentProvideIngester(store),
		oras.WithAllowedMediaTypes(KnownMediaTypes()),
		oras.WithPullEmptyNameAllowed(),
	}
	// TODO: print progress.
	fmt.Fprintf(writer, "pulling %s ...\n", ref)
	desc, layers, err := oras.Pull(ctx, p.cfg.Resolver, ref, store, pullOpts...)
	if err != nil {
		return err
	}
	store.AddReference(ref, desc)
	if err := store.SaveIndex(); err != nil {
		return err
	}
	var contentDigest digest.Digest
	for _, layer := range layers {
		if layer.MediaType == TorchServeModelContentLayerMediaType {
			contentDigest = layer.Digest
		}
	}
	info, err := store.Info(ctx, contentDigest)
	if err != nil {
		return err
	}

	table := uitable.New()
	table.Wrap = true
	table.AddRow("Ref:", ref)
	table.AddRow("Digest:", info.Digest)
	table.AddRow("Size:", byteCountBinary(info.Size))
	table.AddRow()
	_, err = writer.Write(table.Bytes())
	return nil
}
