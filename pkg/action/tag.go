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
)

type Tag struct {
	cfg *Configuration
}

func NewTag(cfg *Configuration) *Tag {
	return &Tag{
		cfg: cfg,
	}
}

func (p *Tag) Run(srcRef string, destRef string) error {
	ctx := orascontext.Background()
	store := p.cfg.OCIStore

	if err := store.LoadIndex(); err != nil {
		return err
	}
	ref, err := p.cfg.FetchReference(ctx, srcRef)
	if err != nil {
		return err
	}
	if !ref.Exists {
		return fmt.Errorf("model not found: %s", ref.Name)
	}
	newManifest := ref.Manifest
	// MARK: `AddReference` updates `Descriptor.Annotations`
	annotations := make(map[string]string)
	for key, value := range ref.Manifest.Annotations {
		annotations[key] = value
	}
	newManifest.Annotations = annotations
	store.AddReference(destRef, newManifest)
	return store.SaveIndex()
}
