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
	"github.com/hideto0710/torchstand/pkg/action"
	"github.com/spf13/cobra"
)

func newCmdTag(cfg *action.Configuration) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag source[:tag] target[:tag]",
		Short: "Create a tag that refers to source model",
		Long:  ``,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.NewTag(cfg).Run(args[0], args[1])
		},
	}
	return cmd
}
