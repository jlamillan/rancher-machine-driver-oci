# rancher-machine-driver-oci

The Oracle Cloud Infrastructure (OCI) Driver for (Rancher) Docker Machine allows Rancher to create and manage Kubernetes clusters on OCI.

## Create and configure your cluster's Virtual Cloud Network (VCN)

(Rancher) Docker Machine requires the VCN in which you want to create nodes to:

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

## Optional, integrate with OCI Kubernetes Addons

* [OCI Flexvolume Driver](https://github.com/oracle/oci-flexvolume-driver) (for provisioning [Block Volumes](https://docs.cloud.oracle.com/en-us/iaas/Content/Block/Concepts/overview.htm))
* [OCI Cloud Controller Manager](https://github.com/oracle/oci-cloud-controller-manager) (for provisioning [LoadBalancers](https://docs.cloud.oracle.com/en-us/iaas/Content/Balance/Concepts/balanceoverview.htm))


## Develop and Test

See [development](docs/development.md) for details.
