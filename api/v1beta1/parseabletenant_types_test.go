/*
 * Parseable Server (C) 2022 - 2023 Parseable, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

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
      parseableConfig: parseableserver
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
      service: {}
    storageConfig:
    - name: pvctest
      mountPath: "/var/lib"
      pvcSpec:
        accessModes:
        - ReadWriteOnce
        storageClassName: "standard"
        resources:
          requests:
            storage: 10Gi
  parseableConfig:
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
