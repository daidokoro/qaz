// Package troposphere contains methods for Executing troposphere
// python files using docker.
// note that Docker is required when using this package.
package troposphere

import (
	"fmt"
	"io/ioutil"

	"github.com/daidokoro/qaz/logger"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

// python container
var (
	troposphereImage = "troposphere:qaz"
)

// default python container
const (
	defaultimage = "troposphere:qaz"
)

// Image - sets troposphere image to be
// used for execution
func Image(img string) {
	defer func() { log.Debug("troposphere container set to: [%s]", troposphereImage) }()
	if img == "" {
		troposphereImage = defaultimage
		return
	}
	troposphereImage = img
}

var log *logger.Logger

// Logger - set logger for package
func Logger(l *logger.Logger) {
	log = l
}

// Execute - run troposphere code
func Execute(code string) (t string, err error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		return t, fmt.Errorf("error creating docker client: %v", err)
	}

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image: troposphereImage,
			Cmd:   []string{"python", "-c", code},
			Tty:   true,
		}, nil, nil, "")

	if err != nil {
		return t, fmt.Errorf("error while creating troposphere container: %v", err)
	}

	// start container
	if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return t, fmt.Errorf("error starting troposphere container %v", err)
	}

	log.Debug("troposphere containter created [%s]", resp.ID)

	// remove container when done
	defer func() {
		if err := cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
			log.Error("error removing troposphere container [%s]: %v", resp.ID, err)
		}
	}()

	// wait till complete
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err = <-errCh:
		if err != nil {
			return t, fmt.Errorf("error while waiting for troposphere container: %v", err)
		}
	case <-statusCh:
		log.Debug("container wait completed")
	}

	// get ouput
	r, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return t, fmt.Errorf("error fetching tropophere container output: %v", err)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return t, fmt.Errorf("error reading container output: %v", err)
	}

	t = string(b)
	return
}
