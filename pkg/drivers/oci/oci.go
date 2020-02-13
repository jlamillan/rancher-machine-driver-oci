package oci

import (
	"errors"
	"fmt"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/rancher/machine/libmachine/drivers"
	"github.com/rancher/machine/libmachine/log"
	"github.com/rancher/machine/libmachine/mcnflag"
	"github.com/rancher/machine/libmachine/state"
	"io/ioutil"
	"strings"
)

const (
	defaultNodeNamePfx = "oci-node-driver-"
	defaultSSHPort     = 22
	defaultSSHUser     = "opc"
	defaultImage       = "Oracle-Linux-7.7"
	defaultDockerPort  = 2376
)

// Driver is the implementation of BaseDriver interface
type Driver struct {
	*drivers.BaseDriver
	VCNID                    string
	SubnetID                 string
	TenancyID                string
	CompartmentID            string
	UserID                   string
	Region                   string
	Fingerprint              string
	PrivateKeyPath           string
	PrivateKeyContents       string
	PrivateKeyPassphrase     string
	NodePublicSSHKeyPath     string
	NodePublicSSHKeyContents string
	AvailabilityDomain       string
	Shape                    string
	Image                    string
	DockerPort               int
	// Runtime values
	InstanceID string
}

// NewDriver creates a new driver
func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	log.Debug("oci.Create()")

	oci, err := d.initOCIClient()
	if err != nil {
		return err
	}

	var image = ""
	if d.Image == "" {
		image = defaultImage
	}

	d.InstanceID, err = oci.CreateInstance(defaultNodeNamePfx+d.MachineName, d.AvailabilityDomain, d.CompartmentID, d.Shape, image, d.SubnetID, d.NodePublicSSHKeyContents)
	if err != nil {
		return err
	}

	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	log.Debug("oci.DriverName()")
	return "oci"
}

// GetCreateFlags returns the mcnflag.Flag slice representing the flags
// that can be set, their descriptions and defaults.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	log.Debug("oci.GetCreateFlags()")
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "oci-tenancy-id",
			Usage:  "The OCID of the tenancy in which to create node(s)",
			EnvVar: "OCI_TENANCY_ID",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "oci-vcn-id",
			Usage:  "Pre-existing VCN id in which you want to create the node(s)",
			EnvVar: "OCI_VCN_ID",
		},
		mcnflag.StringFlag{
			Name:   "oci-subnet-id",
			Usage:  "Pre-existing subnet id in which you want to create the node(s)",
			EnvVar: "OCI_SUBNET_ID",
		},
		mcnflag.StringFlag{
			Name:   "oci-compartment-id",
			Usage:  "The OCID of the compartment in which to create node(s)",
			EnvVar: "OCI_COMPARTMENT_ID",
		},
		mcnflag.StringFlag{
			Name:   "oci-user-id",
			Usage:  "The OCID of a user who has access to the specified tenancy/compartment",
			EnvVar: "OCI_USER_ID",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "oci-region",
			Usage:  "The region in which to create node(s)",
			EnvVar: "OCI_REGION",
		},
		mcnflag.StringFlag{
			Name:   "oci-fingerprint",
			Usage:  "The fingerprint corresponding to the specified user's private API Key",
			EnvVar: "OCI_FINGERPRINT",
		},
		mcnflag.StringFlag{
			Name:   "oci-private-key-path",
			Usage:  "The private API key path for the specified OCI user, in PEM format",
			EnvVar: "OCI_PRIVATE_KEY_PATH",
		},
		mcnflag.StringFlag{
			Name:   "oci-private-key-contents",
			Usage:  "The private API key contents for the specified OCI user, in PEM format",
			EnvVar: "OCI_PRIVATE_KEY_CONTENTS",
		},
		mcnflag.StringFlag{
			Name:   "oci-private-key-passphrase",
			Usage:  "The passphrase (if any) that protects private key file the specified OCI user",
			EnvVar: "OCI_PRIVATE_KEY_PASSPHRASE",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "oci-node-public-key-path",
			Usage:  "Optional SSH public key for the nodes",
			EnvVar: "OCI_NODE_PUBLIC_KEY_PATH",
		},
		mcnflag.StringFlag{
			Name:   "oci-node-availability-domain",
			Usage:  "The availability domain of the node(s) should be placed in",
			EnvVar: "OCI_NODE_AVAILABILITY_DOMAIN",
		},
		mcnflag.StringFlag{
			Name:   "oci-node-shape",
			Usage:  "The instance shape of the node(s)",
			EnvVar: "OCI_NODE_SHAPE",
		},
		mcnflag.StringFlag{
			Name:   "oci-node-image",
			Usage:  "The image to use for the node(s)",
			EnvVar: "OCI_NODE_IMAGE",
		},
		mcnflag.IntFlag{
			Name:   "oci-node-docker-port",
			Usage:  "Docker Port",
			Value:  defaultDockerPort,
			EnvVar: "OCI_NODE_DOCKER_PORT",
		},
	}
}

// GetIP returns an IP or hostname that this host is available at
// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
func (d *Driver) GetIP() (string, error) {
	log.Debug("oci.GetIP()")
	oci, err := d.initOCIClient()
	if err != nil {
		return "", err
	}

	return oci.GetInstanceIP(d.InstanceID, d.CompartmentID)
}

// GetMachineName returns the name of the machine
func (d *Driver) GetMachineName() string {
	log.Debug("oci.GetMachineName()")
	return d.MachineName
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	log.Debug("oci.GetSSHHostname()")
	return d.GetIP()
}

// GetSSHKeyPath returns key path for use with ssh
func (d *Driver) GetSSHKeyPath() string {
	log.Debug("oci.GetSSHKeyPath()")
	return strings.Replace(d.NodePublicSSHKeyPath, ".pub", "", 1)
}

