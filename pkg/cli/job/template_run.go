package job

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
	"strings"
	vcbatch "volcano.sh/apis/pkg/apis/batch/v1alpha1"
	"volcano.sh/apis/pkg/client/clientset/versioned"
	"volcano.sh/volcano/pkg/cli/util"
)

type runTemplateFlags struct {
	commonFlags
	FileName          string
	TemplateName      string
	TemplateNamespace string
	GenerateName      string
}

const createSourceKey = "volcano.sh/createByJobTemplate"

var launchJobTemplateFlags = &runTemplateFlags{}

// InitTemplateRunFlags init the run flags.
func InitTemplateRunFlags(cmd *cobra.Command) {
	initFlags(cmd, &launchJobTemplateFlags.commonFlags)

	cmd.Flags().StringVarP(&launchJobTemplateFlags.FileName, "filename", "f", "", "the yaml file of jobTemplate")
	cmd.Flags().StringVarP(&launchJobTemplateFlags.TemplateNamespace, "namespace", "n", "default", "the namespace of job template")
	cmd.Flags().StringVarP(&launchJobTemplateFlags.TemplateName, "name", "N", "", "the name of job template")
	cmd.Flags().StringVarP(&launchJobTemplateFlags.GenerateName, "generateName", "g", "", "the name for new job")
}

func RunJobTemplate() error {
	config, err := util.BuildConfig(launchJobTemplateFlags.Master, launchJobTemplateFlags.Kubeconfig)
	if err != nil {
		return err
	}
	jobTemplate, err := readTemplateFile(launchJobTemplateFlags.FileName)
	if err != nil {
		return err
	}
	if jobTemplate == nil {
		if launchJobTemplateFlags.TemplateName == "" {
			return fmt.Errorf("the filename and template name cannot both be empty")
		}
		myDynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return err
		}
		myDynamicResourceClient := myDynamicClient.
			Resource(schema.GroupVersionResource{Group: "batch.volcano.sh", Version: "v1alpha1", Resource: "jobtemplates"}).
			Namespace(launchJobTemplateFlags.TemplateNamespace)
		jobTemplate, err = myDynamicResourceClient.Get(context.TODO(), launchJobTemplateFlags.TemplateName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("fail to get unstructed jobTemplate for: %v", err)
		}
	}
	jobTemplate.SetKind("Job")
	jobTemplate.SetManagedFields(nil)
	jobTemplate.SetGeneration(1)
	jobTemplate.SetResourceVersion("")
	jobTemplate.SetUID("")
	if jobTemplate.GetNamespace() == "" {
		jobTemplate.SetNamespace("default")
	}
	annotations := jobTemplate.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[createSourceKey] = fmt.Sprintf("%s.%s", jobTemplate.GetNamespace(), jobTemplate.GetName())
	jobTemplate.SetAnnotations(annotations)
	var job vcbatch.Job
	result, err := yaml.Marshal(jobTemplate.Object)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(result, &job); err != nil {
		return err
	}
	if launchJobTemplateFlags.GenerateName != "" {
		job.Name = launchJobTemplateFlags.GenerateName
	}
	jobClient := versioned.NewForConfigOrDie(config)
	newJob, err := jobClient.BatchV1alpha1().Jobs(launchJobTemplateFlags.TemplateNamespace).Create(context.TODO(), &job, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	if newJob.Spec.Queue == "" {
		newJob.Spec.Queue = "default"
	}

	fmt.Printf("run job %v successfully\n", newJob.Name)

	return nil
}

func readTemplateFile(filename string) (*unstructured.Unstructured, error) {
	if filename == "" {
		return nil, nil
	}

	if !strings.Contains(filename, ".yaml") && !strings.Contains(filename, ".yml") {
		return nil, fmt.Errorf("only support yaml file")
	}

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file, err: %v", err)
	}

	jobTemplate := &unstructured.Unstructured{Object: map[string]interface{}{}}
	if err := yaml.Unmarshal(file, &jobTemplate.Object); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file, err:  %v", err)
	}

	return jobTemplate, nil
}
