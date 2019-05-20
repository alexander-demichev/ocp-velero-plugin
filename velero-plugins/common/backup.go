package common

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	v1 "github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/plugin/velero"
	"github.com/sirupsen/logrus"
)

// BackupPlugin is a backup item action plugin for Heptio Ark.
type BackupPlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a velero.ResourceSelector that applies to everything.
func (p *BackupPlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{}, nil
}

// Execute sets a custom annotation on the item being backed up.
func (p *BackupPlugin) Execute(item runtime.Unstructured, backup *v1.Backup) (runtime.Unstructured, []velero.ResourceIdentifier, error) {
	p.Log.Info("Hello from common backup plugin!!")

	metadata, annotations, err := getMetadataAndAnnotations(item)
	if err != nil {
		return nil, nil, err
	}

	if backup.Annotations[MigrateCopyPhaseAnnotation] != "" {
		version, err := GetServerVersion()
		if err != nil {
			return nil, nil, err
		}

		annotations[BackupServerVersion] = fmt.Sprintf("%v.%v", version.Major, version.Minor)
		registryHostname, err := GetRegistryInfo(version.Major, version.Minor)
		if err != nil {
			return nil, nil, err
		}
		annotations[BackupRegistryHostname] = registryHostname
		metadata.SetAnnotations(annotations)
	}

	return item, nil, nil
}
