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

# Without Ceph
Take for example,to setup a `IPFS` node
```yaml ipfs.yaml
namespace: ipfs
name: ipfs-daemon
runas: 0
entrypoint: /bin/sh /configmaps/ipfs.sh 
envs:
  IPFS_PATH: /work 
workdir: "/work"
image: docker.io/ipfs/go-ipfs:v0.7.0
cpu: 1
memorymb: 10
ingressmb: 1
egressmb: 1
configfiles:
  - from: ./config
    maptokey: config
  - from: ./ipfs.sh
    maptokey: ipfs.sh
replica: 1 
```
While config are IPFS config, and `ipfs.sh` start script are looks like below.   
```sh ipfs.sh
#!/bin/sh
/usr/local/bin/ipfs daemon --init --init-config /configmaps/config
```

Using 
```bash
kbasectl gen -f ipfs.yaml > ipfs-k8s.yaml
```

Genreate
```yaml
---
apiVersion: v1
kind: Namespace
metadata:
  creationTimestamp: null
  name: ipfs
spec: {}
status: {}
...
---
apiVersion: v1
data:
  config: |
    {
      "API": {
        "HTTPHeaders": {}
      },
      "Addresses": {
        "API": "/ip4/127.0.0.1/tcp/5001",
        "Announce": [],
        "Gateway": "/ip4/127.0.0.1/tcp/8080",
        "NoAnnounce": [],
        "Swarm": [
          "/ip4/0.0.0.0/tcp/4001",
          "/ip6/::/tcp/4001",
          "/ip4/0.0.0.0/udp/4001/quic",
          "/ip6/::/udp/4001/quic"
        ]
      },
      "AutoNAT": {},
      "Bootstrap": [
        "/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
        "/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
        "/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
        "/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
        "/ip4/104.131.131.82/udp/4001/quic/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
        "/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN"
      ],
      "Datastore": {
        "BloomFilterSize": 0,
        "GCPeriod": "1h",
        "HashOnRead": false,
        "Spec": {
          "mounts": [
            {
              "child": {
                "path": "blocks",
                "shardFunc": "/repo/flatfs/shard/v1/next-to-last/2",
                "sync": true,
                "type": "flatfs"
              },
              "mountpoint": "/blocks",
              "prefix": "flatfs.datastore",
              "type": "measure"
            },
            {
              "child": {
                "compression": "none",
                "path": "datastore",
                "type": "levelds"
              },
              "mountpoint": "/",
              "prefix": "leveldb.datastore",
              "type": "measure"
            }
          ],
          "type": "mount"
        },
        "StorageGCWatermark": 90,
        "StorageMax": "100MB"
      },
      "Discovery": {
        "MDNS": {
          "Enabled": true,
          "Interval": 10
        }
      },
      "DontCheckOSXFUSE": false,
      "Experimental": {
        "FilestoreEnabled": true,
        "GraphsyncEnabled": false,
        "Libp2pStreamMounting": true,
        "P2pHttpProxy": false,
        "ShardingEnabled": false,
        "StrategicProviding": false,
        "UrlstoreEnabled": true
      },
      "Gateway": {
        "APICommands": [],
        "HTTPHeaders": {
          "Access-Control-Allow-Headers": [
            "X-Requested-With",
            "Range",
            "User-Agent"
          ],
          "Access-Control-Allow-Methods": [
            "GET"
          ],
          "Access-Control-Allow-Origin": [
            "*"
          ]
        },
        "NoDNSLink": false,
        "NoFetch": false,
        "PathPrefixes": [],
        "PublicGateways": null,
        "RootRedirect": "",
        "Writable": false
      },
      "Identity": {
        "PeerID": "",
        "PrivKey": ""
      },
      "Ipns": {
        "RecordLifetime": "",
        "RepublishPeriod": "",
        "ResolveCacheSize": 128
      },
      "Mounts": {
        "FuseAllowOther": false,
        "IPFS": "~/ipfs",
        "IPNS": "~/ipns"
      },
      "Peering": {
        "Peers": null
      },
      "Plugins": {
        "Plugins": null
      },
      "Provider": {
        "Strategy": ""
      },
      "Pubsub": {
        "DisableSigning": false,
        "Router": ""
      },
      "Reprovider": {
        "Interval": "12h",
        "Strategy": "all"
      },
      "Routing": {
        "Type": "dht"
      },
      "Swarm": {
        "AddrFilters": null,
        "ConnMgr": {
          "GracePeriod": "60s",
          "HighWater": 300,
          "LowWater": 50,
          "Type": "basic"
        },
        "DisableBandwidthMetrics": false,
        "DisableNatPortMap": false,
        "EnableAutoRelay": false,
        "EnableRelayHop": false,
        "Transports": {
          "Multiplexers": {},
          "Network": {
            "TCP": false,
            "Websocket": false
          },
          "Security": {}
        }
      }
    }
  ipfs.sh: |
    #!/bin/bash
    /usr/local/bin/ipfs daemon --init --init-config /configmaps/config
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: ipfs-daemon-configmap
  namespace: ipfs
...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    kbase-pod: ipfs-daemon
  name: ipfs-daemon
  namespace: ipfs
spec:
  replicas: 1
  selector:
    matchLabels:
      kbase-pod: ipfs-daemon
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
  template:
    metadata:
      creationTimestamp: null
      labels:
        kbase-pod: ipfs-daemon
      name: ipfs-daemon
      namespace: ipfs
    spec:
      containers:
      - command:
        - /bin/sh
        - /configmaps/ipfs.sh
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
        - name: IPFS_PATH
          value: /work
        image: docker.io/ipfs/go-ipfs:v0.7.0
        name: ipfs-daemon
        resources:
          requests:
            cpu: "1"
            memory: 10M
        volumeMounts:
        - mountPath: /configmaps
          name: ipfs-daemon-configmap
          readOnly: true
        workingDir: /work
      hostNetwork: true
      securityContext:
        runAsGroup: 0
        runAsUser: 0
      volumes:
      - configMap:
          name: ipfs-daemon-configmap
        name: ipfs-daemon-configmap
status: {}
...

```

# Know Issues
1. PVs can not be modified after creation.  
  Repeatly apply the sampe config gen above will cause kubectl to complain. To workaround, simple add an -m(inmal) flag to check if a resource already exits in k8s.  
  If so, hide it in generated config
2. PV/PVC capacity is currently meaning less
3. Container use host network
