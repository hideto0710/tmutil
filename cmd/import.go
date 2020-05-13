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

package cmd

import (
	"bytes"
	"io/ioutil"

	"github.com/hideto0710/torchstand/pkg/action"
	"github.com/hideto0710/torchstand/pkg/model"
	"github.com/spf13/cobra"
)

func newCmdImport(cfg *action.Configuration) *cobra.Command {
	m := &model.Model{}

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import a model from mar file",
		Long:  ``,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]
			path := args[1]
			var content []byte
			var err error
			if path == "-" {
				buf := new(bytes.Buffer)
				_, err = buf.ReadFrom(cmd.InOrStdin())
				content = buf.Bytes()
			} else {
				content, err = ioutil.ReadFile(path)
			}
			if err != nil {
				return err
			}
			// TODO: read manifest file inside zip
			return action.NewImport(cfg).Run(ref, content, m, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&m.ModelName, "model-name", "", "model name")
	cmd.MarkFlagRequired("model-name")

	return cmd
}
