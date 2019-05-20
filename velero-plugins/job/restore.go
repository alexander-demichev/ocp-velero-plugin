package job

import (
	"encoding/json"

	"github.com/fusor/ocp-velero-plugin/velero-plugins/common"
	"github.com/heptio/velero/pkg/plugin/velero"
	"github.com/sirupsen/logrus"
	batchv1API "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// RestorePlugin is a restore item action plugin for Velero
type RestorePlugin struct {
	Log logrus.FieldLogger
}

// AppliesTo returns a velero.ResourceSelector that applies to jobs
func (p *RestorePlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedResources: []string{"jobs"},
	}, nil
}

// Execute action for the restore plugin for the job resource
func (p *RestorePlugin) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	p.Log.Info("[job-restore] Hello from Job RestorePlugin!")

	job := batchv1API.Job{}
	itemMarshal, _ := json.Marshal(input.Item)
	json.Unmarshal(itemMarshal, &job)
	p.Log.Infof("[job-restore] job: %s", job.Name)

	if input.Restore.Annotations[common.MigrateCopyPhaseAnnotation] != "" {

		backupRegistry, registry, err := common.GetSrcAndDestRegistryInfo(input.Item)
		if err != nil {
			return nil, err
		}
		common.SwapContainerImageRefs(job.Spec.Template.Spec.Containers, backupRegistry, registry, p.Log)
		common.SwapContainerImageRefs(job.Spec.Template.Spec.InitContainers, backupRegistry, registry, p.Log)
	}

	var out map[string]interface{}
	objrec, _ := json.Marshal(job)
	json.Unmarshal(objrec, &out)

	return velero.NewRestoreItemActionExecuteOutput(&unstructured.Unstructured{Object: out}), nil
}
