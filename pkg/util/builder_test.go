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

package util

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/hideto0710/torchstand/pkg/types"
)

func TestBuilder_Build(t *testing.T) {
	modelName := "testmodel"
	version := "0.1"
	runtime := "python"
	handler := "image_classifier"
	modelFile := "../testdata/build/model.py"
	serializedFile := "../testdata/build/densenet161.pth"
	extraFile := "../testdata/build/index_to_name.json"

	manifest := &types.TorchServeModelfile{
		ModelName:      modelName,
		Version:        version,
		ModelFile:      modelFile,
		SerializedFile: serializedFile,
		ExtraFiles:     []string{extraFile},
		Handler:        handler,
		SourceVocab:    "",
		Runtime:        runtime,
	}

	bs, err := NewBuilder(manifest).Build()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("manifest", func(t *testing.T) {
		m := &types.Manifest{}
		if err := json.Unmarshal(bs.Config, &m); err != nil {
			t.Fatal(err)
		}
		table := []struct {
			got  string
			want string
		}{
			{m.Runtime, runtime},
			{m.Model.ModelName, modelName},
			{m.Model.ModelVersion, version},
			{m.Model.SerializedFile, filepath.Base(serializedFile)},
			{m.Model.ModelFile, filepath.Base(modelFile)},
			{m.Model.Handler, handler},
		}
		for _, e := range table {
			if e.got != e.want {
				t.Errorf("want: %s, got: %s", e.want, e.got)
			}
		}
	})

	t.Run("model", func(t *testing.T) {
		got := string(bs.PyTorchModel)
		want := "hello"
		if got != want {
			t.Errorf("want: %s, got: %s", want, got)
		}
	})

	t.Run("contents", func(t *testing.T) {
		reader := tar.NewReader(bytes.NewReader(bs.Contents))
		var fs []string
		for {
			hdr, err := reader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Error(err)
			}
			fs = append(fs, hdr.Name)
		}
		sort.Strings(fs)

		got := strings.Join(fs, ",")
		want := strings.Join([]string{
			"index_to_name.json",
			"model.py",
		}, ",")

		if got != want {
			t.Errorf("want: %s, got: %s", want, got)
		}
	})
}
