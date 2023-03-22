package parseabletenantcontroller

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/parseablehq/parseable-operator/api/v1beta1"
	"github.com/parseablehq/parseable-operator/pkg/operator-builder/builder"
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

	parseableConfigMap := []builder.BuilderConfigMap{}
	parseableConfigMapHash := []builder.BuilderConfigMapHash{}
	parseableDeploymentOrStatefulset := []builder.BuilderDeploymentStatefulSet{}
	parseableStorage := []builder.BuilderStorageConfig{}

	// For all the nodeSpec ie nodeType to nodeSpec
	// Get all the config group defined and append to configMap builder
	// For each config group defined create a configmap hash and append to configmaphash builder
	// Get all the k8s config group defined and append to deploymentstatefulset builder
	// For all the storage config defined in k8s config group append
	for _, nodeSpec := range nodeSpecs {
		for _, parseableConfig := range pt.Spec.ParseableConfigGroup {
			if nodeSpec.NodeSpec.ParseableConfigGroup == parseableConfig.Name {
				cm := *makeParseableConfigMap(pt, &parseableConfig, client, getOwnerRef)
				parseableConfigMap = append(parseableConfigMap, cm)
				parseableConfigMapHash = append(parseableConfigMapHash, builder.BuilderConfigMapHash{Object: &v1.ConfigMap{Data: cm.Data, ObjectMeta: cm.ObjectMeta}})
				for _, k8sConfig := range pt.Spec.K8sConfigGroup {
					if nodeSpec.NodeSpec.K8sConfigGroup == k8sConfig.Name {
						parseableDeploymentOrStatefulset = append(parseableDeploymentOrStatefulset, *makeStsOrDeploy(pt, &nodeSpec.NodeSpec, &k8sConfig, &k8sConfig.StorageConfig, &parseableConfig, client, getOwnerRef))
						for _, sc := range k8sConfig.StorageConfig {
							parseableStorage = append(parseableStorage, *makePvc(pt, client, getOwnerRef, &sc))
						}
					}
				}
			}
		}
	}

	// append external config and hash to configmap builder
	if pt.Spec.External != (v1beta1.ExternalSpec{}) {
		cm := *makeExternalConfigMap(pt, client, getOwnerRef)
		parseableConfigMap = append(parseableConfigMap, cm)
		parseableConfigMapHash = append(parseableConfigMapHash, builder.BuilderConfigMapHash{Object: cm.DesiredState})

	}

	// construct builder
	builder := builder.NewBuilder(
		builder.ToNewConfigMapBuilder(parseableConfigMap),
		builder.ToNewConfigMapHashBuilder(parseableConfigMapHash),
		builder.ToNewDeploymentStatefulSetBuilder(parseableDeploymentOrStatefulset),
		builder.ToNewBuilderStorageConfig(parseableStorage),
	)

	// All builder methods called are responsible for reconciling
	// and triggering reconcilers in case of state change.

	// build configmap
	resultCm, err := builder.BuildConfigMap()
	if err != nil {
		return err
	}

	fmt.Printf("Cm %s", resultCm)

	// build configmap hash
	cmhashes, err := builder.BuildConfigMapHash()
	if err != nil {
		return err
	}

	// build depoyment or statefulset
	resultDeploy, err := builder.BuildDeployOrSts(cmhashes)
	if err != nil {
		return err
	}
	fmt.Printf("Deploy %s", resultDeploy)

	// build storage
	resultPvc, err := builder.BuildPvc()
	if err != nil {
		return err
	}
	fmt.Printf("Pvc %s", resultPvc)

	return nil
}

