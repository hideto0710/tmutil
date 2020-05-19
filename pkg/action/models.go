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
	"go.uber.org/zap"
)

type Models struct {
	cfg *Configuration
}

func NewModels(cfg *Configuration) *Models {
	return &Models{
		cfg: cfg,
	}
}

func (p *Models) Run(writer io.Writer) error {
	ctx := orascontext.Background()
	logger := zap.L().Named("models")
	store := p.cfg.OCIStore

	if err := store.LoadIndex(); err != nil {
		return err
	}
	var modelReferences []*types.Ref
	for ref, desc := range store.ListReferences() {
		if a, err := p.cfg.SummarizeModel(ctx, ref, desc); err != nil {
			logger.Warn("invalid reference", zap.String("ref", ref), zap.String("error", err.Error()))
		} else {
			modelReferences = append(modelReferences, a)
		}
	}

	table := uitable.New()
	table.AddRow("REF", "DIGEST", "CREATED", "TOTAL")
	table.MaxColWidth = 60
	table.Separator = "\t\t"
	for _, r := range modelReferences {
		table.AddRow(r.Name,
			shortDigest(r.Digest.Hex()),
			timeAgo(r.CreatedAt),
			byteCountBinary(r.Size))
	}
	table.AddRow()
	_, err := writer.Write(table.Bytes())
	return err
}
