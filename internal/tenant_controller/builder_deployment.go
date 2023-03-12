package parseabletenantcontroller

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type BuilderDeployment struct {
	Replicas int32
	Labels   map[string]string
	PodSpec  *v1.PodSpec
	CommonBuilder
}

func ToNewDeploymentBuilder(builder []BuilderDeployment) func(*Builder) {
	return func(s *Builder) {
		s.Deployment = builder
	}
}

func (b BuilderDeployment) MakeDeployment() (*appsv1.Deployment, error) {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Deployment",
		},
		ObjectMeta: b.ObjectMeta,
		Spec: appsv1.DeploymentSpec{
			Replicas: &b.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: b.Labels,
			},
			Template: v1.PodTemplateSpec{
				Spec: *b.PodSpec,
			},
		},
	}, nil
}

func (s *Builder) BuildDeployment() (controllerutil.OperationResult, error) {

	for _, deploy := range s.Deployment {
		deployment, err := deploy.MakeDeployment()
		if err != nil {
			return controllerutil.OperationResultNone, err
		}

		deploy.DesiredState = deployment
		deploy.CurrentState = &appsv1.Deployment{}

		_, err = deploy.CreateOrUpdate()
		if err != nil {
			return controllerutil.OperationResultNone, nil
		}
	}
	return controllerutil.OperationResultNone, nil
}