func makeExternalConfigMap(pt *v1beta1.ParseableTenant,
	client client.Client,
	ownerRef *metav1.OwnerReference,
) *builder.BuilderConfigMap {
	return &builder.BuilderConfigMap{
		CommonBuilder: builder.CommonBuilder{
			ObjectMeta: metav1.ObjectMeta{Name: pt.GetName() + "-external",
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
	ownerRef *metav1.OwnerReference) *builder.BuilderConfigMap {

	configMap := &builder.BuilderConfigMap{
		CommonBuilder: builder.CommonBuilder{
			ObjectMeta: metav1.ObjectMeta{Name: parseableConfigGroup.Name,
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
	storageConfig *[]v1beta1.StorageConfig,
	parseableConfigGroup *v1beta1.ParseableConfigGroupSpec,
	client client.Client,
	ownerRef *metav1.OwnerReference) *builder.BuilderDeploymentStatefulSet {

	var args = []string{"parseable"}

	for _, arg := range ptNode.CliArgs {
		args = append(args, arg)
	}

	b := false

	var envFrom []v1.EnvFromSource
	configCm := v1.EnvFromSource{
		ConfigMapRef: &v1.ConfigMapEnvSource{
			LocalObjectReference: v1.LocalObjectReference{
				Name: parseableConfigGroup.Name,
			},
		},
	}
	envFrom = append(envFrom, configCm)

	if pt.Spec.External != (v1beta1.ExternalSpec{}) {
		externalCm := v1.EnvFromSource{
			ConfigMapRef: &v1.ConfigMapEnvSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: pt.Name + "external-cm",
				},
			},
		}
		envFrom = append(envFrom, externalCm)

	}

	podSpec := v1.PodSpec{
		NodeSelector: k8sConfigGroup.NodeSelector,
		Tolerations:  getTolerations(k8sConfigGroup),
		Containers: []v1.Container{

			{
				Name:            ptNode.NodeType,
				Image:           k8sConfigGroup.Image,
				Args:            args,
				ImagePullPolicy: k8sConfigGroup.ImagePullPolicy,
				SecurityContext: &v1.SecurityContext{
					AllowPrivilegeEscalation: &b,
				},
				Ports: []v1.ContainerPort{
					{
						ContainerPort: 8000,
					},
				},
				Env: getEnv(*k8sConfigGroup),
				EnvFrom: []v1.EnvFromSource{
					{
						ConfigMapRef: &v1.ConfigMapEnvSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: parseableConfigGroup.Name,
							},
						},
					},
				},
				VolumeMounts: getVolumeMounts(k8sConfigGroup, storageConfig),
			},
		},
		Volumes:            getVolume(k8sConfigGroup, storageConfig),
		ServiceAccountName: k8sConfigGroup.ServiceAccountName,
	}

	deployment := builder.BuilderDeploymentStatefulSet{
		CommonBuilder: builder.CommonBuilder{
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
		PodSpec: &podSpec,
	}

	return &deployment
}

func makePvc(
	pt *v1beta1.ParseableTenant,
	client client.Client,
	ownerRef *metav1.OwnerReference,
	pvc *v1beta1.StorageConfig,
) *builder.BuilderStorageConfig {
	return &builder.BuilderStorageConfig{
		CommonBuilder: builder.CommonBuilder{
			ObjectMeta: metav1.ObjectMeta{Name: pvc.Name,
				Namespace: pt.GetNamespace()},
			Client:   client,
			CrObject: pt,
			OwnerRef: *ownerRef,
		},
		PvcSpec: &pvc.PvcSpec,
	}
}

func getTolerations(k8sConfig *v1beta1.K8sConfigGroupSpec) []v1.Toleration {
	tolerations := []v1.Toleration{}

	for _, val := range k8sConfig.Tolerations {
		tolerations = append(tolerations, val)
	}

	return tolerations
}

func getVolumeMounts(k8sConfig *v1beta1.K8sConfigGroupSpec, storageConfig *[]v1beta1.StorageConfig) []v1.VolumeMount {

	var volumeMount = []v1.VolumeMount{}
	for _, sc := range *storageConfig {
		volumeMount = append(volumeMount, v1.VolumeMount{
			MountPath: sc.MountPath,
			Name:      sc.Name,
		})
	}

	volumeMount = append(volumeMount, k8sConfig.VolumeMount...)
	return volumeMount
}

func getVolume(k8sConfig *v1beta1.K8sConfigGroupSpec, storageConfig *[]v1beta1.StorageConfig) []v1.Volume {
	var volumeHolder = []v1.Volume{}

	for _, sc := range *storageConfig {
		volumeHolder = append(volumeHolder, v1.Volume{
			Name: sc.Name,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: sc.Name,
				}},
		})
	}

	volumeHolder = append(volumeHolder, k8sConfig.Volumes...)
	return volumeHolder
}

func getEnv(k8sConfigGroup v1beta1.K8sConfigGroupSpec) []v1.EnvVar {
	var envHolder []v1.EnvVar

	for _, env := range k8sConfigGroup.Env {
		envHolder = append(envHolder, env)
	}

	return envHolder
}
