package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/1azunna/zapgo/internal/defaults"
	"github.com/1azunna/zapgo/internal/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/sirupsen/logrus"
)

// Struct representing events returned from image pulling
type pullEvent struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
	Progress       string `json:"progress,omitempty"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

type imageClient interface {
	ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error)
	ImagePull(ctx context.Context, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)
}

func (c Docker) ImageExists(dockerClient imageClient, image string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), defaults.TimeoutInS)
	defer cancel()

	args := filters.NewArgs(filters.Arg("reference", image))
	logrus.Debugf("Checking if %s already exists", image)
	imgs, err := dockerClient.ImageList(ctx, types.ImageListOptions{
		Filters: args,
	})
	if err != nil {
		logrus.Fatalf("Failed to list image with filters %v due to %v", args, err)
	}

	for _, img := range imgs {
		for _, repotag := range img.RepoTags {
			if strings.HasPrefix(repotag, image) {
				logrus.Debug("Image exists !!")
				return true
			}
		}
	}
	return false
}

func (c Docker) PullImage(dockerClient imageClient, image string) bool {
	ctx := context.Background()
	// pull the zap image from DockerHub
	logrus.Info(fmt.Sprintf("Pulling the latest image verson for %s", image))
	resp, err := dockerClient.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		logrus.Fatalf("Failed to pull image %s due to %v", image, err)
	}

	defer resp.Close()

	cursor := utils.Cursor{}
	layers := make([]string, 0)
	oldIndex := len(layers)

	var event *pullEvent
	decoder := json.NewDecoder(resp)

	fmt.Printf("\n")
	cursor.Hide()

	for {
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}

			panic(err)
		}

		imageID := event.ID

		// Check if the line is one of the final two ones
		if strings.HasPrefix(event.Status, "Digest:") || strings.HasPrefix(event.Status, "Status:") {
			logrus.Info(event.Status)
			continue
		}

		// Check if ID has already passed once
		index := 0
		for i, v := range layers {
			if v == imageID {
				index = i + 1
				break
			}
		}

		// Move the cursor
		if index > 0 {
			diff := index - oldIndex

			if diff > 1 {
				down := diff - 1
				cursor.MoveDown(down)
			} else if diff < 1 {
				up := diff*(-1) + 1
				cursor.MoveUp(up)
			}

			oldIndex = index
		} else {
			layers = append(layers, event.ID)
			diff := len(layers) - oldIndex

			if diff > 1 {
				cursor.MoveDown(diff) // Return to the last row
			}

			oldIndex = len(layers)
		}

		cursor.ClearLine()

		if event.Status == "Pull complete" {
			logrus.Infof("%s: %s", event.ID, event.Status)
		} else {
			logrus.Infof("%s: %s %s", event.ID, event.Status, event.Progress)
		}

	}

	cursor.Show()

	if strings.Contains(event.Status, fmt.Sprintf("Downloaded newer image for %s", image)) || strings.Contains(event.Status, fmt.Sprintf("Image is up to date for %s", image)) {
		return true
	}
	return false
}
