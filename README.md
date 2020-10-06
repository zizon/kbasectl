# kbasectl
Kubernetes config generator for ceph base deployment. 

# Prerequirement
- A running `CephFS` Cluster(optional if Pod/Deploymnet not mouting a cephfs), with a moutable auth token
- A running `Kubernets` Cluster, with an `ServiceAccount` bindto `cluster-admin` `ClusterRole` (not strictly required.)
  - ReadWrite of Namespace (for creation of seperated namespace of each deployment)
  - ReadWrite of PersistenVolume (for pv creation of each deployment namesapce)
  - ReadWrite of PersistenVolumeClaim (for pvc creation of each deployment namesapce)
  - ReadWrite of Secrets (ceph token creation if not exists,can be revoke if surealy exists)
  - ReadWrite of ConfigMaps (config maps associated with each deployment)
  
# Setup
init a Config file located at ~/.kbasectl/config.yaml
```yaml
rest:
  bearertoken: $(service-account-token)
  host: $(k8s-api-server)
  tlsclientconfig:
    insecure: true 
ceph:
  monitors: 
    - $(cpeh-monitor-1-ip)
    - $(ceph-monitor-2-ip)
  user: $(ceph-auth-token-owner)
  token: $(ceph-token)
```
- $(service-account-token) token of service account that has ability describel above
- $(k8s-api-server) Kubernetes API Server
- $(cpeh-monitor-n-ip) Ceph Monitor nodes. Better to have an DNS name, Since Ceph Cluster can change its monitor groups.  
In case of all `old` monitors are removed, Ceph PersistenVolumes assoicalted with this monitors maybe dyfunctional.  
With DNS resolvalbe name help overcome this situations.
- $(ceph-auth-token-owner) users that initial mount for each ceph container volume.(default to admin)
- $(ceph-token) tokens that use to authenticate $(ceph-auth-token-owner).(default ot admin-token)

# Example
Said, 
1. you want to create a deployment in namesapce `test-namespace`,assining name `test-deploy` for reference.  
2. user docker.io/golang:v15.2 as single container image in pod of this deployment.  
  - bring ins env: `hello`:`kitty` as contaienr exexcution enviroment variable
  - /work as working dir and /work/entrypoint.sh as startup script
  - resource requiremnts are:
    - CPU 1
    - Memory 100m
    - Network Ingress/Outgress: 60/50 mbps
  - runas root(uid=0)
3. mounting ceph path /cephfs/template/some-binary to contaienr path /work/bin/some-binary unmodifiable.
4. also you may want to map local config file `./some-config.xml` to be mounted to `/configmaps/default.xml`
5. and may be label this deployment with label just:test
6. finnlay run 5 replca of this kind of pod  

Given config files bellow can satisfied above.
```yaml template.yaml
namespace: test-namespace
name: test-deploy
labels:
  just: test
envs:
  hello: kitty
runas: 0
entrypoint: "/bin/bash /work/entrypoint.sh"
workdir: "/work"
image: docker.io/golang:v15.2
cpu: 1
memorymb: 100
ingressmb: 60
egressmb: 50
configfiles:
- from: ./some-config.xml
  maptokey: default.xml
replica: 5
cephbind:
- from: /cephfs/template
  to: /work/bin/
  readonly: true
  capacitymb: 65536
  filter:
  - some-binary
```

With simple command
```bash
./kbasectl gen -f template.yaml
```

Will Generate Kubernetes/kubectp apply usable configs
```yaml
---
apiVersion: v1
kind: Namespace
metadata:
  creationTimestamp: null
  name: test-namespace
spec: {}
status: {}
...
---
apiVersion: v1
kind: Namespace
metadata:
  creationTimestamp: null
  name: kbase
spec: {}
status: {}
...
---
apiVersion: v1
kind: Secret
metadata:
  creationTimestamp: null
  name: ceph-kbase-token
  namespace: kbase
stringData:
  $(ceph-auth-token-owner): $(ceph-token)
...
---
apiVersion: v1
data:
  default.xml: |
    <example-config/>
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: test-deploy-configmap
  namespace: test-namespace
...
---
apiVersion: v1
kind: PersistentVolume
metadata:
  creationTimestamp: null
  name: ceph-volume-read--cephfs-template
spec:
  accessModes:
  - ReadWriteMany
  - ReadOnlyMany
  capacity:
    storage: 65536M
  cephfs:
    monitors:
    - $(ceph-monitor-1-ip)
    - $(ceph-monitor-2-ip)
    path: /cephfs/template
    secretRef:
      name: ceph-kbase-token
      namespace: kbase
    user: admin
  persistentVolumeReclaimPolicy: Retain
status: {}
...
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  creationTimestamp: null
  name: ceph-volume-read--cephfs-template
  namespace: test-namespace
spec:
  accessModes:
  - ReadOnlyMany
  resources:
    requests:
      storage: 65536M
  volumeName: ceph-volume-read--cephfs-template
status: {}
...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    just: test
    kbase-pod: test-deploy
  name: test-deploy
  namespace: test-namespace
spec:
  replicas: 5
  selector:
    matchLabels:
      just: test
      kbase-pod: test-deploy
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
  template:
    metadata:
      creationTimestamp: null
      labels:
        just: test
        kbase-pod: test-deploy
      name: test-deploy
      namespace: test-namespace
    spec:
      containers:
      - command:
        - /bin/bash /work/entrypoint.sh
        env:
        - name: CONTAINER_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: CONTAINER_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: CONTAINER_HOST_IP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        - name: CONTAINER_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: hello
          value: kitty
        image: docker.io/golang:v15.2
        name: test-deploy
        resources:
          requests:
            cpu: "1"
            memory: 100M
        volumeMounts:
        - mountPath: /work/bin/some-binary
          name: ceph-volume-read--cephfs-template
          readOnly: true
          subPathExpr: some-binary
        - mountPath: /configmaps
          name: test-deploy-configmap
          readOnly: true
        workingDir: /work
      hostNetwork: true
      securityContext:
        runAsGroup: 0
        runAsUser: 0
      volumes:
      - name: ceph-volume-read--cephfs-template
        persistentVolumeClaim:
          claimName: ceph-volume-read--cephfs-template
          readOnly: true
      - configMap:
          name: test-deploy-configmap
        name: test-deploy-configmap
status: {}
...
```

# Know Issues
1. PVs can not be modified after creation.  
  Repeatly apply the sampe config gen above will cause kubectl to complain. To workaround, simple add an -m(inmal) flag to check if a resource already exits in k8s.  
  If so, hide it in generated config
2. PV/PVC capacity is currently meaning less
