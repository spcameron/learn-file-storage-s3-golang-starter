package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os/exec"
)

func getVideoAspectRatio(filepath string) (string, error) {
	type streams struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}

	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath)

	var b bytes.Buffer
	cmd.Stdout = &b

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cmd.Run() failed: %w", err)
	}

	var dimensions streams
	if err := json.Unmarshal(b.Bytes(), &dimensions); err != nil {
		return "", errors.New("json unmarshal error")
	}

	return determineAspectRatio(dimensions.Streams[0].Width, dimensions.Streams[0].Height)
}

func determineAspectRatio(width, height int) (string, error) {
	if width <= 0 || height <= 0 {
		return "", fmt.Errorf("invalid dimensions: width=%d height=%d", width, height)
	}

	r := float64(width) / float64(height)
	epsilon := 0.02

	switch {
	case math.Abs(r-(16.0/9.0)) < epsilon:
		return "16:9", nil
	case math.Abs(r-(9.0/16.0)) < epsilon:
		return "9:16", nil
	default:
		return "other", nil
	}
}
