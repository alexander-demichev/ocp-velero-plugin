package deployment

import (
	"encoding/json"

	"github.com/fusor/ocp-velero-plugin/velero-plugins/clients"
	"github.com/fusor/ocp-velero-plugin/velero-plugins/common"
	v1 "github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/plugin/velero"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// BackupPlugin is a backup item action plugin for Heptio Ark.
type BackupPlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a backup.ResourceSelector that applies to everything.
func (p *BackupPlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedResources: []string{"deployments.apps"},
	}, nil
}

// Execute sets a custom annotation on the item being backed up.
func (p *BackupPlugin) Execute(item runtime.Unstructured, backup *v1.Backup) (runtime.Unstructured, []velero.ResourceIdentifier, error) {
	// If this is stage don't do anything
	if backup.Annotations[common.MigrateCopyPhaseAnnotation] != "final" || backup.Annotations[common.MigrateQuiesceAnnotation] != "true" {
		return item, nil, nil
	}

	p.Log.Info("[deployment-backup] Hello from Deployment Backup plugin")

	// Convert to DC
	backupDeployment := appsv1.Deployment{}
	itemMarshal, _ := json.Marshal(item)
	json.Unmarshal(itemMarshal, &backupDeployment)

	client, err := clients.AppsClient()
	if err != nil {
		return nil, nil, err
	}
	// Get and update deployment on the running cluster to have 0 replicas if
	// this is final migration
	deployment, err := client.Deployments(backupDeployment.Namespace).Get(backupDeployment.Name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	// Scale down DC
	zeroReplica := int32(0)
	deployment.Spec.Replicas = &zeroReplica

	// Update DC
	deployment, err = client.Deployments(backupDeployment.Namespace).Update(deployment)
	if err != nil {
		return nil, nil, err
	}

	return item, nil, nil
}
