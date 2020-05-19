package action

import (
	"path/filepath"
	"strings"

	"github.com/hideto0710/torchstand/pkg/types"
)

var DefaultHandlers = []string{
	"text_classifier",
	"image_classifier",
	"object_detector",
	"image_segmenter",
}

type TorchServeModelfile struct {
	ModelName      string   `json:"modelName,omitempty",validate:"required"`
	Version        string   `json:"version,omitempty"`
	ModelFile      string   `json:"modelFile,omitempty",validate:"required"`
	SerializedFile string   `json:"serializedFile,omitempty"validate:"required"`
	ExtraFiles     []string `json:"extraFiles,omitempty"`
	Handler        string   `json:"handler,omitempty"`
	SourceVocab    string   `json:"sourceVocab,omitempty"`
	Runtime        string   `json:"runtime,omitempty"`
}

func (m *TorchServeModelfile) Manifest() *types.Manifest {
	mm := &types.Model{
		ModelName:      m.ModelName,
		ModelVersion:   m.Version,
		ModelFile:      filepath.Base(m.ModelFile),
		SerializedFile: filepath.Base(m.SerializedFile),
		Handler:        m.Handler,
	}
	if strings.HasSuffix(mm.Handler, ".py") {
		mm.Handler = filepath.Base(m.Handler)
	}
	if m.SourceVocab != "" {
		mm.SourceVocab = filepath.Base(m.SourceVocab)
	}
	return &types.Manifest{
		Runtime:               m.Runtime,
		Model:                 mm,
		ModelServerVersion:    "1.0",
		ImplementationVersion: "1.0",
		SpecificationVersion:  "1.0",
	}
}

func (m *TorchServeModelfile) IsDefaultHandler() bool {
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

func (m *TorchServeModelfile) IsCustomHandler() bool {
	return strings.HasSuffix(m.Handler, ".py")
}
