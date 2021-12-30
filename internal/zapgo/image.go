package zapgo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	u "github.com/1azunna/zapgo/utils"
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

func ZapImageExists(image string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutInS)
	defer cancel()

	dockerClient := NewClient()
	args := filters.NewArgs(filters.Arg("reference", image))
	imgs, err := dockerClient.ImageList(ctx, types.ImageListOptions{
		Filters: args,
	})
	if err != nil {
		logrus.Fatalf("Failed to list image with filters %v due to %v", args, err)
	}

	for _, img := range imgs {
		for _, repotag := range img.RepoTags {
			if strings.HasPrefix(repotag, image) {
				return true
			}
		}
	}
	return false
}

func PullZapImage(image string, logger Logger) bool {
	// create a new docker client
	dockerClient := NewClient()
	ctx := context.Background()
	// pull the zap image from DockerHub
	logger.Info(fmt.Sprintf("Pulling the latest image verson for %s", image))
	resp, err := dockerClient.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		logrus.Fatalf("Failed to pull image %s due to %v", image, err)
	}

	defer resp.Close()

	cursor := u.Cursor{}
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
			logger.Info(event.Status)
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
			logger.Info(fmt.Sprintf("%s: %s", event.ID, event.Status))
		} else {
			logger.Info(fmt.Sprintf("%s: %s %s", event.ID, event.Status, event.Progress))
		}

	}

	cursor.Show()

	if strings.Contains(event.Status, fmt.Sprintf("Downloaded newer image for %s", image)) || strings.Contains(event.Status, fmt.Sprintf("Image is up to date for %s", image)) {
		return true
	}
	return false
}
