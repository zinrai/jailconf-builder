package main

const (
	// Directory paths
	JailConfDir = "/etc/jail.conf.d"
	JailRootDir = "/var/jails"
	BaseDir     = "/var/db/jailconf-builder/base"

	// Main configuration file
	MainJailConf = "/etc/jail.conf"

	// Include directive
	IncludeLine = `.include "/etc/jail.conf.d/*.conf";`
)
