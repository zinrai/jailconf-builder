package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "0.2.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		if err := runInit(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "create":
		createCmd := flag.NewFlagSet("create", flag.ExitOnError)
		nameFlag := createCmd.String("name", "", "Jail name (required)")
		versionFlag := createCmd.String("version", "", "FreeBSD version (required, e.g., 14.1-RELEASE)")

		createCmd.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage: jailconf-builder create -name <jail_name> -version <freebsd_version>\n\n")
			fmt.Fprintf(os.Stderr, "Options:\n")
			createCmd.PrintDefaults()
			fmt.Fprintf(os.Stderr, "\nExample:\n")
			fmt.Fprintf(os.Stderr, "  jailconf-builder create -name myjail -version 14.1-RELEASE\n")
		}

		if err := createCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}

		if *nameFlag == "" || *versionFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: -name and -version are required\n\n")
			createCmd.Usage()
			os.Exit(1)
		}

		if err := runCreate(*nameFlag, *versionFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "delete":
		deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
		nameFlag := deleteCmd.String("name", "", "Jail name to delete (required)")
		forceFlag := deleteCmd.Bool("f", false, "Force delete without confirmation")

		deleteCmd.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage: jailconf-builder delete -name <jail_name> [-f]\n\n")
			fmt.Fprintf(os.Stderr, "Options:\n")
			deleteCmd.PrintDefaults()
			fmt.Fprintf(os.Stderr, "\nExample:\n")
			fmt.Fprintf(os.Stderr, "  jailconf-builder delete -name myjail\n")
			fmt.Fprintf(os.Stderr, "  jailconf-builder delete -name myjail -f\n")
		}

		if err := deleteCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}

		if *nameFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: -name is required\n\n")
			deleteCmd.Usage()
			os.Exit(1)
		}

		if err := runDelete(*nameFlag, *forceFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "dl-base":
		dlBaseCmd := flag.NewFlagSet("dl-base", flag.ExitOnError)
		sourceFlag := dlBaseCmd.String("s", "", "URL to base.txz (required)")

		dlBaseCmd.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage: jailconf-builder dl-base -s <url>\n\n")
			fmt.Fprintf(os.Stderr, "Options:\n")
			dlBaseCmd.PrintDefaults()
			fmt.Fprintf(os.Stderr, "\nExample:\n")
			fmt.Fprintf(os.Stderr, "  jailconf-builder dl-base -s https://download.freebsd.org/ftp/releases/amd64/14.1-RELEASE/base.txz\n")
		}

		if err := dlBaseCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}

		if *sourceFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: -s is required\n\n")
			dlBaseCmd.Usage()
			os.Exit(1)
		}

		if err := runDlBase(*sourceFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "version", "-v", "--version":
		fmt.Printf("jailconf-builder %s\n", version)

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `jailconf-builder - FreeBSD Jail Manager using standard jail.conf

Usage:
  jailconf-builder <command> [options]

Commands:
  init        Initialize jailconf-builder environment
  create      Create a new jail
  delete      Delete an existing jail
  dl-base     Download FreeBSD base system

Options:
  version     Show version
  help        Show this help message

For more information on each command, use:
  jailconf-builder <command> -h

`)
}
