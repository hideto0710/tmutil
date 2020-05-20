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
	"archive/zip"
	"bytes"
	"sort"
	"strings"
	"testing"
	"time"

	torchstandTypes "github.com/hideto0710/torchstand/pkg/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestArchiver_Archive(t *testing.T) {
	name := "localhost:5000/densenet161:v1"
	manifestDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageConfig,
		Digest:    "sha256:70bfdf5e",
		Size:      359,
		Annotations: map[string]string{
			"org.opencontainers.image.title": name,
		},
	}
	configDesc := ocispec.Descriptor{
		MediaType: torchstandTypes.TorchServeModelConfigMediaType,
		Digest:    "sha256:64250a66",
		Size:      259,
	}
	pyTorchModelDesc := ocispec.Descriptor{
		MediaType: torchstandTypes.PyTorchModelMediaType,
		Digest:    "sha256:8d451a50",
		Size:      115730790,
		Annotations: map[string]string{
			"org.opencontainers.image.title": "densenet161.pth",
		},
	}
	contentDesc := ocispec.Descriptor{
		MediaType: torchstandTypes.TorchServeModelContentLayerMediaType,
		Digest:    "sha256:736fda8d",
		Size:      37923,
		Annotations: map[string]string{
			"org.opencontainers.image.title": "densenet161",
		},
	}

	t.Run("ok", func(t *testing.T) {
		ref := &torchstandTypes.Ref{
			Name:         "localhost:5000/densenet161:v1",
			Exists:       true,
			Manifest:     manifestDesc,
			Config:       configDesc,
			PyTorchModel: pyTorchModelDesc,
			Content:      contentDesc,
			Size:         115768972,
			Digest:       "sha256:70bfdf5e",
			CreatedAt:    time.Date(2020, 5, 20, 10, 0, 0, 0, time.UTC),
		}

		writer := &bytes.Buffer{}
		if err := NewArchiver(ref, "../testdata/registry").Archive(writer); err != nil {
			t.Fatal(err)
		}
		reader, err := zip.NewReader(bytes.NewReader(writer.Bytes()), int64(len(writer.Bytes())))
		if err != nil {
			t.Fatal(err)
		}
		var fs []string
		for _, f := range reader.File {
			fs = append(fs, f.Name)
		}
		sort.Strings(fs)

		want := strings.Join([]string{
			"MAR-INF/MANIFEST.json",
			"densenet161.pth",
			"index_to_name.json",
			"model.py",
		}, ",")
		got := strings.Join(fs, ",")
		if got != want {
			t.Errorf("want: %s, got: %s", want, got)
		}
	})

	t.Run("notfound", func(t *testing.T) {
		configDesc.Digest = "sha256:invalid"
		ref := &torchstandTypes.Ref{
			Name:         "localhost:5000/densenet161:v1",
			Exists:       true,
			Manifest:     manifestDesc,
			Config:       configDesc,
			PyTorchModel: pyTorchModelDesc,
			Content:      contentDesc,
			Size:         115768972,
			Digest:       "sha256:70bfdf5e",
			CreatedAt:    time.Date(2020, 5, 20, 10, 0, 0, 0, time.UTC),
		}

		writer := &bytes.Buffer{}
		err := NewArchiver(ref, "../testdata/registry").Archive(writer)
		want := "open ../testdata/registry/blobs/sha256/invalid: no such file or directory"
		if err == nil {
			t.Fatalf("want: %s, got: no error", want)
		}
		got := err.Error()
		if got != want {
			t.Errorf("want: %s, got: %s", want, got)
		}
	})
}
