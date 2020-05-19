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
	"io"

	orascontext "github.com/deislabs/oras/pkg/context"
	"github.com/gosuri/uitable"
	"github.com/hideto0710/torchstand/pkg/types"
	"github.com/hideto0710/torchstand/pkg/util"
)

type ArchiveOpts struct {
	Tag string
}

type Build struct {
	cfg *Configuration
}

func NewBuild(cfg *Configuration) *Build {
	return &Build{
		cfg: cfg,
	}
}

func (a *Build) Run(m *types.TorchServeModelfile, opts *ArchiveOpts, writer io.Writer) error {
	ctx := orascontext.Background()
	store := a.cfg.OCIStore
	bb, err := util.NewBuilder(m).Build()

	descs, err := storeAll(ctx, bb, m.Manifest(), a.cfg.StoreBlob)
	if err != nil {
		return err
	}

	if err := store.LoadIndex(); err != nil {
		return err
	}
	store.AddReference(opts.Tag, *descs.Manifest)
	if err := store.SaveIndex(); err != nil {
		return err
	}

	table := uitable.New()
	table.Wrap = true
	table.AddRow("Ref:", opts.Tag)
	table.AddRow("Digest:", descs.Manifest.Digest.Hex())
	table.AddRow("Model Digest:", descs.PyTorchModel.Digest.Hex())
	table.AddRow("Size:",
		byteCountBinary(descs.PyTorchModel.Size+descs.Content.Size))
	table.AddRow()
	_, err = writer.Write(table.Bytes())
	return err
}
