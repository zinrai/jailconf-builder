package main

import (
	"fmt"
	"os"
)

// Jail represents a jail configuration
type Jail struct {
	Name    string
	Number  int
	IPAddr  string
	Gateway string
}

// NewJail creates a new Jail instance
func NewJail(name string, number int) *Jail {
	return &Jail{
		Name:    name,
		Number:  number,
		IPAddr:  numberToIP(number),
		Gateway: DefaultGW,
	}
}

// GenerateConf generates the jail.conf content
func (j *Jail) GenerateConf() string {
	return fmt.Sprintf(`%s {
    host.hostname = "%s.jail";
    path = "%s/%s";

    vnet;
    vnet.interface = "epair%db";

    $ip4_addr = "%s";
    $gw = "%s";

    exec.prestart  = "ifconfig epair%d create up";
    exec.prestart += "ifconfig bridge0 addm epair%da";
    exec.start     = "ifconfig lo0 up 127.0.0.1";
    exec.start    += "ifconfig epair%db up $ip4_addr";
    exec.start    += "route add default $gw";
    exec.start    += "sh /etc/rc";
    exec.stop      = "sh /etc/rc.shutdown";
    exec.poststop  = "ifconfig epair%da destroy";

    mount.devfs;
    devfs_ruleset = 5;
    persist;
}
`, j.Name, j.Name, JailRootDir, j.Name,
		j.Number, j.IPAddr, j.Gateway,
		j.Number, j.Number, j.Number, j.Number)
}

// ConfPath returns the path to the jail.conf file
func (j *Jail) ConfPath() string {
	return fmt.Sprintf("%s/%d-%s.conf", JailConfDir, j.Number, j.Name)
}

// RootPath returns the path to the jail root directory
func (j *Jail) RootPath() string {
	return fmt.Sprintf("%s/%s", JailRootDir, j.Name)
}

// WriteConf writes the jail.conf file
func (j *Jail) WriteConf() error {
	content := j.GenerateConf()
	return os.WriteFile(j.ConfPath(), []byte(content), 0644)
}
