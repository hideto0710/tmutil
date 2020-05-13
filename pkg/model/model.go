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
package model

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var DefaultHandlers = []string{
	"text_classifier",
	"image_classifier",
	"object_detector",
	"image_segmenter",
}

type Model struct {
	ModelName      string   `json:"modelName,omitempty",validate:"required"`
	Version        string   `json:"version,omitempty"`
	ModelFile      string   `json:"modelFile,omitempty",validate:"required"`
	SerializedFile string   `json:"serializedFile,omitempty"validate:"required"`
	ExtraFiles     []string `json:"extraFiles,omitempty"`
	Handler        string   `json:"handler,omitempty"`
	SourceVocab    string   `json:"sourceVocab,omitempty"`
	Runtime        string   `json:"runtime,omitempty"`
}

type Ref struct {
	Name      string
	Exists    bool
	Manifest  ocispec.Descriptor
	Config    ocispec.Descriptor
	Content   []ocispec.Descriptor
	Size      int64
	Digest    digest.Digest
	CreatedAt time.Time
}

func (m *Model) Manifest() *Manifest {
	mm := &ManifestModel{
		ModelName:        m.ModelName,
		ModelNameVersion: m.Version,
		ModelFile:        filepath.Base(m.ModelFile),
		SerializedFile:   filepath.Base(m.SerializedFile),
		Handler:          m.Handler,
	}
	if strings.HasSuffix(mm.Handler, ".py") {
		mm.Handler = filepath.Base(m.Handler)
	}
	if m.SourceVocab != "" {
		mm.SourceVocab = filepath.Base(m.SourceVocab)
	}
	return &Manifest{
		Runtime:               m.Runtime,
		Model:                 mm,
		ModelServerVersion:    "1.0",
		ImplementationVersion: "1.0",
		SpecificationVersion:  "1.0",
	}
}

func (m *Model) IsDefaultHandler() bool {
	if strings.HasSuffix(m.Handler, ".py") {
		return false
	}
	isDefault := false
	for _, h := range DefaultHandlers {
		if m.Handler == h {
			isDefault = true
		}
	}
	return isDefault
}

func (m *Model) IsCustomHandler() bool {
	return strings.HasSuffix(m.Handler, ".py")
}
