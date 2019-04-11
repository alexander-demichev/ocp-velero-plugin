package common

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	corev1API "k8s.io/api/core/v1"
)

// ReplaceImageRefPrefix replaces an image reference prefix with newPrefix.
// If the input image reference does not start with oldPrefix, an error is returned
func ReplaceImageRefPrefix(s, oldPrefix, newPrefix string) (string, error) {
	refSplit := strings.SplitN(s, "/", 2)
	if len(refSplit) != 2 {
		err := fmt.Errorf("image reference [%v] does not have prefix [%v]", s, oldPrefix)
		return "", err
	}
	if refSplit[0] != oldPrefix {
		err := fmt.Errorf("image reference [%v] does not have prefix [%v]", s, oldPrefix)
		return "", err
	}
	return fmt.Sprintf("%s/%s", newPrefix, refSplit[1]), nil
}

// HasImageRefPrefix returns true if the input image reference begins with
// the input prefix followed by "/"
func HasImageRefPrefix(s, prefix string) bool {
	refSplit := strings.SplitN(s, "/", 2)
	if len(refSplit) != 2 {
		return false
	}
	return refSplit[0] == prefix
}

// LocalImageReference describes an image in the internal openshift registry
type LocalImageReference struct {
	Registry  string
	Namespace string
	Name      string
	Tag       string
	Digest    string
}

// ParseLocalImageReference
func ParseLocalImageReference(s, prefix string) (*LocalImageReference, error) {
	refSplit := strings.Split(s, "/")
	if refSplit[0] != prefix {
		return nil, fmt.Errorf("image reference is not local")
	}
	if len(refSplit) != 3 {
		return nil, fmt.Errorf("Unexpected image reference %s", s)
	}
	parsed := LocalImageReference{Registry: prefix, Namespace: refSplit[1]}
	digestSplit := strings.Split(refSplit[2], "@")
	if len(digestSplit) > 2 {
		return nil, fmt.Errorf("Unexpected image reference %s", s)
	} else if len(digestSplit) == 2 {
		parsed.Name = digestSplit[0]
		parsed.Digest = digestSplit[1]
		return &parsed, nil
	}
	tagSplit := strings.Split(refSplit[2], ":")
	if len(tagSplit) > 2 {
		return nil, fmt.Errorf("Unexpected image reference %s", s)
	} else if len(tagSplit) == 2 {
		parsed.Tag = tagSplit[1]
	}
	parsed.Name = tagSplit[0]
	return &parsed, nil
}

// SwapContainerImageRefs updates internal image references from
// backup registry to restore registry pathnames
func SwapContainerImageRefs (containers []corev1API.Container, oldRegistry, newRegistry string, log logrus.FieldLogger) {
	for n, container := range containers {
		imageRef := container.Image
		log.Infof("[util] container image ref %s", imageRef)
		newImageRef, err := ReplaceImageRefPrefix(imageRef, oldRegistry, newRegistry)
		if err == nil {
			// Replace local image
			log.Infof("[util] replacing container image ref %s with %s", imageRef, newImageRef)
			containers[n].Image = newImageRef
		}
	}

}
