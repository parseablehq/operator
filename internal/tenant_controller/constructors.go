package parseabletenantcontroller

import (
	"github.com/parseablehq/parseable-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// create owner ref ie parseable tenant controller
func makeOwnerRef(apiVersion, kind, name string, uid types.UID) *metav1.OwnerReference {
	controller := true

	return &metav1.OwnerReference{
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       name,
		UID:        uid,
		Controller: &controller,
	}
}

// NodeType and nodeSpec makes it easier to iterate decisions
// around N nodespec each to a nodeType
type NodeTypeNodeSpec struct {
	NodeType string
	NodeSpec v1beta1.NodeSpec
}

// constructor to nodeTypeNodeSpec. Order is constructed based on the deployment Order
func getAllNodeSpecForNodeType(pt *v1beta1.ParseableTenant) []NodeTypeNodeSpec {

	// add more nodes types
	nodeSpecsByNodeType := map[string][]NodeTypeNodeSpec{
		pt.Spec.DeploymentOrder[0]: make([]NodeTypeNodeSpec, 0, 1),
	}

	for _, nodeSpec := range pt.Spec.Nodes {
		nodeSpecs := nodeSpecsByNodeType[nodeSpec.Type]
		nodeSpecsByNodeType[nodeSpec.Type] = append(nodeSpecs, NodeTypeNodeSpec{nodeSpec.Type, nodeSpec})

	}

	allNodeSpecs := make([]NodeTypeNodeSpec, 0, len(pt.Spec.Nodes))

	allNodeSpecs = append(allNodeSpecs, nodeSpecsByNodeType[pt.Spec.DeploymentOrder[0]]...)

	return allNodeSpecs
}
