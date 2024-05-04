package k8s

import (
	"context"
	"github.com/distribution/reference"
	"github.com/secureworks/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetContainerImage(ctx context.Context, c client.Client, podNamespace, podName, containerName string) (string, error) {
	podObjectKey := client.ObjectKey{Namespace: podNamespace, Name: podName}

	pod := &corev1.Pod{}
	if err := c.Get(ctx, podObjectKey, pod); err != nil {
		return "", errors.New("failed fetching pod '%s': %w", podObjectKey, err)
	}

	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			return container.Image, nil
		}
	}

	return "", errors.New("container '%s' not found in pod '%s'", containerName, podObjectKey)
}

func GetImageTag(image string) (string, error) {
	if r, err := reference.Parse(image); err != nil {
		return "", errors.New("failed parsing image '%s': %w", image, err)
	} else if tagged, ok := r.(reference.Tagged); ok {
		return tagged.Tag(), nil
	} else {
		return "", errors.New("could not parse tag from image '%s'", image)
	}
}
