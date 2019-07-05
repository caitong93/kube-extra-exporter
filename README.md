
# Project kube-extra-exporter

<!-- Write one paragraph of this project description here -->
kube-extra-exporter exports tcp connection usages stats. 

```
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp",tcp_state="close"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp",tcp_state="closewait"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp",tcp_state="closing"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp",tcp_state="established"} 10
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp",tcp_state="finwait1"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp",tcp_state="finwait2"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp",tcp_state="lastack"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp",tcp_state="listen"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp",tcp_state="synrecv"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp",tcp_state="synsent"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp",tcp_state="timewait"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp6",tcp_state="close"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp6",tcp_state="closewait"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp6",tcp_state="closing"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp6",tcp_state="established"} 4
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp6",tcp_state="finwait1"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp6",tcp_state="finwait2"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp6",tcp_state="lastack"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp6",tcp_state="listen"} 1
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp6",tcp_state="synrecv"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp6",tcp_state="synsent"} 0
pod_tcp_connections{namespace="monitoring",pod="prometheus-5788fcb75-b6c2z",proto="tcp6",tcp_state="timewait"} 0
```

## Getting Started

```
kubectl create -f deploy/daemonset.yaml
```

### Prerequisites

<!-- Describe packages, tools and everything we needed here -->

### Building

<!-- Describe how to build this project -->

```
make container
```

### Running

<!-- Describe how to run this project -->

## Versioning

<!-- Place versions of this project and write comments for every version -->

## Contributing

<!-- Tell others how to contribute this project -->

## Authors

<!-- Put authors here -->

## License

<!-- A link to license file -->

