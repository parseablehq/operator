package parseabletenantcontroller

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/parseablehq/parseable-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func reconcileParseable(client client.Client, pt *v1beta1.ParseableTenant) error {

	// create ownerRef passed to each object created
	getOwnerRef := makeOwnerRef(
		pt.APIVersion,
		pt.Kind,
		pt.Name,
		pt.UID,
	)

	nodeSpecs := getAllNodeSpecForNodeType(pt)

	var parseableConfigMap []BuilderConfigMap
	var parseableDeploymentOrStatefulset []BuilderDeploymentStatefulSet

	for _, nodeSpec := range nodeSpecs {
		for _, parseableConfig := range pt.Spec.ParseableConfigGroup {
			if nodeSpec.NodeSpec.ParseableConfigGroup == parseableConfig.Name {
				parseableConfigMap = append(parseableConfigMap, *makeParseableConfigMap(pt, &parseableConfig, client, getOwnerRef))
			}
		}
		for _, k8sConfig := range pt.Spec.K8sConfigGroup {
			fmt.Println(k8sConfig.Name)
			fmt.Println(nodeSpec.NodeSpec.K8sConfigGroup)
			if nodeSpec.NodeSpec.K8sConfigGroup == k8sConfig.Name {

				parseableDeploymentOrStatefulset = append(parseableDeploymentOrStatefulset, *makeStsOrDeploy(pt, &nodeSpec.NodeSpec, &k8sConfig, client, getOwnerRef))
			}
		}
	}

	builder := NewBuilder(
		ToNewConfigMapBuilder(parseableConfigMap),
		ToNewDeploymentStatefulSetBuilder(parseableDeploymentOrStatefulset),
	)

	_, err := builder.BuildConfigMap()
	if err != nil {
		return err
	}

	resultDeploy, err := builder.BuildDeployOrSts()
	if err != nil {
		return err
	}
	fmt.Printf("Deploy %s", resultDeploy)

	return nil
}

func makeExternalConfigMap(pt *v1beta1.ParseableTenant, client client.Client, ownerRef *metav1.OwnerReference) *BuilderConfigMap {
	return &BuilderConfigMap{
		CommonBuilder: CommonBuilder{
			ObjectMeta: metav1.ObjectMeta{Name: pt.GetName() + "-external-cm",
				Namespace: pt.GetNamespace()},
			Client:   client,
			CrObject: pt,
			OwnerRef: *ownerRef,
		},
		Data: map[string]string{
			"data": fmt.Sprintf("%s", pt.Spec.External.ObjectStore.Spec.Data),
		},
	}
}

func makeParseableConfigMap(
	pt *v1beta1.ParseableTenant,
	parseableConfigGroup *v1beta1.ParseableConfigGroupSpec,
	client client.Client,
	ownerRef *metav1.OwnerReference) *BuilderConfigMap {

	configMap := &BuilderConfigMap{
		CommonBuilder: CommonBuilder{
			ObjectMeta: metav1.ObjectMeta{Name: parseableConfigGroup.Name + "-parseable-cm",
				Namespace: pt.GetNamespace()},
			Client:   client,
			CrObject: pt,
			OwnerRef: *ownerRef,
		},
		Data: map[string]string{
			"data": fmt.Sprintf("%s", parseableConfigGroup.Data),
		},
	}

	return configMap
}

func makeStsOrDeploy(
	pt *v1beta1.ParseableTenant,
	ptNode *v1beta1.NodeSpec,
	k8sConfigGroup *v1beta1.K8sConfigGroupSpec,
	client client.Client,
	ownerRef *metav1.OwnerReference) *BuilderDeploymentStatefulSet {

	deployment := BuilderDeploymentStatefulSet{
		CommonBuilder: CommonBuilder{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ptNode.K8sConfigGroup + ptNode.Name,
				Namespace: pt.GetNamespace(),
				Labels: map[string]string{
					"app": "parseable",
				},
			},
			Client:   client,
			CrObject: pt,
			OwnerRef: *ownerRef,
		},
		Replicas: int32(ptNode.Replicas),
		Labels: map[string]string{
			"app": "parseable",
		},
		Kind:    ptNode.Kind,
		PodSpec: &k8sConfigGroup.Spec,
	}

	return &deployment
}
