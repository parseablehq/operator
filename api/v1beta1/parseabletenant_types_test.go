package v1beta1

import (
	"testing"

	"sigs.k8s.io/yaml"
)

var CR = `
apiVersion: parseable.io/v1beta1
kind: ParseableTenant
spec:
  nodes:
    - name: parseable-server
      kind: Statefulset
      replicas: 1
      nodeType: server
      k8sConfigGroup: parseableserver
      parseableConfigGroup: parseableserver
  deploymentOrder:
  - server
  external:
    objectStore:
      spec:
        type: s3
        data: |-
          s3.url=http://minio.minio.svc.cluster.local:9000
          s3.access.key=minioadmin
          s3 .secret.key=minioadmin
          s3.region=us-east-1
          s3.bucket=parseable
  k8sConfigGroup:
  - name: parseable-server
    spec:
      serviceAccountName: "parsebale-server"
      nodeSelector: {}
      toleration: {}
      affinity: {}
      labels: {}
  parseableConfigGroup:
  - name: parseable-server
    args: "s3-store"
    data: |-
      addr=0.0.0.0:8000 
      staging.dir=./staging
      fs.dir=./data
      username=admin
      password=admin  
`

func TestParseableTenant(t *testing.T) {
	var spec ParseableTenant

	t.Logf("%+v", spec.Spec.Nodes)
	err := yaml.Unmarshal([]byte(CR), &spec)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", spec.Spec)
}
