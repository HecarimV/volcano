package job

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

func TestRunJobTemplate(t *testing.T) {
	response := v1alpha1.Job{}
	jobTemplate := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "JobTemplate",
			"apiVersion": "batch.volcano.sh/v1alpha1",
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.String(), "jobtemplates") {
			val, err := json.Marshal(jobTemplate)
			if err == nil {
				w.Write(val)
			}
			return
		}
		val, err := json.Marshal(response)
		if err == nil {
			w.Write(val)
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	fileName := "testTemplate.yaml"
	val, err := json.Marshal(jobTemplate)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(fileName, val, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer os.Remove(fileName)

	testCases := []struct {
		Name        string
		ExpectValue error
		FileName    string
	}{
		{
			Name:        "RunJob",
			ExpectValue: nil,
		},
		{
			Name:        "RunJobWithFile",
			FileName:    fileName,
			ExpectValue: nil,
		},
	}

	for i, testcase := range testCases {
		launchJobTemplateFlags = &runTemplateFlags{
			commonFlags: commonFlags{
				Master: server.URL,
			},
			TemplateName:      "test",
			TemplateNamespace: "test",
		}
		if testcase.FileName != "" {
			launchJobTemplateFlags.FileName = testcase.FileName
		}

		err := RunJobTemplate()
		if err != nil {
			t.Errorf("case %d (%s): expected: %v, got %v ", i, testcase.Name, testcase.ExpectValue, err)
		}
	}

}

func TestInitTemplateRunFlags(t *testing.T) {
	var cmd cobra.Command
	InitTemplateRunFlags(&cmd)

	if cmd.Flag("namespace") == nil {
		t.Errorf("Could not find the flag namespace")
	}
}
