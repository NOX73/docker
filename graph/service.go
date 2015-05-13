package graph

import (
	"fmt"
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/utils"
)

func (s *TagStore) LookupRaw(name string) ([]byte, error) {
	image, err := s.LookupImage(name)
	if err != nil || image == nil {
		return nil, fmt.Errorf("No such image %s", name)
	}

	imageInspectRaw, err := image.RawJson()
	if err != nil {
		return nil, err
	}

	return imageInspectRaw, nil
}

// Lookup return an image encoded in JSON
func (s *TagStore) Lookup(name string) (*types.ImageInspect, error) {
	image, err := s.LookupImage(name)
	if err != nil || image == nil {
		return nil, fmt.Errorf("No such image: %s", name)
	}

	var tags = make([]string, 0)

	s.Lock()
	for repoName, repository := range s.Repositories {
		for ref, id := range repository {
			if id == image.ID {
				imgRef := utils.ImageReference(repoName, ref)
				tags = append(tags, imgRef)
			}
		}
	}
	s.Unlock()

	imageInspect := &types.ImageInspect{
		Id:              image.ID,
		Tags:            tags,
		Parent:          image.Parent,
		Comment:         image.Comment,
		Created:         image.Created,
		Container:       image.Container,
		ContainerConfig: &image.ContainerConfig,
		DockerVersion:   image.DockerVersion,
		Author:          image.Author,
		Config:          image.Config,
		Architecture:    image.Architecture,
		Os:              image.OS,
		Size:            image.Size,
		VirtualSize:     image.GetParentsSize(0) + image.Size,
	}

	return imageInspect, nil
}

// ImageTarLayer return the tarLayer of the image
func (s *TagStore) ImageTarLayer(name string, dest io.Writer) error {
	if image, err := s.LookupImage(name); err == nil && image != nil {
		fs, err := image.TarLayer()
		if err != nil {
			return err
		}
		defer fs.Close()

		written, err := io.Copy(dest, fs)
		if err != nil {
			return err
		}
		logrus.Debugf("rendered layer for %s of [%d] size", image.ID, written)
		return nil
	}
	return fmt.Errorf("No such image: %s", name)
}
