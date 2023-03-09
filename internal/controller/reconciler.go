package controller

import (
	"fmt"

	"github.com/parseablehq/parseable-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func reconcileParseable(client client.Client, pt *v1beta1.ParseableTenant) error {

	fmt.Println(getAllNodeSpecForNodeType(pt))
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
