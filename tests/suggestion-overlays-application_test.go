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

func writeSuggestionOverlaysApplication(th *KustTestHarness) {
	th.writeF("/manifests/katib-v1alpha2/suggestion/overlays/application/application.yaml", `
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: $(generateName)
spec:
  selector:
    matchLabels:
      app.kubernetes.io/instance: $(generateName)
  componentKinds:
  - group: core
    kind: Service
  - group: apps
    kind: Deployment
  descriptor:
    type: "suggestion"
    version: "v1alpha2"
    description: "Katib suggestion provides various algorithms to AutoML jobs."
    maintainers:
    - name: Zhongxuan Wu
      email: wuzhongxuan@caicloud.io
    - name: Ce Gao
      email: gaoce@caicloud.io
    - name: Johnu George
      email: johnugeo@cisco.com
    - name: Hougang Liu
      email: liuhougang6@126.com
    - name: Richard Liu
      email: ricliu@google.com
    - name: YujiOshima
      email: yuji.oshima0x3fd@gmail.com
    owners:
    - name: Ce Gao
      email: gaoce@caicloud.io
    - name: Johnu George
      email: johnugeo@cisco.com
    - name: Hougang Liu
      email: liuhougang6@126.com
    - name: Richard Liu
      email: ricliu@google.com
    - name: YujiOshima
      email: yuji.oshima0x3fd@gmail.com
    keywords:
    - katib
    - suggestion
    - hyperparameter tuning
    links:
    - description: About
      url: "https://github.com/kubeflow/katib"
  addOwnerRef: true
`)
	th.writeF("/manifests/katib-v1alpha2/suggestion/overlays/application/params.yaml", `
varReference:
- path: metadata/name
  kind: Application
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Application
- path: spec/selector/app.kubernetes.io\/instance
  kind: Service
- path: spec/selector/matchLabels/app.kubernetes.io\/instance
  kind: Deployment
- path: spec/template/metadata/labels/app.kubernetes.io\/instance
  kind: Deployment
`)
	th.writeF("/manifests/katib-v1alpha2/suggestion/overlays/application/params.env", `
generateName=
`)
	th.writeK("/manifests/katib-v1alpha2/suggestion/overlays/application", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- ../../base
resources:
- application.yaml
configMapGenerator:
- name: suggestion-parameters
  env: params.env
vars:
- name: generateName
  objref:
    kind: ConfigMap
    name: suggestion-parameters
    apiVersion: v1
  fieldref:
    fieldpath: data.generateName
configurations:
- params.yaml
commonLabels:
  app.kubernetes.io/name: suggestion
  app.kubernetes.io/instance: $(generateName)
  app.kubernetes.io/managed-by: kfctl
  app.kubernetes.io/component: katib
  app.kubernetes.io/part-of: kubeflow
  app.kubernetes.io/version: v0.6
`)
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-bayesianoptimization-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-suggestion-bayesianoptimization
  labels:
    app: katib
    component: suggestion-bayesianoptimization
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: suggestion-bayesianoptimization
  template:
    metadata:
      name: katib-suggestion-bayesianoptimization
      labels:
        app: katib
        component: suggestion-bayesianoptimization
    spec:
      containers:
      - name: katib-suggestion-bayesianoptimization
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-bayesianoptimization:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        ports:
        - name: api
          containerPort: 6789
`)
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-bayesianoptimization-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-suggestion-bayesianoptimization
  labels:
    app: katib
    component: suggestion-bayesianoptimization
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    app: katib
    component: suggestion-bayesianoptimization
`)
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-grid-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-suggestion-grid
  labels:
    app: katib
    component: suggestion-grid
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: suggestion-grid
  template:
    metadata:
      name: katib-suggestion-grid
      labels:
        app: katib
        component: suggestion-grid
    spec:
      containers:
      - name: katib-suggestion-grid
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-grid:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        ports:
        - name: api
          containerPort: 6789
`)
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-grid-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-suggestion-grid
  labels:
    app: katib
    component: suggestion-grid
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    app: katib
    component: suggestion-grid
`)
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-hyperband-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-suggestion-hyperband
  labels:
    app: katib
    component: suggestion-hyperband
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: suggestion-hyperband
  template:
    metadata:
      name: katib-suggestion-hyperband
      labels:
        app: katib
        component: suggestion-hyperband
    spec:
      containers:
      - name: katib-suggestion-hyperband
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-hyperband:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        ports:
        - name: api
          containerPort: 6789
`)
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-hyperband-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-suggestion-hyperband
  labels:
    app: katib
    component: suggestion-hyperband
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    app: katib
    component: suggestion-hyperband
`)
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-nasrl-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-suggestion-nasrl
  labels:
    app: katib
    component: suggestion-nasrl
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: suggestion-nasrl
  template:
    metadata:
      name: katib-suggestion-nasrl
      labels:
        app: katib
        component: suggestion-nasrl
    spec:
      containers:
      - name: katib-suggestion-nasrl
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-nasrl:v0.1.2-alpha-289-g14dad8b
        ports:
        - name: api
          containerPort: 6789
`)
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-nasrl-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-suggestion-nasrl
  labels:
    app: katib
    component: suggestion-nasrl
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    app: katib
    component: suggestion-nasrl
`)
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-random-deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: katib-suggestion-random
  labels:
    app: katib
    component: suggestion-random
spec:
  replicas: 1
  selector:
    matchLabels:
      app: katib
      component: suggestion-random
  template:
    metadata:
      name: katib-suggestion-random
      labels:
        app: katib
        component: suggestion-random
    spec:
      containers:
      - name: katib-suggestion-random
        image: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-random:v0.1.2-alpha-289-g14dad8b
        imagePullPolicy: IfNotPresent
        ports:
        - name: api
          containerPort: 6789
`)
	th.writeF("/manifests/katib-v1alpha2/suggestion/base/suggestion-random-service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: katib-suggestion-random
  labels:
    app: katib
    component: suggestion-random
spec:
  type: ClusterIP
  ports:
    - port: 6789
      protocol: TCP
      name: api
  selector:
    app: katib
    component: suggestion-random
`)
	th.writeK("/manifests/katib-v1alpha2/suggestion/base", `
namespace: kubeflow
resources:
- suggestion-bayesianoptimization-deployment.yaml
- suggestion-bayesianoptimization-service.yaml
- suggestion-grid-deployment.yaml
- suggestion-grid-service.yaml
- suggestion-hyperband-deployment.yaml
- suggestion-hyperband-service.yaml
- suggestion-nasrl-deployment.yaml
- suggestion-nasrl-service.yaml
- suggestion-random-deployment.yaml
- suggestion-random-service.yaml
generatorOptions:
  disableNameSuffixHash: true
images:
- name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-hyperband
  newTag: v0.6.0-rc.0
  newName: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-hyperband
- name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-bayesianoptimization
  newTag: v0.6.0-rc.0
  newName: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-bayesianoptimization
- name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-grid
  newTag: v0.6.0-rc.0
  newName: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-grid
- name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-random
  newTag: v0.6.0-rc.0
  newName: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-random
- name: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-nasrl
  newTag: v0.6.0-rc.0
  newName: gcr.io/kubeflow-images-public/katib/v1alpha2/suggestion-nasrl
`)
}

func TestSuggestionOverlaysApplication(t *testing.T) {
	th := NewKustTestHarness(t, "/manifests/katib-v1alpha2/suggestion/overlays/application")
	writeSuggestionOverlaysApplication(th)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	expected, err := m.AsYaml()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	targetPath := "../katib-v1alpha2/suggestion/overlays/application"
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
