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

package types

// https://github.com/pytorch/serve/blob/v0.1.0/model-archiver/model_archiver/manifest_components/model.py
type Model struct {
	ModelName      string `json:"modelName",validate:"required"`
	SerializedFile string `json:"serializedFile"validate:"required"`
	Handler        string `json:"handler"validate:"required"`
	SourceVocab    string `json:"sourceVocab,omitempty"`
	ModelFile      string `json:"modelFile",validate:"required"`
	Description    string `json:"description,omitempty"`
	ModelVersion   string `json:"modelVersion,omitempty"`
	Extensions     string `json:"extensions,omitempty"`
}

// https://github.com/pytorch/serve/blob/v0.1.0/model-archiver/model_archiver/manifest_components/engine.py
type Engine struct {
	EngineName    string `json:"engineName"`
	EngineVersion string `json:"engineVersion,omitempty"`
}

// https://github.com/pytorch/serve/blob/v0.1.0/model-archiver/model_archiver/manifest_components/publisher.py
type Publisher struct {
	Author string `json:"author"`
	Email  string `json:"email"`
}

// https://github.com/pytorch/serve/blob/v0.1.0/model-archiver/model_archiver/manifest_components/manifest.py
type Manifest struct {
	Runtime               string     `json:"runtime"`
	Model                 *Model     `json:"model"`
	Engine                *Engine    `json:"engine,omitempty"`
	License               string     `json:"license,omitempty"`
	ModelServerVersion    string     `json:"modelServerVersion,omitempty"`
	Description           string     `json:"description,omitempty"`
	ImplementationVersion string     `json:"implementationVersion,omitempty"`
	SpecificationVersion  string     `json:"specificationVersion,omitempty"`
	UserData              string     `json:"userData,omitempty"`
	Publisher             *Publisher `json:"publisher,omitempty"`
}
