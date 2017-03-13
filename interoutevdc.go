package interoutevdc

import (
        "fmt"
        "io/ioutil"

        "github.com/docker/machine/libmachine/drivers"
        "github.com/docker/machine/libmachine/log"
        "github.com/docker/machine/libmachine/mcnflag"
        "github.com/docker/machine/libmachine/ssh"
        "github.com/docker/machine/libmachine/state"
        "github.com/Interoute/go-cloudstack/cloudstack"
)

const (
        driverName = "interoutevdc"
        dockerPort = 2376
)

type configError struct {
        option string
}

func (e *configError) Error() string {
        return fmt.Sprintf("Interoute VDC driver requires the --interoutevdc-%s option", e.option)
}

type Driver struct {
        *drivers.BaseDriver
        Id                   string
        ApiURL               string
        ApiKey               string
        SecretKey            string
        UsePrivateIP         bool
        SSHKeyPair           string
        PrivateIP            string
        TemplateID           string
	TemplateFilter	     string
        ServiceOfferingID    string
        NetworkID            string
        ZoneID               string
        NetworkType          string
        VDCRegion            string
	DiskOfferingID	     string
	DiskSize	     int
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
        return []mcnflag.Flag{
                mcnflag.StringFlag{
                        Name:   "interoutevdc-apiurl",
                        Usage:  "Interoute VDC API URL",
                        EnvVar: "INTEROUTEVDC_API_URL",
                        Value: "https://myservices.interoute.com/myservices/api/vdc",
                },
                mcnflag.StringFlag{
                        Name:   "interoutevdc-apikey",
                        Usage:  "Interoute VDC API key",
                        EnvVar: "INTEROUTEVDC_API_KEY",
                },
                mcnflag.StringFlag{
                        Name:   "interoutevdc-secretkey",
                        Usage:  "Interoute VDC API secret key",
                        EnvVar: "INTEROUTEVDC_SECRET_KEY",
                },
                mcnflag.StringFlag{
                        Name:  "interoutevdc-templateid",
                        Usage: "Interoute VDC template ID",
                },
                mcnflag.StringFlag{
                        Name:  "interoutevdc-serviceofferingid",
                        Usage: "Interoute VDC service offering ID",
                },
                mcnflag.StringFlag{
                        Name:  "interoutevdc-networkid",
                        Usage: "Interoute VDC network ID",
                },
                mcnflag.StringFlag{
                        Name:  "interoutevdc-zoneid",
                        Usage: "Interoute VDC zone ID",
                },
		mcnflag.StringFlag{
                        Name:  "interoutevdc-templatefilter",
                        Usage: "Interoute VDC template filter",
                },
		mcnflag.StringFlag{
                        Name:  "interoutevdc-diskofferingid",
                        Usage: "Interoute VDC disk offering ID",
                },
                mcnflag.IntFlag{
                        Name:  "interoutevdc-disksize",
                        Usage: "Interoute VDC additional disk size",
                },
                mcnflag.StringFlag{
                        Name:  "interoutevdc-vdcregion",
                        Usage: "Interoute VDC Region",
			EnvVar: "INTEROUTEVDC_REGION",
                },
        }
}

func NewDriver(hostName, storePath string) drivers.Driver {

        driver := &Driver{
                BaseDriver: &drivers.BaseDriver{
                        MachineName: hostName,
                        StorePath:   storePath,
                },
        }
        return driver
}

func (d *Driver) DriverName() string {
        return driverName
}

