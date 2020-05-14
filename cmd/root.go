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
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	authdocker "github.com/deislabs/oras/pkg/auth/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/go-playground/validator/v10"
	"github.com/hideto0710/torchstand/pkg/action"
	"github.com/hideto0710/torchstand/pkg/path"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	cfgFile   string
	verbose   bool
	insecure  bool
	plainHTTP bool

	logger         *zap.Logger
	validate       *validator.Validate
	torchstandPath *path.TorchstandPath
)

func newCmdRoot(cfg *action.Configuration) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "torchstand",
		Short:   "Utilities for TorchServe",
		Long:    ``,
		Version: "0.0.1",
	}
	validate = validator.New()
	cobra.OnInitialize(initConfig)
	cmd.Flags().Bool("version", false, "Show the TorchStand version information")
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.torchstand.yaml)")
	cmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "verbose output")
	cmd.PersistentFlags().BoolVar(&insecure, "insecure", false, "allow connections to SSL registry without certs")
	cmd.PersistentFlags().BoolVar(&plainHTTP, "plain-http", false, "use plain http and not https")

	cmd.AddCommand(
		newCmdPush(cfg),
		newCmdPull(cfg),
		newCmdModels(cfg),
		newCmdImport(cfg),
		newCmdRun(cfg),
		newCmdRemoveModel(cfg),
		newCmdArchive(cfg),
		newCmdTag(cfg),
	)
	return cmd
}

func Execute() {
	home, err := homedir.Dir()
	checkError(err)
	torchstandPath = path.NewTorchstandPath(home)

	actionConfig := new(action.Configuration)
	store, err := content.NewOCIStore(torchstandPath.CachePath())
	checkError(err)
	actionConfig.OCIStore = store

	cli, err := authdocker.NewClient()
	checkError(err)
	client := http.DefaultClient
	if insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	resolver, err := cli.Resolver(context.Background(), client, plainHTTP)
	checkError(err)
	actionConfig.Resolver = resolver

	actionConfig.Path = torchstandPath

	cmd := newCmdRoot(actionConfig)
	err = cmd.Execute()
	checkError(err)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		checkError(err)
		viper.AddConfigPath(home)
		viper.SetConfigName(".torchstand")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zapcore.WarnLevel)
	if verbose {
		config.Level.SetLevel(zapcore.DebugLevel)
	}
	var err error
	logger, err = config.Build()
	checkError(err)
	zap.ReplaceGlobals(logger)
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
