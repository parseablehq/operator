package parseabletenantcontroller

import (
	"fmt"

	"github.com/datainfrahq/operator-builder/builder"
	"github.com/datainfrahq/operator-builder/utils"
	"github.com/parseablehq/parseable-operator/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ib interface {
	makeExternalConfigMap() *builder.BuilderConfigMap
	makeParseableConfigMap(parseableConfigGroup *v1beta1.ParseableConfigGroupSpec) *builder.BuilderConfigMap
	makeStsOrDeploy(ptNode *v1beta1.NodeSpec, k8sConfigGroup *v1beta1.K8sConfigGroupSpec, storageConfig *[]v1beta1.StorageConfig, parseableConfigGroup *v1beta1.ParseableConfigGroupSpec) *builder.BuilderDeploymentStatefulSet
	makePvc(pvc *v1beta1.StorageConfig) *builder.BuilderStorageConfig
	makeService(k8sConfig *v1beta1.K8sConfigGroupSpec, selectorLabel map[string]string) *builder.BuilderService
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
				Name:      ib.parseableTenant.GetName() + "-external",
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

func (ib *internalBuilder) makeParseableConfigMap(parseableConfigGroup *v1beta1.ParseableConfigGroupSpec) *builder.BuilderConfigMap {

	configMap := &builder.BuilderConfigMap{
		CommonBuilder: builder.CommonBuilder{
			ObjectMeta: metav1.ObjectMeta{
				Name:      parseableConfigGroup.Name,
				Namespace: ib.parseableTenant.GetNamespace(),
				Labels:    ib.commonLabels,
			},
			Client:   ib.client,
			CrObject: ib.parseableTenant,
			OwnerRef: *ib.ownerRef,
		},
		Data: map[string]string{
			"data": fmt.Sprintf("%s", parseableConfigGroup.Data),
		},
	}

	return configMap
}

func (ib *internalBuilder) makeStsOrDeploy(
	ptNode *v1beta1.NodeSpec,
	k8sConfigGroup *v1beta1.K8sConfigGroupSpec,
	storageConfig *[]v1beta1.StorageConfig,
	parseableConfigGroup *v1beta1.ParseableConfigGroupSpec,
	configHash []utils.ConfigMapHash,
) *builder.BuilderDeploymentStatefulSet {

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

	if ib.parseableTenant.Spec.External != (v1beta1.ExternalSpec{}) {
		externalCm := v1.EnvFromSource{
			ConfigMapRef: &v1.ConfigMapEnvSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: ib.parseableTenant.Name + "external-cm",
				},
			},
		}
		envFrom = append(envFrom, externalCm)
	}

	var runner int64 = 1000
	fsPolicy := v1.FSGroupChangeAlways

	podSpec := v1.PodSpec{
		NodeSelector: k8sConfigGroup.NodeSelector,
		Tolerations:  getTolerations(k8sConfigGroup),
		SecurityContext: &v1.PodSecurityContext{
			RunAsUser:           &runner,
			RunAsGroup:          &runner,
			FSGroup:             &runner,
			FSGroupChangePolicy: &fsPolicy,
		},
		Containers: []v1.Container{

			{
				Name:            ptNode.Name + "-" + ptNode.NodeType,
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
				Env: getEnv(*k8sConfigGroup, configHash),
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
		Volumes:            getVolume(k8sConfigGroup, storageConfig, ptNode),
		ServiceAccountName: k8sConfigGroup.ServiceAccountName,
	}

	deployment := builder.BuilderDeploymentStatefulSet{
		CommonBuilder: builder.CommonBuilder{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ptNode.K8sConfigGroup + ptNode.Name,
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
	k8sConfig *v1beta1.K8sConfigGroupSpec,
	nodeSpec *v1beta1.NodeSpec,
) *builder.BuilderStorageConfig {
	return &builder.BuilderStorageConfig{
		CommonBuilder: builder.CommonBuilder{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nodeSpec.Name + "-" + k8sConfig.Name + "-" + sc.Name,
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
	k8sConfig *v1beta1.K8sConfigGroupSpec,
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
		"app":                  "parseable",
		"parseable_cr":         pt.Name,
		"nodeType":             nodeSpec.NodeType,
		"parseableConfigGroup": nodeSpec.ParseableConfigGroupName,
		"k8sConfigGroup":       nodeSpec.K8sConfigGroup,
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

func getVolume(
	k8sConfig *v1beta1.K8sConfigGroupSpec,
	storageConfig *[]v1beta1.StorageConfig,
	nodeSpec *v1beta1.NodeSpec,
) []v1.Volume {
	var volumeHolder = []v1.Volume{}

	for _, sc := range *storageConfig {
		volumeHolder = append(volumeHolder, v1.Volume{
			Name: sc.Name,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: nodeSpec.Name + "-" + k8sConfig.Name + "-" + sc.Name,
				}},
		})
	}

	volumeHolder = append(volumeHolder, k8sConfig.Volumes...)
	return volumeHolder
}

func getEnv(k8sConfigGroup v1beta1.K8sConfigGroupSpec, configHash []utils.ConfigMapHash) []v1.EnvVar {
	var envHolder, hashHolder []v1.EnvVar

	for _, env := range k8sConfigGroup.Env {
		envHolder = append(envHolder, env)
	}
	hashes, _ := utils.MakeConfigMapHash(configHash)

	for _, cmhash := range hashes {
		hashHolder = append(hashHolder, v1.EnvVar{Name: cmhash.Name, Value: cmhash.HashVaule})
	}

	envHolder = append(envHolder, hashHolder...)

	return envHolder
}
