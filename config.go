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

	// Network settings
	IPPrefix  = "192.168.2."
	IPOffset  = 10 // number 1 -> 192.168.2.11
	DefaultGW = "192.168.2.1"
)

// numberToIP derives an IP address from the epair number
func numberToIP(n int) string {
	return IPPrefix + itoa(n+IPOffset)
}

// itoa converts int to string (simple version of strconv.Itoa)
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
