# docker-machine-driver-oci
Oracle Cloud Infrastructure Driver Plugin for Docker/Rancher Machine

## Install docker-machine

`docker-machine` is required, [see documentation](https://docs.docker.com/machine/install-machine/).

```bash
go get github.com/docker/machine
cd $GOPATH/src/github.com/docker/machine
make build
make install
```

## Pre-create and configure your VCN

The VCN in which you want to create nodes must:

- allow inbound traffic to port 22 (SSH) to the node subnet.
- allow inbound traffic to port 2376 (Docker) to the node subnet.

## Build the OCI Plugin

```bash
go get github.com/jlamillan/docker-machine-driver-oci
cd $GOPATH/src/github.com/jlamillan/docker-machine-driver-oci
make install
```

## Provision a node using OCI plugin for docker-machine

```bash
$ docker-machine create -d oci --oci-region us-phoenix-1 --oci-subnet-id ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaa --oci-tenancy-id ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaa --oci-vcn-id ocid1.vcn.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaa --oci-fingerprint xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx --oci-node-availability-domain jGnV:PHX-1-AD2 --oci-node-image Oracle-Linux-7.6 --oci-user-id ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaa --oci-vcn-compartment-id ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaa --oci-node-compartment-id ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaa --oci-node-docker-port 2376 --oci-private-key-path /path/to/api.key.priv.pem  --oci-node-shape VM.Standard2.1 --oci-node-public-key-path /path/to/.ssh/id_rsa.pub node

Running pre-create checks...
(node) Verifying node image availability... 
Creating machine...
(node) Using node image Oracle-Linux-7.7-2019.12.18-0
Waiting for machine to be running, this may take a few minutes...
Detecting operating system of created instance...
Waiting for SSH to be available...
Detecting the provisioner...
Provisioning with ol...
Copying certs to the local machine directory...
Copying certs to the remote machine...
Setting Docker configuration on the remote daemon...
Checking connection to Docker...
Docker is up and running!
To see how to connect your Docker Client to the Docker Engine running on this virtual machine, run: docker-machine env node
```
