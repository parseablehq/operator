package controller

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/parseablehq/parseable-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func reconcileParseable(client client.Client, pt *v1beta1.ParseableTenant) error {

	controller := true

	OwnerRef := metav1.OwnerReference{
		APIVersion: pt.APIVersion,
		Kind:       pt.Kind,
		Name:       pt.Name,
		UID:        pt.UID,
		Controller: &controller,
	}
	//nodeTypeNodeSpecs := getAllNodeSpecForNodeType(pt)

	cmBuilder := NewBuilderObject(
		pt,
		metav1.ObjectMeta{Name: pt.GetName() + "-external-cm",
			Namespace: pt.GetNamespace()},
		map[string]string{
			"data": fmt.Sprintf("%s", pt.Spec.External.ObjectStore.Spec.Data),
		})
	cm, err := cmBuilder.MakeConfigMap()
	if err != nil {
		return err
	}

	cmBuilderState := BuilderState{
		Client:       client,
		DesiredState: cm,
		CurrentState: &v1.ConfigMap{},
		CrObject:     pt,
		OwnerRef:     OwnerRef,
	}

	_, err = cmBuilderState.CreateOrUpdate()
	if err != nil {
		return err
	}

	for _, parseableConfig := range pt.Spec.ParseableConfigGroup {
		cmBuilder := NewBuilderObject(
			pt,
			metav1.ObjectMeta{Name: parseableConfig.Name + "-parseable-configs-cm",
				Namespace: pt.GetNamespace()},
			map[string]string{
				"data": fmt.Sprintf("%s", parseableConfig.Data),
			})
		cm, err := cmBuilder.MakeConfigMap()
		if err != nil {
			return err
		}

		cmBuilderState := BuilderState{
			Client:       client,
			DesiredState: cm,
			CurrentState: &v1.ConfigMap{},
			CrObject:     pt,
			OwnerRef:     OwnerRef,
		}

		result, err := cmBuilderState.CreateOrUpdate()
		if err != nil {
			return err
		}
		fmt.Println(result)
	}

	return nil
}

type NodeTypeNodeSpec struct {
	NodeType string
	NodeSpec v1beta1.NodeSpec
}

func getAllNodeSpecForNodeType(pt *v1beta1.ParseableTenant) []NodeTypeNodeSpec {

	nodeSpecsByNodeType := map[string][]NodeTypeNodeSpec{
		pt.Spec.DeploymentOrder[0]: make([]NodeTypeNodeSpec, 0, 1),
		pt.Spec.DeploymentOrder[1]: make([]NodeTypeNodeSpec, 0, 1),
	}

	for _, nodeSpec := range pt.Spec.Nodes {
		nodeSpecs := nodeSpecsByNodeType[nodeSpec.NodeType]
		nodeSpecsByNodeType[nodeSpec.NodeType] = append(nodeSpecs, NodeTypeNodeSpec{nodeSpec.NodeType, nodeSpec})

	}

	allNodeSpecs := make([]NodeTypeNodeSpec, 0, len(pt.Spec.Nodes))
	allNodeSpecs = append(allNodeSpecs, nodeSpecsByNodeType[pt.Spec.DeploymentOrder[0]]...)
	allNodeSpecs = append(allNodeSpecs, nodeSpecsByNodeType[pt.Spec.DeploymentOrder[1]]...)

	return allNodeSpecs
}
