package tests_test

import (
	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/v3/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/loader"
	"sigs.k8s.io/kustomize/v3/pkg/plugins"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/target"
	"sigs.k8s.io/kustomize/v3/pkg/validators"
	"testing"
)

func writeArgoOverlaysIstio(th *KustTestHarness) {
	th.writeF("/manifests/argo/overlays/istio/virtual-service.yaml", `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: argo-ui
spec:
  gateways:
  - kubeflow-gateway
  hosts:
  - '*'
  http:
  - match:
    - uri:
        prefix: /argo/
    rewrite:
      uri: /
    route:
    - destination:
        host: argo-ui.$(namespace).svc.$(clusterDomain)
        port:
          number: 80
`)
	th.writeF("/manifests/argo/overlays/istio/params.yaml", `
varReference:
- path: spec/http/route/destination/host
  kind: VirtualService
`)
	th.writeK("/manifests/argo/overlays/istio", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- virtual-service.yaml
configurations:
- params.yaml
`)
	th.writeF("/manifests/argo/base/cluster-role-binding.yaml", `
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    app: argo
  name: argo
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: argo
subjects:
- kind: ServiceAccount
  name: argo
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    app: argo-ui
  name: argo-ui
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: argo-ui
subjects:
- kind: ServiceAccount
  name: argo-ui
`)
	th.writeF("/manifests/argo/base/cluster-role.yaml", `
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: argo
  name: argo
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - pods/exec
  verbs:
  - create
  - get
  - list
  - watch
  - update
  - patch
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - watch
  - list
- apiGroups:
  - ""
  resources:
  - persistentvolumeclaims
  verbs:
  - create
  - delete
- apiGroups:
  - argoproj.io
  resources:
  - workflows
  verbs:
  - get
  - list
  - watch
  - update
  - patch
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: argo
  name: argo-ui
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - pods/exec
  - pods/log
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
- apiGroups:
  - argoproj.io
  resources:
  - workflows
  verbs:
  - get
  - list
  - watch
`)
	th.writeF("/manifests/argo/base/config-map.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: workflow-controller-configmap
data:
  config: |
    {
    executorImage: $(executorImage),
    containerRuntimeExecutor: $(containerRuntimeExecutor),
    artifactRepository:
    {
        s3: {
            bucket: $(artifactRepositoryBucket),
            keyPrefix: $(artifactRepositoryKeyPrefix),
            endpoint: $(artifactRepositoryEndpoint),
            insecure: $(artifactRepositoryInsecure),
            accessKeySecret: {
                name: $(artifactRepositoryAccessKeySecretName),
                key: $(artifactRepositoryAccessKeySecretKey)
            },
            secretKeySecret: {
                name: $(artifactRepositorySecretKeySecretName),
                key: $(artifactRepositorySecretKeySecretKey)
            }
        }
    }
    }
`)
	th.writeF("/manifests/argo/base/crd.yaml", `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: workflows.argoproj.io
spec:
  group: argoproj.io
  names:
    kind: Workflow
    listKind: WorkflowList
    plural: workflows
    shortNames:
    - wf
    singular: workflow
  scope: Namespaced
  version: v1alpha1
`)
	th.writeF("/manifests/argo/base/deployment.yaml", `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: argo-ui
  name: argo-ui
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: argo-ui
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: argo-ui
    spec:
      containers:
      - env:
        - name: ARGO_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: IN_CLUSTER
          value: "true"
        - name: BASE_HREF
          value: /argo/
        image: argoproj/argoui:v2.3.0
        imagePullPolicy: IfNotPresent
        name: argo-ui
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        readinessProbe:
          httpGet:
            path: /
            port: 8001
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: argo-ui
      serviceAccountName: argo-ui
      terminationGracePeriodSeconds: 30
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: workflow-controller
  name: workflow-controller
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: workflow-controller
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: workflow-controller
    spec:
      containers:
      - args:
        - --configmap
        - workflow-controller-configmap
        command:
        - workflow-controller
        env:
        - name: ARGO_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        image: argoproj/workflow-controller:v2.3.0
        imagePullPolicy: IfNotPresent
        name: workflow-controller
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: argo
      serviceAccountName: argo
      terminationGracePeriodSeconds: 30
`)
	th.writeF("/manifests/argo/base/service-account.yaml", `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: argo
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: argo-ui
`)
	th.writeF("/manifests/argo/base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  annotations:
    getambassador.io/config: |-
      ---
      apiVersion: ambassador/v0
      kind:  Mapping
      name: argo-ui-mapping
      prefix: /argo/
      service: argo-ui.$(namespace)
  labels:
    app: argo-ui
  name: argo-ui
spec:
  ports:
  - port: 80
    targetPort: 8001
  selector:
    app: argo-ui
  sessionAffinity: None
  type: NodePort
`)
	th.writeF("/manifests/argo/base/params.yaml", `
varReference:
- path: data/config
  kind: ConfigMap
- path: data/config
  kind: Deployment
- path: metadata/annotations/getambassador.io\/config
  kind: Service
`)
	th.writeF("/manifests/argo/base/params.env", `
namespace=
executorImage=argoproj/argoexec:v2.3.0
containerRuntimeExecutor=docker
artifactRepositoryBucket=mlpipeline
artifactRepositoryKeyPrefix=artifacts
artifactRepositoryEndpoint=minio-service.kubeflow:9000
artifactRepositoryInsecure=true
artifactRepositoryAccessKeySecretName=mlpipeline-minio-artifact
artifactRepositoryAccessKeySecretKey=accesskey
artifactRepositorySecretKeySecretName=mlpipeline-minio-artifact
artifactRepositorySecretKeySecretKey=secretkey
clusterDomain=cluster.local
`)
	th.writeK("/manifests/argo/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- cluster-role-binding.yaml
- cluster-role.yaml
- config-map.yaml
- crd.yaml
- deployment.yaml
- service-account.yaml
- service.yaml
namespace: kubeflow
commonLabels:
  kustomize.component: argo
images:
- name: argoproj/argoui
  newName: argoproj/argoui
  newTag: v2.3.0
- name: argoproj/workflow-controller
  newName: argoproj/workflow-controller
  newTag: v2.3.0
configMapGenerator:
- name: workflow-controller-parameters
  env: params.env
generatorOptions:
  disableNameSuffixHash: true
vars:
- name: executorImage
  objref:
    kind: ConfigMap
    name: workflow-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.executorImage
- name: containerRuntimeExecutor
  objref:
    kind: ConfigMap
    name: workflow-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.containerRuntimeExecutor
- name: artifactRepositoryBucket
  objref:
    kind: ConfigMap
    name: workflow-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.artifactRepositoryBucket
- name: artifactRepositoryKeyPrefix
  objref:
    kind: ConfigMap
    name: workflow-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.artifactRepositoryKeyPrefix
- name: artifactRepositoryEndpoint
  objref:
    kind: ConfigMap
    name: workflow-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.artifactRepositoryEndpoint
- name: artifactRepositoryInsecure
  objref:
    kind: ConfigMap
    name: workflow-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.artifactRepositoryInsecure
- name: artifactRepositoryAccessKeySecretName
  objref:
    kind: ConfigMap
    name: workflow-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.artifactRepositoryAccessKeySecretName
- name: artifactRepositoryAccessKeySecretKey
  objref:
    kind: ConfigMap
    name: workflow-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.artifactRepositoryAccessKeySecretKey
- name: artifactRepositorySecretKeySecretName
  objref:
    kind: ConfigMap
    name: workflow-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.artifactRepositorySecretKeySecretName
- name: artifactRepositorySecretKeySecretKey
  objref:
    kind: ConfigMap
    name: workflow-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.artifactRepositorySecretKeySecretKey
- name: namespace
  objref:
    kind: ConfigMap
    name: workflow-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.namespace
- name: clusterDomain
  objref:
    kind: ConfigMap
    name: workflow-controller-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.clusterDomain
configurations:
- params.yaml
`)
}

func TestArgoOverlaysIstio(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/argo/overlays/istio")
	writeArgoOverlaysIstio(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../argo/overlays/istio"
	fsys := fs.MakeRealFS()
	lrc := loader.RestrictionRootOnly
	_loader, loaderErr := loader.NewLoader(lrc, validators.MakeFakeValidator(), targetPath, fsys)
	if loaderErr != nil {
		t.Fatalf("could not load kustomize loader: %v", loaderErr)
	}
	rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())
	pc := plugins.DefaultPluginConfig()
	kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl(), plugins.NewLoader(pc, rf))
	if err != nil {
		th.t.Fatalf("Unexpected construction error %v", err)
	}
	actual, err := kt.MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(actual, string(expected))
}
