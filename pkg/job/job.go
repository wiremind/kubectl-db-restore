package job

import (
	"context"
	"fmt"
	"time"

	"github.com/wiremind/kubectl-db-restore/pkg/k8screds"
	"github.com/wiremind/kubectl-db-restore/pkg/logger"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

type EnvVarSource struct {
	Name      string
	Value     *string                // if set, from env
	SecretRef *k8screds.SecretKeyRef // if set, from secret
}

type JobSpec struct {
	Namespace         string
	JobName           string
	Image             string
	Command           []string
	Args              []string
	EnvVars           []EnvVarSource
	JobSuccessMessage string
	JobFailureHeader  string
}

func CreateJob(configFlags *genericclioptions.ConfigFlags, spec JobSpec) error {
	restConfig, err := configFlags.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes REST config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	return CreateJobWithClient(clientset, spec)
}

func CreateJobWithClient(clientset kubernetes.Interface, spec JobSpec) error {
	// Build env vars
	envVars := []corev1.EnvVar{}
	for _, ev := range spec.EnvVars {
		if ev.SecretRef != nil {
			envVars = append(envVars, corev1.EnvVar{
				Name: ev.Name,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: ev.SecretRef.SecretName,
						},
						Key: ev.SecretRef.Key,
					},
				},
			})
		} else if ev.Value != nil {
			envVars = append(envVars, corev1.EnvVar{
				Name:  ev.Name,
				Value: *ev.Value,
			})
		}
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.JobName,
			Namespace: spec.Namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: int32Ptr(0),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    "task",
							Image:   spec.Image,
							Command: spec.Command,
							Args:    spec.Args,
							Env:     envVars,
						},
					},
				},
			},
		},
	}

	jobClient := clientset.BatchV1().Jobs(spec.Namespace)
	_, err := jobClient.Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create Job: %w", err)
	}

	fmt.Printf("✅ Created Job %s in namespace %s\n", spec.JobName, spec.Namespace)

	// Watch job status
	for {
		jobStatus, err := jobClient.Get(context.TODO(), spec.JobName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get Job status: %w", err)
		}

		if jobStatus.Status.Succeeded > 0 {
			msg := spec.JobSuccessMessage
			if msg == "" {
				msg = "🎉 Job completed successfully!"
			}
			logger.Global.Info("%s", msg)
			break
		}

		if jobStatus.Status.Failed > 0 {
			var failMsg string
			for _, c := range jobStatus.Status.Conditions {
				if c.Type == batchv1.JobFailed {
					failMsg = fmt.Sprintf("❌ Job failed: %s - %s", c.Reason, c.Message)
					break
				}
			}
			if failMsg == "" {
				failMsg = "❌ Job failed for an unknown reason."
			}

			header := spec.JobFailureHeader
			if header == "" {
				header = "💥 Kubernetes Job Failed"
			}

			logger.Global.Instructions(`
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
%s
🔁 Job Name: %s
📂 Namespace: %s

%s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`, header, spec.JobName, spec.Namespace, failMsg)

			return fmt.Errorf("job '%s' failed", spec.JobName)
		}

		logger.Global.Info("⏳ Waiting for Job to complete...")
		time.Sleep(3 * time.Second)
	}

	return nil
}

func int32Ptr(i int32) *int32 { return &i }
