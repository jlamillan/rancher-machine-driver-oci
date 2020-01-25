package oci

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/rancher/machine/libmachine/log"
	"strings"
	"time"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/identity"
)

// Client defines / contains the OCI/Identity clients and operations.
type Client struct {
	configuration        common.ConfigurationProvider
	computeClient        core.ComputeClient
	virtualNetworkClient core.VirtualNetworkClient
	identityClient       identity.IdentityClient
	sleepDuration        time.Duration
	// TODO we could also include the retry settings here
}

func newClient(configuration common.ConfigurationProvider) (*Client, error) {

	computeClient, err := core.NewComputeClientWithConfigurationProvider(configuration)
	if err != nil {
		log.Debugf("create new Compute client failed with err %v", err)
		return nil, err
	}
	vNetClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(configuration)
	if err != nil {
		log.Debugf("create new VirtualNetwork client failed with err %v", err)
		return nil, err
	}
	identityClient, err := identity.NewIdentityClientWithConfigurationProvider(configuration)
	if err != nil {
		log.Debugf("create new Identity client failed with err %v", err)
		return nil, err
	}
	c := &Client{
		configuration:        configuration,
		computeClient:        computeClient,
		virtualNetworkClient: vNetClient,
		identityClient:       identityClient,
		sleepDuration:        5,
	}
	return c, nil
}

// CreateInstance creates a new compute instance.
func (c *Client) CreateInstance(displayName, availabilityDomain, compartmentID, nodeShape, nodeImageName, nodeSubnetID, authorizedKeys string) (string, error) {

	// Create the launch compute instance request
	request := core.LaunchInstanceRequest{
		LaunchInstanceDetails: core.LaunchInstanceDetails{
			AvailabilityDomain: &availabilityDomain,
			CompartmentId:      &compartmentID,
			Shape:              &nodeShape,
			CreateVnicDetails: &core.CreateVnicDetails{
				SubnetId: &nodeSubnetID,
			},
			DisplayName: &displayName,
			Metadata: map[string]string{
				"ssh_authorized_keys": authorizedKeys,
				"user_data":           base64.StdEncoding.EncodeToString(createCloudInitScript()),
			},
			SourceDetails: core.InstanceSourceViaImageDetails{
				ImageId: c.getImageID(compartmentID, nodeImageName),
			},
		},
	}

	createResp, err := c.computeClient.LaunchInstance(context.Background(), request)
	if err != nil {
		return "", err
	}

	/*
		// should retry condition check which returns a bool value indicating whether to do retry or not
		// it checks the lifecycle status equals to Running or not for this case
		shouldRetryFunc := func(r common.OCIOperationResponse) bool {
			if converted, ok := r.Response.(core.GetInstanceResponse); ok {
				return converted.LifecycleState != core.InstanceLifecycleStateRunning
			}
			return true
		}
	*/
	// create get instance request with a retry policy which takes a function
	// to determine shouldRetry or not
	pollingGetRequest := core.GetInstanceRequest{
		InstanceId: createResp.Instance.Id,
		//RequestMetadata: helpers.GetRequestMetadataWithCustomizedRetryPolicy(shouldRetryFunc),
	}

	instance, pollError := c.computeClient.GetInstance(context.Background(), pollingGetRequest)
	if pollError != nil {
		return "", err
	}

	// Give the instance a bit of a head start to initialize.
	time.Sleep(30 * time.Second)

	return *instance.Id, nil
}

// GetInstance gets a compute instance by id.
func (c *Client) GetInstance(id string) (core.Instance, error) {
	instanceResp, err := c.computeClient.GetInstance(context.Background(), core.GetInstanceRequest{InstanceId: &id})
	return instanceResp.Instance, err
}

// TerminateInstance terminates a compute instance by id (does not wait).
func (c *Client) TerminateInstance(id string) error {
	_, err := c.computeClient.TerminateInstance(context.Background(), core.TerminateInstanceRequest{InstanceId: &id})
	return err
}

// GetInstanceIP returns the public IP (or private IP if that is what it has).
func (c *Client) GetInstanceIP(id, compartmentID string) (string, error) {
	vnics, err := c.computeClient.ListVnicAttachments(context.Background(), core.ListVnicAttachmentsRequest{
		InstanceId:    &id,
		CompartmentId: &compartmentID,
	})
	if err != nil {
		return "", err
	}

	if len(vnics.Items) == 0 {
		return "", errors.New("instance does not have any configured VNICs")
	}

	vnic, err := c.virtualNetworkClient.GetVnic(context.Background(), core.GetVnicRequest{VnicId: vnics.Items[0].VnicId})
	if err != nil {
		return "", err
	}

	if vnic.PublicIp == nil {
		return *vnic.PrivateIp, nil
	}

	return *vnic.PublicIp, nil
}

// Create the cloud init script
func createCloudInitScript() []byte {
	cloudInit := []string{
		"#!/bin/sh",
		"#echo \"Disabling OS firewall...\"",
		"sudo /usr/sbin/ethtool --offload $(/usr/sbin/ip -o -4 route show to default | awk '{print $5}') tx off",
		"sudo iptables -F",
		"",
		"# Update to sellinux that fixes write permission error",
		"sudo yum install -y http://mirror.centos.org/centos/7/extras/x86_64/Packages/container-selinux-2.99-1.el7_6.noarch.rpm",
		"#sudo sed -i  s/SELINUX=enforcing/SELINUX=permissive/ /etc/selinux/config",
		"sudo setenforce 0",
		"sudo systemctl stop firewalld.service",
		"sudo systemctl disable firewalld.service",
		"",
		"echo \"Installing Docker...\"",
		"curl https://releases.rancher.com/install-docker/18.09.9.sh | sh",
		"sudo usermod -aG docker opc",
		"sudo systemctl enable docker",
		"",
		"# Elasticsearch requirement",
		"sudo sysctl -w vm.max_map_count=262144",
	}
	return []byte(strings.Join(cloudInit, "\n"))
}

// getImageID gets the most recent ImageId for the node image name
func (c *Client) getImageID(compartmentID, nodeImageName string) *string {
	// Get list of images
	request := core.ListImagesRequest{
		CompartmentId:   &compartmentID,
		SortBy:          core.ListImagesSortByTimecreated,
		RequestMetadata: helpers.GetRequestMetadataWithDefaultRetryPolicy(),
	}
	r, err := c.computeClient.ListImages(context.Background(), request)
	if err != nil {
		return nil
	}

	// Loop through the items to find an image to use.  The list is sorted by time created in descending order
	for _, image := range r.Items {
		if strings.HasPrefix(*image.DisplayName, nodeImageName) {
			if !strings.Contains(*image.DisplayName, "GPU") {
				log.Debugf("Using node image %s", *image.DisplayName)
				return image.Id
			}
		}
	}

	return nil
}
