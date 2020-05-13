package model

type ManifestModel struct {
	ModelName        string `json:"modelName,omitempty",validate:"required"`
	ModelNameVersion string `json:"modelVersion,omitempty"`
	ModelFile        string `json:"modelFile,omitempty",validate:"required"`
	SerializedFile   string `json:"serializedFile,omitempty"validate:"required"`
	Handler          string `json:"handler,omitempty"`
	SourceVocab      string `json:"sourceVocab,omitempty"`
}

type Manifest struct {
	Runtime               string         `json:"runtime"`
	Model                 *ManifestModel `json:"model"`
	ModelServerVersion    string         `json:"modelServerVersion"`
	ImplementationVersion string         `json:"implementationVersion"`
	SpecificationVersion  string         `json:"specificationVersion"`
}
