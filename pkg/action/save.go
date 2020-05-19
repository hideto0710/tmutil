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
	"github.com/hideto0710/torchstand/pkg/util"
)

type Save struct {
	cfg *Configuration
}

func NewSave(cfg *Configuration) *Save {
	return &Save{cfg: cfg}
}

func (s *Save) Run(argRef string, writer io.Writer) error {
	ctx := orascontext.Background()
	store := s.cfg.OCIStore

	if err := store.LoadIndex(); err != nil {
		return err
	}
	ref, err := s.cfg.FetchReference(ctx, argRef)
	if err != nil {
		return err
	}
	if !ref.Exists {
		_, err := fmt.Fprintf(writer, "Ref: %s not found\n", argRef)
		return err
	}
	return util.NewArchiver(ref, s.cfg.Path.RegistryPath()).Archive(writer)
}
