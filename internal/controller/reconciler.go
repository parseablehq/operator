package controller

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/parseablehq/parseable-operator/api/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	nodeTypeNodeSpec := getAllNodeSpecForNodeType(pt)

	cmBuilder := NewBuilderObject(pt, metav1.ObjectMeta{Name: pt.GetName() + "cm", Namespace: pt.GetNamespace()}, map[string]string{"data": pt.Spec.External.ObjectStore.Spec.Data})
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
	for range nodeTypeNodeSpec {

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

type BuilderState struct {
	Client       client.Client
	DesiredState client.Object
	CurrentState client.Object
	CrObject     client.Object
	OwnerRef     metav1.OwnerReference
}

func (b BuilderState) CreateOrUpdate() (controllerutil.OperationResult, error) {
	addOwnerRefToObject(b.DesiredState, b.OwnerRef)
	addHashToObject(b.DesiredState, b.OwnerRef.Kind+"OperatorHash")
	if err := b.Client.Get(context.TODO(), types.NamespacedName{Name: b.DesiredState.GetName(), Namespace: b.DesiredState.GetNamespace()}, b.CurrentState); err != nil {
		if apierrors.IsNotFound(err) {
			result, err := b.Create(context.TODO())
			if err != nil {
				return controllerutil.OperationResultNone, err
			}
			return result, nil
		} else {
			fmt.Println("Delete")
			return "", err
		}
	} else {

		fmt.Println("Update")
	}
	return controllerutil.OperationResultNone, nil
}

func (b BuilderState) Create(ctx context.Context) (controllerutil.OperationResult, error) {
	if err := b.Client.Create(ctx, b.DesiredState); err != nil {
		return "", err
	} else {
		return controllerutil.OperationResultCreated, nil
	}
}

type BuilderObject struct {
	CrObject   client.Object
	ObjectMeta metav1.ObjectMeta
	Data       map[string]string
}

func NewBuilderObject(
	crObject client.Object,
	objectMeta metav1.ObjectMeta,
	data map[string]string,
) *BuilderObject {
	return &BuilderObject{
		CrObject:   crObject,
		ObjectMeta: objectMeta,
		Data:       data,
	}
}

func (b BuilderObject) MakeConfigMap() (*v1.ConfigMap, error) {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: b.ObjectMeta,
		Data:       b.Data,
	}, nil
}

// addOwnerRefToObject appends the desired OwnerReference to the object
func addOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	trueVar := true
	ownerRef = metav1.OwnerReference{
		APIVersion: ownerRef.APIVersion,
		Kind:       ownerRef.Kind,
		Name:       ownerRef.Name,
		UID:        ownerRef.UID,
		Controller: &trueVar,
	}
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

func addHashToObject(obj client.Object, name string) error {
	if sha, err := getObjectHash(obj); err != nil {
		return err
	} else {
		annotations := obj.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
			obj.SetAnnotations(annotations)
		}
		annotations[name] = sha
		return nil
	}
}

func getObjectHash(obj client.Object) (string, error) {
	if bytes, err := json.Marshal(obj); err != nil {
		return "", err
	} else {
		sha1Bytes := sha1.Sum(bytes)
		return base64.StdEncoding.EncodeToString(sha1Bytes[:]), nil
	}
}