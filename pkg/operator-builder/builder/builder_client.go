package builder

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (b *CommonBuilder) Create(ctx context.Context, buildRecorder BuilderRecorder) (controllerutil.OperationResult, error) {

	if err := b.Client.Create(ctx, b.DesiredState); err != nil {
		buildRecorder.createEvent(b.CrObject, b.DesiredState, err)
		return "", err
	} else {
		buildRecorder.createEvent(b.CrObject, b.DesiredState, nil)
		return controllerutil.OperationResultCreated, nil
	}
}

func (b *CommonBuilder) Update(ctx context.Context, buildRecorder BuilderRecorder) (controllerutil.OperationResult, error) {
	if err := b.Client.Update(ctx, b.DesiredState); err != nil {
		buildRecorder.updateEvent(b.CrObject, b.DesiredState, err)
		return "", err
	} else {
		buildRecorder.updateEvent(b.CrObject, b.DesiredState, nil)
		return controllerutil.OperationResultUpdated, nil
	}
}

func (b CommonBuilder) List(ctx context.Context) (client.ObjectList, error) {
	listOpts := []client.ListOption{
		client.InNamespace(b.ObjectMeta.Namespace),
		client.MatchingLabels(b.Labels),
	}

	deployment := b.ObjectList
	if err := b.Client.List(ctx, deployment, listOpts...); err != nil {
		return nil, err
	} else {
		return deployment, nil
	}

}
