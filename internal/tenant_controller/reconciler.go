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

	for _, nodeSpec := range nodeSpecs {
		for _, parseableConfig := range pt.Spec.ParseableConfigGroup {
			if nodeSpec.NodeSpec.ParseableConfigGroup == parseableConfig.Name {
				parseableConfigMap = append(parseableConfigMap, *makeParseableConfigMaps(pt, &parseableConfig, client, getOwnerRef))
			}
		}
	}
	// var cmBuilder []BuilderConfigMap

	// parseableConfigMaps := makeParseableConfigMaps(pt, client, getOwnerRef)

	// for _, parseableConfigMap := range *parseableConfigMaps {
	// 	cmBuilder = append(cmBuilder, parseableConfigMap, *makeExternalConfigMap(pt, client, &parseableConfigMap.OwnerRef))
	// }

	builder := NewBuilder(
		ToNewConfigMapBuilder(parseableConfigMap),
	)

	result, err := builder.BuildConfigMap()
	if err != nil {
		return err
	}

	fmt.Printf("%s", result)

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

func makeParseableConfigMaps(
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

// func makeK8sConfigGroup(pt *v1beta1.ParseableTenant, client client.Client, ownerRef *metav1.OwnerReference) *[]BuilderDeployment {
// 	var parseableK8sConfigs []BuilderDeployment

// 	for _, parseableK8sConfig := range pt.Spec.K8sConfigGroup {
// 		parseableK8sConfigs = append(parseableK8sConfigs, BuilderDeployment{
// 			Replicas: getReplicasForK8sConfigName(parseableK8sConfig.Name),
// 		})
// 	}

// 	return &parseableK8sConfigs
// }
