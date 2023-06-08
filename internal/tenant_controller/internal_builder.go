package parseabletenantcontroller

import (
	"fmt"

	"github.com/datainfrahq/operator-runtime/builder"
	"github.com/datainfrahq/operator-runtime/utils"
	"github.com/parseablehq/operator/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ib interface {
	makeExternalConfigMap() *builder.BuilderConfigMap
	makeParseableConfigMap(parseableConfig *v1beta1.ParseableConfigSpec) *builder.BuilderConfigMap
	makeStsOrDeploy(ptNode *v1beta1.NodeSpec, k8sConfigGroup *v1beta1.K8sConfigSpec, storageConfig *[]v1beta1.StorageConfig, parseableConfig *v1beta1.ParseableConfigSpec) *builder.BuilderDeploymentStatefulSet
	makePvc(pvc *v1beta1.StorageConfig) *builder.BuilderStorageConfig
	makeService(k8sConfig *v1beta1.K8sConfigSpec, selectorLabel map[string]string) *builder.BuilderService
}

type internalBuilder struct {
	parseableTenant *v1beta1.ParseableTenant
	client          client.Client
	ownerRef        *metav1.OwnerReference
	commonLabels    map[string]string
}

func newInternalBuilder(
	pt *v1beta1.ParseableTenant,
	client client.Client,
	nodeSpec *v1beta1.NodeSpec,
	ownerRef *metav1.OwnerReference) *internalBuilder {
	return &internalBuilder{
		parseableTenant: pt,
		client:          client,
		ownerRef:        ownerRef,
		commonLabels:    makeLabels(pt, nodeSpec),
	}
}

func (ib *internalBuilder) makeExternalConfigMap() *builder.BuilderConfigMap {
	return &builder.BuilderConfigMap{
		CommonBuilder: builder.CommonBuilder{
			ObjectMeta: metav1.ObjectMeta{
				Name:      makeConfigMapNameExternal(ib.parseableTenant.GetName()),
				Namespace: ib.parseableTenant.GetNamespace(),
				Labels:    ib.commonLabels,
			},
			Client:   ib.client,
			CrObject: ib.parseableTenant,
			OwnerRef: *ib.ownerRef,
		},
		Data: map[string]string{
			"data": fmt.Sprintf("%s", ib.parseableTenant.Spec.External.ObjectStore.Spec.Data),
		},
	}
}

func (ib *internalBuilder) makeParseableConfigMap(
	parseableConfig *v1beta1.ParseableConfigSpec,
	ptNode *v1beta1.NodeSpec,
) *builder.BuilderConfigMap {

	configMap := &builder.BuilderConfigMap{
		CommonBuilder: builder.CommonBuilder{
			ObjectMeta: metav1.ObjectMeta{
				Name:      makeConfigMapName(ptNode.Name, parseableConfig.Name),
				Namespace: ib.parseableTenant.GetNamespace(),
				Labels:    ib.commonLabels,
			},
			Client:   ib.client,
			CrObject: ib.parseableTenant,
			OwnerRef: *ib.ownerRef,
		},
		Data: parseableConfig.EnvVars,
	}

	return configMap
}

func (ib *internalBuilder) makeStsOrDeploy(
	ptNode *v1beta1.NodeSpec,
	k8sConfigSpec *v1beta1.K8sConfigSpec,
	storageConfig *[]v1beta1.StorageConfig,
	parseableConfigSpec *v1beta1.ParseableConfigSpec,
	configHash []utils.ConfigMapHash,
) *builder.BuilderDeploymentStatefulSet {

	b := false
	args := []string{"parseable"}
	args = append(args, parseableConfigSpec.CliArgs...)

	var envFrom []v1.EnvFromSource
	configCm := v1.EnvFromSource{
		ConfigMapRef: &v1.ConfigMapEnvSource{
			LocalObjectReference: v1.LocalObjectReference{
				Name: makeConfigMapName(ptNode.Name, parseableConfigSpec.Name),
			},
		},
	}
	envFrom = append(envFrom, configCm)

	if ib.parseableTenant.Spec.External != (v1beta1.ExternalSpec{}) {
		externalCm := v1.EnvFromSource{
			ConfigMapRef: &v1.ConfigMapEnvSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: makeConfigMapNameExternal(ib.parseableTenant.GetName()),
				},
			},
		}
		envFrom = append(envFrom, externalCm)
	}

	var runner int64 = 1000
	fsPolicy := v1.FSGroupChangeAlways

	podSpec := v1.PodSpec{
		NodeSelector: k8sConfigSpec.NodeSelector,
		Tolerations:  getTolerations(k8sConfigSpec),
		SecurityContext: &v1.PodSecurityContext{
			RunAsUser:           &runner,
			RunAsGroup:          &runner,
			FSGroup:             &runner,
			FSGroupChangePolicy: &fsPolicy,
		},
		Containers: []v1.Container{
			{
				Name:            ptNode.Name + "-" + ptNode.Type,
				Image:           k8sConfigSpec.Image,
				Args:            args,
				ImagePullPolicy: k8sConfigSpec.ImagePullPolicy,
				SecurityContext: &v1.SecurityContext{
					AllowPrivilegeEscalation: &b,
				},
				Ports: []v1.ContainerPort{
					{
						ContainerPort: 8000,
					},
				},
				Env:          getEnv(ib.parseableTenant, parseableConfigSpec, k8sConfigSpec, configHash),
				EnvFrom:      envFrom,
				VolumeMounts: getVolumeMounts(k8sConfigSpec, storageConfig),
				Resources:    k8sConfigSpec.Resources,
			},
		},
		Volumes:            getVolume(k8sConfigSpec, storageConfig, ptNode),
		ServiceAccountName: k8sConfigSpec.ServiceAccountName,
	}

	deployment := builder.BuilderDeploymentStatefulSet{
		CommonBuilder: builder.CommonBuilder{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ptNode.K8sConfig + "-" + ptNode.Name,
				Namespace: ib.parseableTenant.GetNamespace(),
				Labels:    ib.commonLabels,
			},
			Client:   ib.client,
			CrObject: ib.parseableTenant,
			OwnerRef: *ib.ownerRef,
		},
		Replicas: int32(ptNode.Replicas),
		Labels:   ib.commonLabels,
		Kind:     ptNode.Kind,
		PodSpec:  &podSpec,
	}

	return &deployment
}

