# Gimlet manifest examples

On this page you can find Gimlet manifest files achieving various goals

Template each file with the following command:

```
gimlet manifest template \
  -f xxx/yyy.yaml \
  --vars demo.env
```

Like 

```
gimlet manifest template \
  -f simple/simple.yaml \
  --vars demo.env
```

prints

```
---
# Source: onechart/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: myapp
  namespace: my-team
  labels:
    helm.sh/chart: onechart-0.32.0
    app.kubernetes.io/name: onechart
    app.kubernetes.io/instance: myapp
    app.kubernetes.io/managed-by: Helm
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: http
      protocol: TCP
      name: http
    
  selector:
    app.kubernetes.io/name: onechart
    app.kubernetes.io/instance: myapp
---
# Source: onechart/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: my-team
  labels:
    helm.sh/chart: onechart-0.32.0
    app.kubernetes.io/name: onechart
    app.kubernetes.io/instance: myapp
    app.kubernetes.io/managed-by: Helm
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: onechart
      app.kubernetes.io/instance: myapp
  template:
    metadata:
      annotations:
        checksum/config: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
      labels:
        app.kubernetes.io/name: onechart
        app.kubernetes.io/instance: myapp
    spec:
      securityContext:
        fsGroup: 999
      containers:
        - name: myapp
          securityContext: &securityContext
            {}
          image: "myapp:1.1.0"
          imagePullPolicy: IfNotPresent
          ports:
            - name: http
              containerPort: 80
              protocol: TCP
          resources:
            limits:
              cpu: 200m
              memory: 200Mi
            requests:
              cpu: 200m
              memory: 200Mi
```

Bootstrap gitops automation CLI with https://github.com/gimlet-io/gimlet-cli#installation
