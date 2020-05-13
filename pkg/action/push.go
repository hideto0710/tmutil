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
	"io"

	orascontext "github.com/deislabs/oras/pkg/context"
	"github.com/deislabs/oras/pkg/oras"
	"github.com/gosuri/uitable"
)

type Push struct {
	cfg *Configuration
}

func NewPush(cfg *Configuration) *Push {
	return &Push{
		cfg: cfg,
	}
}

func (p *Push) Run(argRef string, writer io.Writer) error {
	ctx := orascontext.Background()
	store := p.cfg.OCIStore

	ref, err := p.cfg.FetchReference(ctx, argRef)
	if err != nil {
		return err
	}
	if !ref.Exists {
		return fmt.Errorf("model not found: %s", ref.Name)
	}
	// TODO: print progress.
	fmt.Fprintf(writer, "pushing %s ...\n", ref.Name)
	_, err = oras.Push(ctx, p.cfg.Resolver, ref.Name, store, ref.Content, oras.WithConfig(ref.Config))
	if err != nil {
		return err
	}

	table := uitable.New()
	table.Wrap = true
	table.AddRow("Ref:", ref.Name)
	table.AddRow("Digest:", ref.Content[0].Digest.Hex())
	table.AddRow("Size:", byteCountBinary(ref.Content[0].Size))
	table.AddRow()
	_, err = writer.Write(table.Bytes())
	return err
}
