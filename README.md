# Vanilla Jail

Vanilla Jail is a CLI tool for managing FreeBSD jails using the standard [jail.conf(5)](https://man.freebsd.org/cgi/man.cgi?jail.conf(5)) configuration.

Back around 2013, when I was working with jail.conf(5) on FreeBSD 9.x, I wished there was an include option to improve usability. Recently, I revisited jail.conf(5) and discovered that the include option is indeed implemented. This discovery led to the development of this `vanilla-jail`, which uses the standard features of the jail system.

## Features

- Initialize the Vanilla Jail environment
- Create new jails
- List existing jails
- Delete jails with safety checks
- Download FreeBSD base system for jails

## Installation

Build the tool:

```
$ GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o vanilla-jail
```

## Usage

### Initialize Vanilla Jail

Before using Vanilla Jail, you need to initialize the environment:

```
$ sudo vanilla-jail init
```

This command creates the necessary directories and the main jail.conf file.

### Create a Jail

To create a new jail:

```
$ sudo vanilla-jail create <jail_name> -v <FreeBSD_version> -i <IP_address> -g <gateway>
```

Example:
```
$ sudo vanilla-jail create myjail -v 14.1-RELEASE -i 192.168.2.100 -g 192.168.2.1
```

### List Jails

To list all existing jails:

```
$ vanilla-jail list
```

### Delete a Jail

To delete a jail:

```
$ sudo vanilla-jail delete <jail_name>
```

This will prompt for confirmation. To skip the confirmation, use the `-f` flag:

```
$ sudo vanilla-jail delete -f <jail_name>
```

### Download FreeBSD Base System

To download the FreeBSD base system for a specific version:

```
$ sudo vanilla-jail dl-base -s <URL_to_base.txz>
```

Example:
```
$ sudo vanilla-jail dl-base -s https://download.freebsd.org/ftp/releases/amd64/14.1-RELEASE/base.txz
```

## License

This project is licensed under the MIT License - see the [LICENSE](https://opensource.org/license/mit) for details.
