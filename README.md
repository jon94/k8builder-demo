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