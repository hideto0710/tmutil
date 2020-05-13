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
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/ghodss/yaml"
	"github.com/go-playground/validator/v10"
	"github.com/hideto0710/torchstand/pkg/action"
	"github.com/hideto0710/torchstand/pkg/model"
	"github.com/spf13/cobra"
)

func newCmdArchive(cfg *action.Configuration) *cobra.Command {
	var file string
	opts := &action.ArchiveOpts{}

	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Archive a model",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			var content []byte
			var err error
			if file == "-" {
				buf := new(bytes.Buffer)
				_, err = buf.ReadFrom(cmd.InOrStdin())
				content = buf.Bytes()
			} else {
				content, err = ioutil.ReadFile(file)
			}
			if err != nil {
				return err
			}
			m := &model.Model{}
			err = yaml.Unmarshal(content, &m)
			if err != nil {
				return err
			}
			validate.RegisterStructValidation(modelNameValidation, model.Model{})
			if err := validate.Struct(m); err != nil {
				// TODO: stdout validation error description.
				return err
			}
			if m.Runtime == "" {
				m.Runtime = "python"
			}
			if opts.Tag == "" {
				opts.Tag = fmt.Sprintf("%s:latest", m.ModelName)
			}
			return action.NewArchive(cfg).Run(m, opts, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVarP(&opts.Tag, "tag", "t", "", "Name and optionally a tag in the 'name:tag' format")
	cmd.Flags().StringVarP(&file, "file", "f", "torchserve.yaml", "Name of TorchServe config file (Default is ‘PATH/torchserve.yaml’)")

	return cmd
}

func modelNameValidation(sl validator.StructLevel) {
	m := sl.Current().Interface().(model.Model)
	matched, err := regexp.Match(`^[A-Za-z0-9][A-Za-z0-9_\-.]*$`, []byte(m.ModelName))
	if err != nil || !matched {
		sl.ReportError(m.ModelName, "modelName", "ModelName", "", "model name must be ^[A-Za-z0-9][A-Za-z0-9_\\-.]*$")
	}
	if m.Handler == "text_classifier" && m.SourceVocab == "" {
		sl.ReportError(m.SourceVocab, "sourceVocab", "SourceVocab", "", "provide the source language vocab")
	}
	if m.IsCustomHandler() && !m.IsDefaultHandler() {
		sl.ReportError(m.Handler, "handler", "Handler", "", "should be one of the default TorchServe handlers")
	}
}
