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
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"

	orascontext "github.com/deislabs/oras/pkg/context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/go-connections/nat"
)

const (
	torchserveImageName = "pytorch/torchserve:0.1-cpu"
	torchservePort      = "8080"
)

type RunOpts struct {
	Port    string
	Network string
}

type Run struct {
	cfg *Configuration
}

func NewRun(cfg *Configuration) *Run {
	return &Run{
		cfg: cfg,
	}
}

func (p *Run) Run(ref string, opts *RunOpts, writer io.Writer) error {
	ctx := orascontext.Background()
	store := p.cfg.OCIStore

	if err := store.LoadIndex(); err != nil {
		return err
	}
	r, err := p.cfg.FetchReference(ctx, ref)
	if err != nil {
		return err
	}
	fullPath, err := filepath.Abs(filepath.Join(p.cfg.Path.CachePath(), "blobs", r.Digest.Algorithm().String(), r.Digest.Hex()))
	if err != nil {
		return err
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	pullOut, err := cli.ImagePull(ctx, torchserveImageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	termFd, isTerm := term.GetFdInfo(writer)
	_ = jsonmessage.DisplayJSONMessagesStream(pullOut, writer, termFd, isTerm, nil)
	_ = pullOut.Close()

	hostBinding := nat.PortBinding{HostIP: "0.0.0.0", HostPort: opts.Port}
	containerPort, err := nat.NewPort("tcp", torchservePort)
	if err != nil {
		return err
	}
	containerName := shortDigest(r.Digest.Hex())
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:      torchserveImageName,
		Entrypoint: []string{"torchserve"},
		Cmd: []string{
			"--start",
			"--ts-config",
			"/home/model-server/config.properties",
			"--model-store",
			"/home/model-server/model-store",
			"--models",
			r.Digest.Hex(),
			"--foreground",
		},
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/home/model-server/model-store", filepath.Dir(fullPath)),
		},
		PortBindings: nat.PortMap{
			containerPort: []nat.PortBinding{hostBinding},
		},
		AutoRemove: true,
	}, nil, containerName)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(writer, "starting container %s\n", containerName)
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	if opts.Network != "" {
		if err := cli.NetworkConnect(ctx, opts.Network, resp.ID, &network.EndpointSettings{}); err != nil {
			return err
		}
	}
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	containerOut, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return err
	}
	defer containerOut.Close()
	go stdcopy.StdCopy(writer, writer, containerOut)

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)

	select {
	case <-statusCh:
		_, _ = fmt.Fprintln(writer, "receive signal")
	case err := <-errCh:
		_, _ = fmt.Fprintln(writer, "receive error")
		if err != nil {
			return err
		}
	case <-quit:
		_, _ = fmt.Fprintln(writer, "stopping container ...")
		if err := cli.ContainerStop(ctx, resp.ID, nil); err != nil {
			return err
		}
	}
	return nil
}
