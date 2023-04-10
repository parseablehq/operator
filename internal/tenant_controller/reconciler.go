package parseabletenantcontroller

import (
	"context"

	v1 "k8s.io/api/core/v1"

	"github.com/datainfrahq/operator-builder/builder"
	"github.com/go-logr/logr"
	"github.com/parseablehq/parseable-operator/api/v1beta1"
)

func (r *ParseableTenantReconciler) do(ctx context.Context, pt *v1beta1.ParseableTenant, log logr.Logger) error {

	// create ownerRef passed to each object created
	getOwnerRef := makeOwnerRef(
		pt.APIVersion,
		pt.Kind,
		pt.Name,
		pt.UID,
	)

	var ib *internalBuilder

	nodeSpecs := getAllNodeSpecForNodeType(pt)

	parseableConfigMap := []builder.BuilderConfigMap{}
	parseableConfigMapHash := []builder.BuilderConfigMapHash{}
	parseableDeploymentOrStatefulset := []builder.BuilderDeploymentStatefulSet{}
	parseableStorage := []builder.BuilderStorageConfig{}
	parseableService := builder.BuilderService{}

	// For all the nodeSpec ie nodeType to nodeSpec
	// Get all the config group defined and append to configMap builder
	// For each config group defined create a configmap hash and append to configmaphash builder
	// Get all the k8s config group defined and append to deploymentstatefulset builder
	// For all the storage config defined in k8s config group append
	for _, nodeSpec := range nodeSpecs {
		ib = newInternalBuilder(pt, r.Client, &nodeSpec.NodeSpec, getOwnerRef)
		for _, parseableConfig := range pt.Spec.ParseableConfigGroup {

			if nodeSpec.NodeSpec.ParseableConfigGroupName == parseableConfig.Name {
				cm := *ib.makeParseableConfigMap(&parseableConfig)
				parseableConfigMap = append(parseableConfigMap, cm)
				parseableConfigMapHash = append(parseableConfigMapHash, builder.BuilderConfigMapHash{Object: &v1.ConfigMap{Data: cm.Data, ObjectMeta: cm.ObjectMeta}})
				for _, k8sConfig := range pt.Spec.K8sConfigGroup {
					if nodeSpec.NodeSpec.K8sConfigGroup == k8sConfig.Name {
						parseableDeploymentOrStatefulset = append(parseableDeploymentOrStatefulset, *ib.makeStsOrDeploy(&nodeSpec.NodeSpec, &k8sConfig, &k8sConfig.StorageConfig, &parseableConfig))
						parseableService = *ib.makeService(&k8sConfig, nodeSpec.NodeType)
						for _, sc := range k8sConfig.StorageConfig {
							parseableStorage = append(parseableStorage, *ib.makePvc(&sc, &k8sConfig))
						}
					}
				}
			}
		}
	}

	// append external config and hash to configmap builder
	if pt.Spec.External != (v1beta1.ExternalSpec{}) {
		cm := *ib.makeExternalConfigMap()
		parseableConfigMap = append(parseableConfigMap, cm)
		parseableConfigMapHash = append(parseableConfigMapHash, builder.BuilderConfigMapHash{Object: cm.DesiredState})
	}

	// construct builder
	builder := builder.NewBuilder(
		builder.ToNewBuilderConfigMap(parseableConfigMap),
		builder.ToNewBuilderConfigMapHash(parseableConfigMapHash),
		builder.ToNewBuilderDeploymentStatefulSet(parseableDeploymentOrStatefulset),
		builder.ToNewBuilderStorageConfig(parseableStorage),
		builder.ToNewBuilderRecorder(builder.BuilderRecorder{Recorder: r.Recorder, ControllerName: "ParseableOperator"}),
		builder.ToNewBuilderContext(builder.BuilderContext{Context: ctx}),
		builder.ToNewBuilderService(parseableService),
		builder.ToNewBuilderStore(*builder.NewStore(ib.client, ib.commonLabels, pt.Namespace, pt)),
	)

	// All builder methods called are responsible for reconciling
	// and triggering reconcilers in case of state change.

	// reconcile configmap
	_, err := builder.ReconcileConfigMap()
	if err != nil {
		return err
	}

	// reconcile configmap hash
	cmhashes, err := builder.ReconcileConfigMapHash()
	if err != nil {
		return err
	}

	// reconcile svc
	_, err = builder.ReconcileService()
	if err != nil {
		return err
	}

	// reconcile depoyment or statefulset
	_, err = builder.ReconcileDeployOrSts(cmhashes)
	if err != nil {
		return err
	}

	// reconcile storage
	_, err = builder.ReconcileStorage()
	if err != nil {
		return err
	}

	// reconcile store
	if err := builder.ReconcileStore(); err != nil {
		return err
	}

	return nil
}
