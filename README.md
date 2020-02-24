# rancher-machine-driver-oci

The Oracle Cloud Infrastructure (OCI) Driver for Rancher Machine allows Rancher to create and manage Kubernetes clusters on OCI.

## Create and configure your cluster's Virtual Cloud Network (VCN)

Rancher-machine requires the VCN in which you want to create nodes to:

- allow inbound traffic to port 22 (SSH) to the node subnet.
- allow inbound traffic to port 2376 (Docker) to the node subnet.

In addition to the above ports, [RKE](https://github.com/rancher/rke) has port requires for the different node types [detailed here](https://rancher.com/docs/rke/latest/en/os/#ports).

## Install OCI Node Driver for Rancher

1. From the Rancher Global view, choose Tools > Drivers > Node Drivers > Add Node Driver in the navigation bar.

2. Fill in the URLs of the latest Linux build of the [OCI Node Driver](https://github.com/rancher-plugins/rancher-machine-driver-oci) as well as the location of its [UI component](https://github.com/rancher-plugins/ui-node-driver-oci).

## Create Cloud Credentials for OCI

1. From your user settings, choose > Cloud Credentials > Add Cloud Credential.

2. Select "Oracle Cloud Infrastructure" from the drop down, and fill in your account credentials (tenancy, user, signing key, etc.).

## Provision Kubernetes cluster on OCI

1. From the Global view, choose Clusters > Add Cluster.

2. From the infrastructure providers, choose the Oracle Cloud Infrastructure icon. 

3. Fill in a cluster name, and add Node Template(s) for the various node types (etcd, Control Plane, or Worker).

4. After you've created a template(s), you can use it provision a new Kubernetes cluster on OCI.

You can access the cluster after its state is updated to Active.

## Optional, Deploy OCI Cloud Controller Manager and FlexVolume Driver

The OCI [CCM](https://github.com/oracle/oci-cloud-controller-manager) and [FlexVolume driver](https://github.com/oracle/oci-flexvolume-driver) for Kubernetes lets you dynamically provision and manage load-balancers and block storage volumes on OCI.

## Optional, build the OCI Plugin binary

```bash
go get github.com/rancher-plugins/rancher-machine-driver-oci
cd $GOPATH/src/github.com/rancher-plugins/rancher-machine-driver-oci
make install
```

## Optional, install rancher-machine CLI binary

Stand-alone `rancher-machine` is useful, but optional since it is included with Rancher server.

```bash
go get github.com/rancher/machine
cd $GOPATH/src/github.com/rancher/machine
make build
make install
```

## Optional, provision a node using OCI plugin for rancher-machine CLI

```bash
$ rancher-machine create -d oci --oci-region us-phoenix-1 --oci-subnet-id ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaa --oci-tenancy-id ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaa --oci-vcn-id ocid1.vcn.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaa --oci-fingerprint xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx --oci-node-availability-domain jGnV:PHX-1-AD2 --oci-node-image Oracle-Linux-7.6 --oci-user-id ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaa --oci-vcn-compartment-id ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaa --oci-node-compartment-id ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaa --oci-node-docker-port 2376 --oci-private-key-path /path/to/api.key.priv.pem  --oci-node-shape VM.Standard2.1 --oci-node-public-key-path /path/to/.ssh/id_rsa.pub node

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