// GetSSHPort returns port for use with ssh
func (d *Driver) GetSSHPort() (int, error) {
	log.Debug("oci.GetSSHPort()")
	return defaultSSHPort, nil
}

// GetSSHUsername returns username for use with ssh
func (d *Driver) GetSSHUsername() string {
	log.Debug("oci.GetSSHUsername()")
	return defaultSSHUser
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *Driver) GetURL() (string, error) {
	log.Debug("oci.GetURL()")
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}

	return fmt.Sprintf("tcp://%s:%d", ip, defaultDockerPort), nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	log.Debug("oci.GetState()")

	oci, err := d.initOCIClient()
	if err != nil {
		return state.None, err
	}

	instance, err := oci.GetInstance(d.InstanceID)
	if err != nil {
		return state.None, err
	}

	switch instance.LifecycleState {
	case core.InstanceLifecycleStateRunning:
		return state.Running, nil
	case core.InstanceLifecycleStateStopped, core.InstanceLifecycleStateTerminated:
		return state.Stopped, nil
	case core.InstanceLifecycleStateStopping, core.InstanceLifecycleStateTerminating:
		return state.Stopping, nil
	case core.InstanceLifecycleStateStarting, core.InstanceLifecycleStateProvisioning, core.InstanceLifecycleStateCreatingImage:
		return state.Starting, nil
	}

	// deleting, migrating, rebuilding, cloning, restoring ...
	return state.None, nil

}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	log.Debug("oci.Kill()")
	return d.Remove()
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	log.Debug("oci.PreCreateCheck()")

	// Check that the node image exists, which will also validate the credentials.
	log.Infof("Verifying node image availability... ")

	oci, err := d.initOCIClient()
	if err != nil {
		return err
	}

	image := oci.getImageID(d.CompartmentID, defaultImage)

	if len(*image) == 0 {
		return fmt.Errorf("could not retrieve node image ID from OCI")
	}

	// TODO, verify VCN and subnet

	return nil
}

// Remove a host
func (d *Driver) Remove() error {
	log.Debug("oci.Remove()")

	oci, err := d.initOCIClient()
	if err != nil {
		return err
	}

	return oci.TerminateInstance(d.InstanceID)
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *Driver) Restart() error {
	// TODO
	log.Debug("oci.Restart()")
	return nil
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	log.Debug("oci.SetConfigFromFlags(...)")
	d.VCNID = flags.String("oci-vcn-id")
	if d.VCNID == "" {
		return errors.New("no OCI VCNID specified (--oci-vcn-id)")
	}
	d.SubnetID = flags.String("oci-subnet-id")
	if d.SubnetID == "" {
		return errors.New("no OCI subnetId specified (--oci-subnet-id)")
	}
	d.TenancyID = flags.String("oci-tenancy-id")
	if d.TenancyID == "" {
		return errors.New("no OCI tenancy specified (--oci-tenancy-id)")
	}
	d.CompartmentID = flags.String("oci-compartment-id")
	if d.CompartmentID == "" {
		return errors.New("no OCI compartment specified (--oci-compartment-id)")
	}
	d.UserID = flags.String("oci-user-id")
	if d.UserID == "" {
		return errors.New("no OCI user id specified (--oci-user-id)")
	}
	d.Region = flags.String("oci-region")
	if d.Region == "" {
		return errors.New("no OCI oci-region specified (--oci-region)")
	}
	d.AvailabilityDomain = flags.String("oci-node-availability-domain")
	if d.AvailabilityDomain == "" {
		return errors.New("no OCI node availability domain specified (--oci-node-availability-domain)")
	}
	d.Shape = flags.String("oci-node-shape")
	if d.Shape == "" {
		return errors.New("no OCI node shape specified (--oci-node-shape)")
	}
	d.Fingerprint = flags.String("oci-fingerprint")
	if d.Fingerprint == "" {
		return errors.New("no OCI oci-fingerprint specified (--oci-fingerprint)")
	}
	d.PrivateKeyPath = flags.String("oci-private-key-path")
	if d.PrivateKeyPath == "" {
		return errors.New("no private key path specified (--oci-private-key-path)")
	}
	if d.PrivateKeyContents == "" && d.PrivateKeyPath != "" {
		privateKeyBytes, err := ioutil.ReadFile(d.PrivateKeyPath)
		if err == nil {
			d.PrivateKeyContents = string(privateKeyBytes)
		}
	}
	d.PrivateKeyContents = flags.String("oci-private-key-contents")
	d.NodePublicSSHKeyPath = flags.String("oci-node-public-key-path")
	if d.NodePublicSSHKeyPath == "" {
		return errors.New("no public key path specified (--oci-node-public-key-path)")
	}
	if d.NodePublicSSHKeyContents == "" && d.NodePublicSSHKeyPath != "" {
		publicKeyBytes, err := ioutil.ReadFile(d.NodePublicSSHKeyPath)
		if err == nil {
			d.NodePublicSSHKeyContents = string(publicKeyBytes)
		}
	}
	d.Image = flags.String("oci-node-image")
	return nil
}

// Start a host
func (d *Driver) Start() error {
	// TODO
	log.Debug("oci.Start()")
	return nil
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	// TODO
	log.Debug("oci.Stop()")
	return nil
}

// initOCIClient is a helper function that constructs a new
// oci.Client based on config values.
func (d *Driver) initOCIClient() (Client, error) {
	configurationProvider := common.NewRawConfigurationProvider(
		d.TenancyID,
		d.UserID,
		d.Region,
		d.Fingerprint,
		d.PrivateKeyContents,
		&d.PrivateKeyPassphrase)

	ociClient, err := newClient(configurationProvider)
	if err != nil {
		return Client{}, err
	}

	return *ociClient, nil
}
