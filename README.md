# k8builder-demo-nolabel
1. Scaffold
```
kubebuilder init --domain demo.jonlimpw.io --repo operator-demo
```
2. Create API
```
kubebuilder create api --group demo --version v1 --kind DeploymentLabelCheck
```
3. Define API
- Check that both labels and annotations for 1 Deployment called flaskpython-deployment in the python namespace contains the required label for library injection to happen. 
- ObjectMeta.spec.template.metadata.labels = admission.datadoghq.com/enabled: "true"
- ObjectMeta.spec.template.metadata.annotations = admission.datadoghq.com/python-lib.version: v2.6.5
   - Might have to find a regex pattern for python-lib to account for other langages.
```
spec:  
  template:
    metadata:
      creationTimestamp: null
      labels:
        admission.datadoghq.com/enabled: "true"
        app: adservice
        tags.datadoghq.com/env: XXX
        tags.datadoghq.com/service: XXX
        tags.datadoghq.com/version: XXX
      annotations:
        admission.datadoghq.com/python-lib.version: v2.6.5
```
4. Ensure proper RBAC
```YAML
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: deployment-reader-role
rules:
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "update", "watch", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: deployment-reader-binding
subjects:
- kind: ServiceAccount
  name: default
  namespace: kubebuilder
roleRef:
  kind: ClusterRole
  name: deployment-reader-role
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: deploymentlabelcheck-reader-role
rules:
- apiGroups: ["demo.demo.jonlimpw.io"]
  resources: ["deploymentlabelchecks"]
  verbs: ["get", "list", "update", "watch", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: deploymentlabelcheck-reader-binding
subjects:
- kind: ServiceAccount
  name: default
  namespace: kubebuilder
roleRef:
  kind: ClusterRole
  name: deploymentlabelcheck-reader-role
  apiGroup: rbac.authorization.k8s.io

```
