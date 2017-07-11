package deploy

import (
	"testing"
)

func TestDivideImageAndTag(t *testing.T) {
	imageWithTag := "nginx:latest"
	image, tag, err := divideImageAndTag(imageWithTag)
	if err != nil {
		t.Error(err)
	}
	if *image != "nginx" {
		t.Errorf("image is invalid: %s", *image)
	}
	if *tag != "latest" {
		t.Errorf("tag is invalid: %s", *tag)
	}
}
