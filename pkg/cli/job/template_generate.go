package job

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
	"strings"
	vcbatch "volcano.sh/apis/pkg/apis/batch/v1alpha1"
	"volcano.sh/apis/pkg/client/clientset/versioned"
	"volcano.sh/volcano/pkg/cli/util"
)

type generateTemplateFlags struct {
	commonFlags
	FileName     string
	JobName      string
	Namespace    string
	GenerateName string
}

var generateJobTemplateFlags = &generateTemplateFlags{}

// InitTemplateGenerateFlags init the generate flags.
func InitTemplateGenerateFlags(cmd *cobra.Command) {
	initFlags(cmd, &generateJobTemplateFlags.commonFlags)

	cmd.Flags().StringVarP(&generateJobTemplateFlags.FileName, "filename", "f", "", "the yaml file of job")
	cmd.Flags().StringVarP(&generateJobTemplateFlags.Namespace, "namespace", "n", "default", "the namespace of job")
	cmd.Flags().StringVarP(&generateJobTemplateFlags.JobName, "name", "N", "", "the name of job")
	cmd.Flags().StringVarP(&generateJobTemplateFlags.GenerateName, "generateName", "g", "", "the name for new job template")
}

func GenerateJobTemplate() error {
	config, err := util.BuildConfig(generateJobTemplateFlags.Master, generateJobTemplateFlags.Kubeconfig)
	if err != nil {
		return err
	}

	if generateJobTemplateFlags.JobName == "" && generateJobTemplateFlags.FileName == "" {
		err = fmt.Errorf("the filename and job name cannot both be left blank")
		return err
	}

	var job *vcbatch.Job
	if generateJobTemplateFlags.FileName != "" {
		job, err = readJobFile(generateJobTemplateFlags.FileName)
		if err != nil {
			return err
		}
	}

	if job == nil {
		jobClient := versioned.NewForConfigOrDie(config)
		job, err = jobClient.BatchV1alpha1().Jobs(generateJobTemplateFlags.Namespace).Get(context.TODO(), generateJobTemplateFlags.JobName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("fail to get job <%s/%s>", generateJobTemplateFlags.Namespace, generateJobTemplateFlags.JobName)
		}
	}

	unstructuredContent, err := runtime.DefaultUnstructuredConverter.ToUnstructured(job)
	if err != nil {
		return fmt.Errorf("fail to convert job <%s/%s> for：%v", job.Namespace, job.Name, err)
	}
	jobTemplate := &unstructured.Unstructured{Object: unstructuredContent}
	if generateJobTemplateFlags.GenerateName != "" {
		jobTemplate.SetName(generateJobTemplateFlags.GenerateName)
	}
	jobTemplate.SetKind("JobTemplate")
	if jobTemplate.GetNamespace() == "" {
		jobTemplate.SetNamespace("default")
	}
	jobTemplate.SetAPIVersion("batch.volcano.sh/v1alpha1")
	jobTemplate.SetManagedFields(nil)
	jobTemplate.SetAnnotations(nil)
	jobTemplate.SetGeneration(1)
	jobTemplate.SetResourceVersion("")
	jobTemplate.SetUID("")
	myDynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}
	myDynamicResourceClient := myDynamicClient.
		Resource(schema.GroupVersionResource{Group: "batch.volcano.sh", Version: "v1alpha1", Resource: "jobtemplates"}).
		Namespace(jobTemplate.GetNamespace())
	createdTemplate, err := myDynamicResourceClient.Create(context.TODO(), jobTemplate, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("fail to create jobTemplate <%s/%s> for：%v", jobTemplate.GetNamespace(), jobTemplate.GetName(), err)
	}
	fmt.Printf("%s/%s created\n", createdTemplate.GetKind(), createdTemplate.GetName())

	return nil
}

func readJobFile(filename string) (*vcbatch.Job, error) {

	if !strings.Contains(filename, ".yaml") && !strings.Contains(filename, ".yml") {
		return nil, fmt.Errorf("only support yaml file")
	}

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file, err: %v", err)
	}

	var job vcbatch.Job
	if err := yaml.Unmarshal(file, &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file, err:  %v", err)
	}

	return &job, nil
}
