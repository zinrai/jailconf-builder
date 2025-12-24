# jailconf-builder

`jailconf-builder` is a CLI tool for managing FreeBSD jails using the standard [jail.conf(5)](https://man.freebsd.org/cgi/man.cgi?jail.conf(5)) configuration.

Back around 2013, when I was working with jail.conf(5) on FreeBSD 9.x, I wished there was an include option to improve usability. Recently, I revisited jail.conf(5) and discovered that the include option is indeed implemented. This discovery led to the development of this `jailconf-builder`, which uses the standard features of the jail system.

## Note

This is a CLI tool for creating jail environments using jail.conf(5). For jail operations, please use commands such as `jail -c <jail_name>` , `jail -r <jail_name>` , `jls` , and `jexec 1 /bin/tcsh`.

## Features

- Initialize `jailconf-builder` environment
- Create new jails with VNET support
- Delete jails with safety checks
- Download FreeBSD base system for jails

## Directory Structure

```
/etc/jail.conf.d/                # Jail configuration files (FreeBSD standard)
/var/jails/                      # Jail root directories
/var/db/jailconf-builder/base/   # FreeBSD base systems
```

## Installation

Build the tool:

```
$ GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o jailconf-builder
```

## Network Setup

Before using `jailconf-builder` , you need to set up the network environment. Run the following commands as root:

Create and configure the bridge interface:

```sh
# ifconfig bridge create
# ifconfig bridge0 inet 192.168.2.1/24
# ifconfig bridge0 up
```

Enable PF (Packet Filter):

```sh
# cat << EOF >> /etc/rc.conf
pf_enable="YES"
pf_flags=""
EOF
```

Configure NAT:

```sh
# echo 'nat on vtnet0 from 192.168.2.0/24 to any -> (vtnet0)' > /etc/pf.conf
# echo 'pass all' >> /etc/pf.conf
```

Enable IP forwarding:

```sh
# sysctl net.inet.ip.forwarding=1
```

Start the PF service:

```sh
# service pf start
```

Note: Replace `vtnet0` with your actual network interface name if different.

For more information on configuring PF, refer to the [FreeBSD Handbook section on Firewalls](https://docs.freebsd.org/en/books/handbook/firewalls/#_enabling_pf).

## Usage

### Initialize

Before using `jailconf-builder` , you need to initialize the environment:

```
$ sudo jailconf-builder init
```

This command:
- Confirms `/etc/jail.conf.d` directory exists
- Adds include directive to `/etc/jail.conf`
- Creates `/var/jails` directory
- Creates `/var/db/jailconf-builder/base` directory

### Download FreeBSD Base System

To download the FreeBSD base system for a specific version:

```
$ sudo jailconf-builder dl-base -s <URL_to_base.txz>
```

Example:
```
$ sudo jailconf-builder dl-base -s https://download.freebsd.org/releases/amd64/14.3-RELEASE/base.txz
```

### Create a Jail

To create a new jail:

```
$ sudo jailconf-builder create -name <jail_name> -version <FreeBSD_version>
```

Example:
```
$ sudo jailconf-builder create -name myjail -version 14.3-RELEASE
```

IP address and epair number are automatically assigned.

To list existing jails, use:
```
$ ls /etc/jail.conf.d/
```

### Delete a Jail

To delete a jail:

```
$ sudo jailconf-builder delete -name <jail_name>
```

This will prompt for confirmation. To skip the confirmation, use the `-f` flag:

```
$ sudo jailconf-builder delete -name <jail_name> -f
```

## License

This project is licensed under the [MIT License](./LICENSE).