func (ib *internalBuilder) makePvc(
	sc *v1beta1.StorageConfig,
	k8sConfig *v1beta1.K8sConfigSpec,
	ptNode *v1beta1.NodeSpec,
) *builder.BuilderStorageConfig {
	return &builder.BuilderStorageConfig{
		CommonBuilder: builder.CommonBuilder{
			ObjectMeta: metav1.ObjectMeta{
				Name:      makePvcName(ptNode.Name, k8sConfig.Name, sc.Name),
				Namespace: ib.parseableTenant.GetNamespace()},
			Client:   ib.client,
			CrObject: ib.parseableTenant,
			Labels:   ib.commonLabels,
			OwnerRef: *ib.ownerRef,
		},
		PvcSpec: &sc.PvcSpec,
	}
}

func (ib *internalBuilder) makeService(
	k8sConfig *v1beta1.K8sConfigSpec,
	nodeSpec *v1beta1.NodeSpec,
) *builder.BuilderService {
	return &builder.BuilderService{
		CommonBuilder: builder.CommonBuilder{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nodeSpec.Name + "-" + k8sConfig.Name,
				Namespace: ib.parseableTenant.GetNamespace()},
			Client:   ib.client,
			CrObject: ib.parseableTenant,
			OwnerRef: *ib.ownerRef,
			Labels:   ib.commonLabels,
		},
		SelectorLabels: ib.commonLabels,
		ServiceSpec:    k8sConfig.Service,
	}
}

func makeLabels(pt *v1beta1.ParseableTenant, nodeSpec *v1beta1.NodeSpec) map[string]string {

	return map[string]string{
		"app":             "parseable",
		"custom_resource": pt.Name,
		"nodeType":        nodeSpec.Type,
		"parseableConfig": nodeSpec.ParseableConfig,
		"k8sConfigGroup":  nodeSpec.K8sConfig,
	}
}

func getTolerations(k8sConfig *v1beta1.K8sConfigSpec) []v1.Toleration {
	tolerations := []v1.Toleration{}
	return append(tolerations, k8sConfig.Tolerations...)
}

func getVolumeMounts(k8sConfig *v1beta1.K8sConfigSpec, storageConfig *[]v1beta1.StorageConfig) []v1.VolumeMount {

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

func getVolume(
	k8sConfig *v1beta1.K8sConfigSpec,
	storageConfig *[]v1beta1.StorageConfig,
	ptNode *v1beta1.NodeSpec,
) []v1.Volume {
	var volumeHolder = []v1.Volume{}

	for _, sc := range *storageConfig {
		volumeHolder = append(volumeHolder, v1.Volume{
			Name: sc.Name,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: makePvcName(ptNode.Name, k8sConfig.Name, sc.Name),
				}},
		})
	}

	volumeHolder = append(volumeHolder, k8sConfig.Volumes...)
	return volumeHolder
}

func getEnv(
	pt *v1beta1.ParseableTenant,
	parseableConfigSpec *v1beta1.ParseableConfigSpec,
	k8sConfigSpec *v1beta1.K8sConfigSpec,
	configHash []utils.ConfigMapHash,
) []v1.EnvVar {
	var envs, hashHolder []v1.EnvVar
	envs = append(envs, k8sConfigSpec.Env...)

	hashes, _ := utils.MakeConfigMapHash(configHash)

	for _, cmhash := range hashes {
		if makeConfigMapName(pt.Name, parseableConfigSpec.Name) == cmhash.Name {
			hashHolder = append(hashHolder, v1.EnvVar{Name: cmhash.Name, Value: cmhash.HashVaule})
		}
	}
	envs = append(envs, hashHolder...)
	return envs
}

func makeConfigMapName(nodeName, configGroupName string) string {
	return nodeName + "-" + configGroupName
}

func makeConfigMapNameExternal(crName string) string { return crName + "-" + "ext" }

func makePvcName(nodeName, k8sConfigGroupName, storageConfigName string) string {
	return nodeName + "-" + k8sConfigGroupName + "-" + storageConfigName
}
