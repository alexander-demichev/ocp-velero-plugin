package replicaset

import (
	"encoding/json"

	"github.com/fusor/ocp-velero-plugin/velero-plugins/common"
	"github.com/heptio/velero/pkg/plugin/velero"
	"github.com/sirupsen/logrus"
	appsv1API "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// RestorePlugin is a restore item action plugin for Velero
type RestorePlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a velero.ResourceSelector that applies to replicasets
func (p *RestorePlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedResources: []string{"replicasets.apps"},
	}, nil
}

// Execute action for the restore plugin for the replicaset resource
func (p *RestorePlugin) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	p.Log.Info("[replicaset-restore] Hello from ReplicaSet RestorePlugin!")

	replicaSet := appsv1API.ReplicaSet{}
	itemMarshal, _ := json.Marshal(input.Item)
	json.Unmarshal(itemMarshal, &replicaSet)
	p.Log.Infof("[replicaset-restore] replicaset: %s", replicaSet.Name)

	if input.Restore.Annotations[common.MigrateCopyPhaseAnnotation] != "" {

		backupRegistry, registry, err := common.GetSrcAndDestRegistryInfo(input.Item)
		if err != nil {
			return nil, err
		}
		common.SwapContainerImageRefs(replicaSet.Spec.Template.Spec.Containers, backupRegistry, registry, p.Log)
		common.SwapContainerImageRefs(replicaSet.Spec.Template.Spec.InitContainers, backupRegistry, registry, p.Log)
	}

	var out map[string]interface{}
	objrec, _ := json.Marshal(replicaSet)
	json.Unmarshal(objrec, &out)

	return velero.NewRestoreItemActionExecuteOutput(&unstructured.Unstructured{Object: out}), nil
}
