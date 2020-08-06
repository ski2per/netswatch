### Netswatch
![Logo](sidecar/static/logo.png) 
Netswatch is a Docker networking scheme that leveraging Flannel(CoreOS) to establish a single-ported VxLAN overlay network between containers in different IDC.

## Prerequisites
* Etcd(v2, For Flannel backend)
* Consul(Name resolution for containers)
* A public IP(only single UDP port is needed, 8472)


## Environment
Name | Description | Default
--- | --- | ---
FLANNELD_ETCD_ENDPOINTS | Etcd endpoint | http://localhost:2379
FLANNELD_ETCD_USERNAME | Etcd username | 
FLANNELD_ETCD_PASSWORD | Etcd password | 
FLANNELD_PUBLIC_IP | Interface to use (IP or name) for inter-host communication | 
FLANNELD_IFACE | IP accessible by other nodes for inter-host communication | 
NETSWATCH_ETCD_ENDPOINT | Netswatch Etcd endpoint | http://localhost:2379
NETSWATCH_ETCD_USERNAME | Etcd username | 
NETSWATCH_ETCD_PASSWORD | Etcd password |
NETSWATCH_DNS_ENDPOINT | Consul DNS endpoint | http://localhost:8500
NETSWATCH_DNS_TOKEN | Consul DNS token | 
NETSWATCH_NETWORK_NAME | Netswatch bridge network name | default-net
NETSWATCH_ORG | Netswatch organization name | default.com
NETSWATCH_LOOP | Netswatch threading loop time(seconds) | 60
NETSWATCH_QUEUE_TIMEOUT | Netswatch Docker events queue timeout(seconds) | 3
NETSWATCH_NODE_TYPE | Netswatch routing type(router, node, internal) | internal
NETSWATCH_NODE | Netswatch node name | default-node(when hostname not found)
NETSWATCH_NETDATA_ENABLED | Extend Netdata service when registering | false
NETSWATCH_NETDATA_PORT | Netdata metrics port | 19999
NETSWATCH_LOG_LEVEL | Logging level | info



## Misc
Develop under Python 3.7+

Pyinstaller must be run on OS with same version


## Test
Still in progress

`python -m unittest discover -s tests -v`

# Sidecar
**sidecar** includes 2 functions:

* Generate Netwswatch network topology for debugging
* Synchronize nodes and subnets, and delete orphan nodes(nodes not in subnets)

![Logo](sidecar/static/preview.png) 

