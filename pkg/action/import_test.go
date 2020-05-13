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
	"os"
	"testing"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/hideto0710/torchstand/pkg/model"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
}

func TestImport_Run(t *testing.T) {
	cfg := new(Configuration)
	store, _ := content.NewOCIStore("./cache")
	cfg.OCIStore = store
	cfg.Resolver = docker.NewResolver(docker.ResolverOptions{})
	instance := NewImport(cfg)

	t.Run("ok", func(t *testing.T) {
		body := []byte("Hello World!\n")
		ref := "localhost:5000/hello-artifact:v1"
		m := &model.Model{ModelName: "hello"}
		if err := instance.Run(ref, body, m, os.Stdout); err != nil {
			t.Error(err)
		}
	})
}
