package runner

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	faryukTypes "FaRyuk/internal/types"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/uuid"
)

type RunnerHandler struct {
	ctx context.Context
	cli *client.Client
}

func NewRunnerHandler() *RunnerHandler {
	ctx := context.Background()
	// cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	return &RunnerHandler{ctx, cli}
}

func NewRunner(tag string, displayName string, cmd []string, owner string, isWeb bool, isPort bool) *faryukTypes.Runner {
	id := uuid.New().String()

	return &faryukTypes.Runner{
		ID:          id,
		Tag:         tag,
		DisplayName: displayName,
		Cmd:         cmd,
		IsWeb:       isWeb,
		IsPort:      isPort,
		Owner:       owner,
	}
}

func (rHandler *RunnerHandler) PullImage(tag string) (string, error) {
	fullname := tag
	if ss := strings.Split(tag, "/"); len(ss) == 1 {
		fullname = fmt.Sprintf("docker.io/library/%s", tag)
	} else if len(ss) == 2 {
		fullname = fmt.Sprintf("docker.io/%s", tag)
	}
	_, err := rHandler.cli.ImagePull(rHandler.ctx, fullname, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}
	return tag, nil
}

func (rHandler *RunnerHandler) RunCmd(imgId string, cmd []string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	resp, err := rHandler.cli.ContainerCreate(rHandler.ctx, &container.Config{
		Image: imgId,
		Cmd:   cmd,
		Tty:   false,
	}, nil, nil, nil, "")
	if err != nil {
		return "", "", err
	}

	defer func() {
		err := rHandler.cli.ContainerRemove(rHandler.ctx, resp.ID, types.ContainerRemoveOptions{})
		if err != nil {
			return
		}
	}()

	if err = rHandler.cli.ContainerStart(rHandler.ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", "", err
	}

	statusCh, errCh := rHandler.cli.ContainerWait(rHandler.ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err = <-errCh:
		if err != nil {
			return "", "", err
		}
	case <-statusCh:
	}

	out, err := rHandler.cli.ContainerLogs(rHandler.ctx, resp.ID,
		types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		return "", "", err
	}
	defer out.Close()

	_, err = stdcopy.StdCopy(&stdout, &stderr, out)
	if err != nil {
		return "", "", err
	}

	return stdout.String(), stderr.String(), nil
}