func (d *Driver) GetSSHHostname() (string, error) {
        return d.GetIP()
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
        d.ApiURL = flags.String("interoutevdc-apiurl")
        d.ApiKey = flags.String("interoutevdc-apikey")
        d.SecretKey = flags.String("interoutevdc-secretkey")
        d.UsePrivateIP = true
        d.VDCRegion = flags.String("interoutevdc-vdcregion")
        d.TemplateID = flags.String("interoutevdc-templateid")
	d.TemplateFilter = flags.String("interoutevdc-templatefilter")
        d.ServiceOfferingID = flags.String("interoutevdc-serviceofferingid")
        d.NetworkID = flags.String("interoutevdc-networkid")
	d.DiskOfferingID = flags.String("interoutevdc-diskofferingid")
	d.DiskSize = flags.Int("interoutevdc-disksize")

        if err := d.setZone(flags.String("interoutevdc-zoneid")); err != nil {
                return err
        }

        d.SSHKeyPair = d.MachineName

        if d.ApiURL == "" {
                return &configError{option: "apiurl"}
        }

        if d.ApiKey == "" {
                return &configError{option: "apikey"}
        }

        if d.SecretKey == "" {
                return &configError{option: "secretkey"}
        }

        if d.VDCRegion == "" {
                return &configError{option: "vdcregion"}
        }

        if d.TemplateID == "" {
                return &configError{option: "templateid"}
        }

        if d.ServiceOfferingID == "" {
                return &configError{option: "serviceofferingid"}
        }

        if d.ZoneID == "" {
                return &configError{option: "zoneid"}
        }

        return nil
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func (d *Driver) GetSSHUsername() string {
        cs := d.getClient()
        template, _, _ := cs.Template.GetTemplateByID(d.TemplateID, d.TemplateFilter, d.ZoneID)

	switch {
		case hasPrefix(template.Ostypename, "CentOS"):
			d.SSHUser = "centos"
		case hasPrefix(template.Ostypename, "Ubuntu"):
			d.SSHUser = "ubuntu"
                case hasPrefix(template.Ostypename, "Red Hat"):
                        d.SSHUser = "redhat"
		default:
			d.SSHUser = "ubuntu"
	}

        return d.SSHUser
}

func (d *Driver) GetURL() (string, error) {
        ip, err := d.GetIP()
        if err != nil {
                return "", err
        }
        return fmt.Sprintf("tcp://%s:%d", ip, dockerPort), nil
}

func (d *Driver) GetIP() (string, error) {
        return d.PrivateIP, nil
}

func (d *Driver) GetState() (state.State, error) {
        cs := d.getClient()
        vm, count, err := cs.VirtualMachine.GetVirtualMachineByID(d.Id)
        if err != nil {
                return state.Error, err
        }

        if count == 0 {
                return state.None, fmt.Errorf("Machine does not exist, use create command to create it")
        }

        switch vm.State {
        case "Starting":
                return state.Starting, nil
        case "Running":
                return state.Running, nil
        case "Stopping":
                return state.Running, nil
        case "Stopped":
                return state.Stopped, nil
        case "Destroyed":
                return state.Stopped, nil
        case "Expunging":
                return state.Stopped, nil
        case "Migrating":
                return state.Paused, nil
        case "Error":
                return state.Error, nil
        case "Unknown":
                return state.Error, nil
        case "Shutdowned":
                return state.Stopped, nil
        }

        return state.None, nil
}

func (d *Driver) PreCreateCheck() error {

        if err := d.checkKeyPair(); err != nil {
                return err
        }

        if err := d.checkInstance(); err != nil {
                return err
        }

        return nil
}

func (d *Driver) Create() error {
        cs := d.getClient()

        if err := d.createKeyPair(); err != nil {
                return err
        }

        p := cs.VirtualMachine.NewDeployVirtualMachineParams(d.ServiceOfferingID, d.TemplateID, d.ZoneID)
        p.SetName(d.MachineName)
        p.SetDisplayname(d.MachineName)
        p.SetDetails(map[string]string{"workload": "ranchernode",})
        p.SetKeypair(d.SSHKeyPair)
	if d.DiskOfferingID != "" {
		p.SetDiskofferingid(d.DiskOfferingID)
		p.SetSize(int64(d.DiskSize))
	}

        if d.NetworkID != "" {
                p.SetNetworkids([]string{d.NetworkID})
        }

        // Create the machine
        log.Info("Creating Interoute VDC instance...")
        vm, err := cs.VirtualMachine.DeployVirtualMachine(p)
        if err != nil {
                return err
        }

        d.Id = vm.Id

        d.PrivateIP = vm.Nic[0].Ipaddress

        return nil
}

func (d *Driver) Remove() error {
        cs := d.getClient()
        p := cs.VirtualMachine.NewDestroyVirtualMachineParams(d.Id)
	p.SetExpunge(true)

        log.Info("Removing Interoute VDC instance...")
        if _, err := cs.VirtualMachine.DestroyVirtualMachine(p); err != nil {
                return err
        }

        if err := d.deleteKeyPair(); err != nil {
                return err
        }

        return nil
}

func (d *Driver) Start() error {
        vmstate, err := d.GetState()
        if err != nil {
                return err
        }

        if vmstate == state.Running {
                log.Info("Machine is already running")
                return nil
        }

        if vmstate == state.Starting {
                log.Info("Machine is already starting")
                return nil
        }

        cs := d.getClient()
        p := cs.VirtualMachine.NewStartVirtualMachineParams(d.Id)

        if _, err = cs.VirtualMachine.StartVirtualMachine(p); err != nil {
                return err
        }

        return nil
}

func (d *Driver) Stop() error {
        vmstate, err := d.GetState()
        if err != nil {
                return err
        }

        if vmstate == state.Stopped {
                log.Info("Machine is already stopped")
                return nil
        }

        cs := d.getClient()
        p := cs.VirtualMachine.NewStopVirtualMachineParams(d.Id)

        if _, err = cs.VirtualMachine.StopVirtualMachine(p); err != nil {
                return err
        }

        return nil
}

func (d *Driver) Restart() error {
        vmstate, err := d.GetState()
        if err != nil {
                return err
        }

        if vmstate == state.Stopped {
                return fmt.Errorf("Machine is stopped, use start command to start it")
        }

        cs := d.getClient()
        p := cs.VirtualMachine.NewRebootVirtualMachineParams(d.Id)

        if _, err = cs.VirtualMachine.RebootVirtualMachine(p); err != nil {
                return err
        }

        return nil
}

func (d *Driver) Kill() error {
        return d.Stop()
}

func (d *Driver) getClient() *cloudstack.CloudStackClient {
        cs := cloudstack.NewAsyncClient(d.ApiURL, d.ApiKey, d.SecretKey, d.VDCRegion, false)
        return cs
}

func (d *Driver) setZone(zoneid string) error {
        d.ZoneID = zoneid
        d.NetworkType = ""

        if d.ZoneID == "" {
                return nil
        }

        cs := d.getClient()
        z, _, err := cs.Zone.GetZoneByID(d.ZoneID)
        if err != nil {
                return fmt.Errorf("Unable to get zoneid: %v", err)
        }

        d.NetworkType = z.Networktype

        log.Debugf("zone id: %q", d.ZoneID)
        log.Debugf("network type: %q", d.NetworkType)

        return nil
}

func (d *Driver) checkKeyPair() error {
        cs := d.getClient()

        log.Infof("Checking if SSH key pair (%v) already exists...", d.SSHKeyPair)

        p := cs.SSH.NewListSSHKeyPairsParams()
        p.SetName(d.SSHKeyPair)
        res, err := cs.SSH.ListSSHKeyPairs(p)
        if err != nil {
                return err
        }
        if res.Count > 0 {
                return fmt.Errorf("SSH key pair (%v) already exists.", d.SSHKeyPair)
        }
        return nil
}

func (d *Driver) checkInstance() error {
        cs := d.getClient()

        log.Infof("Checking if instance (%v) already exists...", d.MachineName)

        p := cs.VirtualMachine.NewListVirtualMachinesParams()
        p.SetName(d.MachineName)
        p.SetZoneid(d.ZoneID)
        res, err := cs.VirtualMachine.ListVirtualMachines(p)
        if err != nil {
                return err
        }
        if res.Count > 0 {
                return fmt.Errorf("Instance (%v) already exists.", d.SSHKeyPair)
        }
        return nil
}

func (d *Driver) createKeyPair() error {
        cs := d.getClient()

        if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
                return err
        }

        publicKey, err := ioutil.ReadFile(d.GetSSHKeyPath() + ".pub")
        if err != nil {
                return err
        }

        log.Infof("Registering SSH key pair...")

        p := cs.SSH.NewRegisterSSHKeyPairParams(d.SSHKeyPair, string(publicKey))
        if _, err := cs.SSH.RegisterSSHKeyPair(p); err != nil {
                return err
        }

        return nil
}

func (d *Driver) deleteKeyPair() error {
        cs := d.getClient()

        log.Infof("Deleting SSH key pair...")

        p := cs.SSH.NewDeleteSSHKeyPairParams(d.SSHKeyPair)
        if _, err := cs.SSH.DeleteSSHKeyPair(p); err != nil {
                return err
        }
        return nil
}
