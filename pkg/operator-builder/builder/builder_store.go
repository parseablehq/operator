package builder

type K8sObjectName string

const (
	// As per k8s naming
	configMap  K8sObjectName = "ConfigMap"
	deployment K8sObjectName = "Deployment"
	pvc        K8sObjectName = "PersistentVolumeClaim"
	svc        K8sObjectName = "Service"
)

type InternalStore struct {
	ObjectNameKind map[string]string
}

func NewStore() *InternalStore {
	return &InternalStore{
		ObjectNameKind: make(map[string]string),
	}
}

func ToNewBuilderStore(builder InternalStore) func(*Builder) {
	return func(s *Builder) {
		s.Store = builder
	}
}

func (s *Builder) Put(key, value string) {
	if _, isKeyExists := s.Store.ObjectNameKind[key]; isKeyExists {
		return
	} else {
		s.Store.ObjectNameKind[key] = value
	}
}
