### Netswatch
![Logo](pics/logo.png) 
Netswatch is a Docker networking fabric that leveraging Flannel(CoreOS) to establish a single-ported VxLAN overlay network between containers in different IDC.

## Prerequisites
* Etcd(v2, For Flannel backend)
* Consul(Service registration, Name resolution for containers)
* A public IP(only single UDP port is needed, default to 8472)

## Quick Start
Run as binary
```
make dist/netswatch-amd64
./dist/netswatch-amd64
```

Run as container
```
make image
```
use *docker-compose.yml* to run

## Environment
Name | Description | Default
--- | --- | ---
NW_ETCD_ENDPOINTS | Etcd endpoint | http://localhost:2379
NW_ETCD_PREFIX | Etcd prefix used for configuration import | /netswatch/network
NW_ETCD_USERNAME | Etcd username | 
NW_ETCD_PASSWORD | Etcd password | 
NW_PUBLIC_IP | Interface to use (IP or name) for inter-IDC communication | 
NW_IFACE | IP accessible by other nodes for inter-host communication | 
NW_DNS_ENDPOINT | Consul DNS endpoint | http://localhost:8500
NW_DNS_TOKEN | Consul DNS token | 
NW_SUBNET_FILE | File to store subnet info | /run/flannel/subnet.env
NW_NETWORK_NAME | Netswatch bridge network name | netswatch
NW_ORG_NAME | Netswatch organization name | default.local
NW_NODE_TYPE | Netswatch routing type(router, node, internal) | internal
NW_NODE_NAME | Netswatch node name | default-node(when hostname not found)
NW_NETDATA_ENABLED | Extend Netdata for service registration | false
NW_NETDATA_PORT | Netdata metrics port | 19999
NW_LOOP | Netswatch max loop(seconds) | 600 
NW_LOG_LEVEL | Logging level | info



## Misc
Develop under Go 1.14


## Test
Still in progress


## Known issues and todo
* Add periodic check for service register(services won't be registered when start before Netswatch)

# [Sidecar](https://github.com/ski2per/s1decar.git)

**sidecar** includes 1 functions:

* Generate Netwswatch network topology for debugging

![Logo](pics/preview.png) 

