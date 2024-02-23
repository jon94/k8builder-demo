# k8builder-demo
1. Scaffold
```
kubebuilder init --domain demo.jonlimpw.io --repo operator-demo
```
2. Create API
```
kubebuilder create api --group demo --version v1 --kind DeploymentLabelCheck
```