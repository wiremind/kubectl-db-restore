package job

import (
	"context"
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/assert"
)

func TestCreateJobWithClient_Success(t *testing.T) {
	client := k8sfake.NewSimpleClientset()

	spec := JobSpec{
		Namespace: "default",
		JobName:   "test-job",
		Image:     "alpine",
		Command:   []string{"/bin/sh"},
		Args:      []string{"-c", "echo Hello"},
		EnvVars:   []EnvVarSource{},
	}

	// Simulate successful job status *after* it is created
	go func() {
		// Fetch the job created by CreateJobWithClient so we can update it
		job, _ := client.BatchV1().Jobs(spec.Namespace).Get(context.TODO(), spec.JobName, metav1.GetOptions{})

		job.Status.Succeeded = 1
		if _, err := client.BatchV1().Jobs(spec.Namespace).Update(context.TODO(), job, metav1.UpdateOptions{}); err != nil {
			t.Errorf("failed to update job status: %v", err)
		}
	}()

	err := CreateJobWithClient(client, spec)
	assert.NoError(t, err)
}

func TestCreateJobWithClient_Failure(t *testing.T) {
	client := k8sfake.NewSimpleClientset()

	spec := JobSpec{
		Namespace: "default",
		JobName:   "fail-job",
		Image:     "alpine",
		Command:   []string{"/bin/sh"},
		Args:      []string{"-c", "exit 1"},
	}

	go func() {
		if _, err := client.BatchV1().Jobs(spec.Namespace).Update(context.TODO(), &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      spec.JobName,
				Namespace: spec.Namespace,
			},
			Status: batchv1.JobStatus{
				Failed: 1,
			},
		}, metav1.UpdateOptions{}); err != nil {
			t.Errorf("failed to simulate job failure: %v", err)
		}
	}()

	err := CreateJobWithClient(client, spec)
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "job 'fail-job' failed")

}
