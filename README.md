### Netswatch
![Logo](pics/logo.png) 
Netswatch is a Docker networking scheme that leveraging Flannel(CoreOS) to establish a single-ported VxLAN overlay network between containers in different IDC.

## Prerequisites
* Etcd(v2, For Flannel backend)
* Consul(Name resolution for containers)
* A public IP(only single UDP port is needed, 8472)


## Environment
Name | Description | Default
--- | --- | ---
NW_ETCD_ENDPOINTS | Etcd endpoint | http://localhost:2379
NW_ETCD_USERNAME | Etcd username | 
NW_ETCD_PASSWORD | Etcd password | 
NW_PUBLIC_IP | Interface to use (IP or name) for inter-host communication | 
NW_IFACE | IP accessible by other nodes for inter-host communication | 
NW_ETCD_ENDPOINT | Netswatch Etcd endpoint | http://localhost:2379
NW_ETCD_USERNAME | Etcd username | 
NW_ETCD_PASSWORD | Etcd password |
NW_DNS_ENDPOINT | Consul DNS endpoint | http://localhost:8500
NW_DNS_TOKEN | Consul DNS token | 
NW_NETWORK_NAME | Netswatch bridge network name | default-net
NW_ORG | Netswatch organization name | default.com
NW_LOOP | Netswatch threading loop time(seconds) | 60
NW_QUEUE_TIMEOUT | Netswatch Docker events queue timeout(seconds) | 3
NW_NODE_TYPE | Netswatch routing type(router, node, internal) | internal
NW_NODE | Netswatch node name | default-node(when hostname not found)
NW_NETDATA_ENABLED | Extend Netdata service when registering | false
NW_NETDATA_PORT | Netdata metrics port | 19999
NW_LOG_LEVEL | Logging level | info



## Misc
Develop under Go 1.14


## Test
Still in progress

# Sidecar(To be migrated from Python version)

**sidecar** includes 2 functions:

* Generate Netwswatch network topology for debugging
* Synchronize nodes and subnets, and delete orphan nodes(nodes not in subnets)

![Logo](pics/preview.png) 

