# jailconf-builder

`jailconf-builder` is a CLI tool for managing FreeBSD jails using the standard [jail.conf(5)](https://man.freebsd.org/cgi/man.cgi?jail.conf(5)) configuration.

Back around 2013, when I was working with jail.conf(5) on FreeBSD 9.x, I wished there was an include option to improve usability. Recently, I revisited jail.conf(5) and discovered that the include option is indeed implemented. This discovery led to the development of this `jailconf-builder`, which uses the standard features of the jail system.

## Note

This is a CLI tool for creating jail environments using jail.conf(5). For jail operations, please use commands such as `jail -c <jail_name>` , `jail -r <jail_name>` , `jls` , and `jexec 1 /bin/tcsh`.

## Features

- Initialize `jailconf-builder` environment
- Preview generated jail.conf before creating
- Create jails with VNET support using Go Template and JSON configuration
- Delete jails with template matching safety checks
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
# cat << EOF >> /etc/rc.conf.d/pf
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

## Configuration Files

`jailconf-builder` uses Go Template and JSON files to generate jail.conf.

### Template File (jail.conf.tmpl)

A Go Template file that defines the jail.conf structure:

```
{{.name}} {
    host.hostname = "{{.name}}.jail";
    path = "/var/jails/{{.name}}";

    vnet;
    vnet.interface = "epair{{.number}}b";

    $ip4_addr = "{{.ip_addr}}";
    $gw = "{{.gateway}}";

    exec.prestart  = "ifconfig epair{{.number}} create up";
    exec.prestart += "ifconfig bridge0 addm epair{{.number}}a";
    exec.start     = "ifconfig lo0 up 127.0.0.1";
    exec.start    += "ifconfig epair{{.number}}b up $ip4_addr";
    exec.start    += "route add default $gw";
    exec.start    += "sh /etc/rc";
    exec.stop      = "sh /etc/rc.shutdown";
    exec.poststop  = "ifconfig epair{{.number}}a destroy";

    mount.devfs;
    devfs_ruleset = 5;
    persist;
}
```

### JSON File (jails.json)

A JSON file that defines jail parameters:

```json
{
  "jails": [
    {
      "name": "myjail",
      "number": 1,
      "version": "14.3-RELEASE",
      "ip_addr": "192.168.2.11",
      "gateway": "192.168.2.1"
    }
  ]
}
```

Required fields:
- `name`: Jail name
- `number`: epair number (used for network interface)
- `version`: FreeBSD version (must match downloaded base.txz)

Additional fields can be added and referenced in the template.

### Examples

The `examples/` directory contains sample configurations:

- `examples/standard/`: Basic single jail configuration
- `examples/custom/`: Advanced configuration with multiple jails and conditional options

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

### Preview Jails

To preview generated jail.conf without creating jails:

```
$ jailconf-builder preview -template <template_file> -config <config_file> [-name <jail_name>]
```

Preview all jails defined in config:
```
$ jailconf-builder preview -template examples/standard/jail.conf.tmpl -config examples/standard/jails.json
```

Preview a specific jail:
```
$ jailconf-builder preview -template examples/standard/jail.conf.tmpl -config examples/standard/jails.json -name myjail
```

### Create Jails

To create jails from template and config:

```
$ sudo jailconf-builder create -template <template_file> -config <config_file> [-name <jail_name>]
```

Create all jails defined in config:
```
$ sudo jailconf-builder create -template examples/standard/jail.conf.tmpl -config examples/standard/jails.json
```

Create a specific jail:
```
$ sudo jailconf-builder create -template examples/standard/jail.conf.tmpl -config examples/standard/jails.json -name myjail
```

To list existing jails, use:
```
$ ls /etc/jail.conf.d/
```

### Delete Jails

To delete jails:

```
$ sudo jailconf-builder delete -template <template_file> -config <config_file> [-name <jail_name>] [-f]
```

Delete all jails defined in config:
```
$ sudo jailconf-builder delete -template examples/standard/jail.conf.tmpl -config examples/standard/jails.json
```

Delete a specific jail:
```
$ sudo jailconf-builder delete -template examples/standard/jail.conf.tmpl -config examples/standard/jails.json -name myjail
```

Skip confirmation prompt:
```
$ sudo jailconf-builder delete -template examples/standard/jail.conf.tmpl -config examples/standard/jails.json -f
```

Note: The delete command compares the existing jail.conf with the template output. If they differ, deletion is refused to prevent accidental removal of manually modified configurations.

## License

This project is licensed under the [MIT License](./LICENSE).
