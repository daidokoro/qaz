package troposphere

import (
	"archive/tar"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"golang.org/x/net/context"
)

// dockerfile string used to create image
const dockerfile = `
FROM python:2.7
RUN pip install troposphere`

// buildlog - used to unmarshal build
// message from Dockerfile build job
type buildlog struct {
	Stream string `json:"stream"`
}

func (b *buildlog) Line(l []byte) string {
	if err := json.Unmarshal(l, b); err != nil {
		log.Error("error unmarshalling build log: %v", err)
		return ""
	}

	return strings.Trim(b.Stream, "\n")
}

// BuildImage - build troposphere docker image
// for executing troposphere code
func BuildImage() (err error) {

	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()
	th := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len([]byte(dockerfile))),
	}

	if err = tw.WriteHeader(th); err != nil {
		return fmt.Errorf("error creating tar for dockerfile: %v", err)
	}

	if _, err := tw.Write([]byte(dockerfile)); err != nil {
		return fmt.Errorf("error writing dockerfile to tar: %v", err)
	}

	// dockerfile tar reader
	dftr := bytes.NewReader(buf.Bytes())

	log.Debug("troposphere stack detected, creating docker image")

	resp, err := cli.ImageBuild(ctx, dftr, types.ImageBuildOptions{
		Context: dftr,
		// Dockerfile: "/Users/daidokoro/Go/src/github.com/daidokoro/qaz/docker/Dockerfile",
		Tags: []string{"troposphere:qaz"},
	})
	if err != nil {
		return fmt.Errorf("failed to create image: %v", err)
	}

	defer resp.Body.Close()
	// io.Copy(os.Stdout, resp.Body)

	scanner := bufio.NewScanner(resp.Body)
	var bl buildlog
	for scanner.Scan() {
		line := bl.Line(scanner.Bytes())

		if regexp.MustCompile(`(?i)from|run|success|cache`).MatchString(line) {
			log.Debug(line)
		}

	}

	return
}
