package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "0.4.0"

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

	case "preview":
		previewCmd := flag.NewFlagSet("preview", flag.ExitOnError)
		templateFlag := previewCmd.String("template", "", "Path to jail.conf template (required)")
		configFlag := previewCmd.String("config", "", "Path to jails.json (required)")
		nameFlag := previewCmd.String("name", "", "Jail name (optional, shows only specified jail)")

		previewCmd.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage: jailconf-builder preview -template <template_file> -config <config_file> [-name <jail_name>]\n\n")
			fmt.Fprintf(os.Stderr, "Options:\n")
			previewCmd.PrintDefaults()
		}

		if err := previewCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}

		if *templateFlag == "" || *configFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: -template and -config are required\n\n")
			previewCmd.Usage()
			os.Exit(1)
		}

		if err := runPreview(*templateFlag, *configFlag, *nameFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "create":
		createCmd := flag.NewFlagSet("create", flag.ExitOnError)
		templateFlag := createCmd.String("template", "", "Path to jail.conf template (required)")
		configFlag := createCmd.String("config", "", "Path to jails.json (required)")
		nameFlag := createCmd.String("name", "", "Jail name (optional, creates only specified jail)")

		createCmd.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage: jailconf-builder create -template <template_file> -config <config_file> [-name <jail_name>]\n\n")
			fmt.Fprintf(os.Stderr, "Options:\n")
			createCmd.PrintDefaults()
		}

		if err := createCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}

		if *templateFlag == "" || *configFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: -template and -config are required\n\n")
			createCmd.Usage()
			os.Exit(1)
		}

		if err := runCreate(*templateFlag, *configFlag, *nameFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "delete":
		deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
		templateFlag := deleteCmd.String("template", "", "Path to jail.conf template (required)")
		configFlag := deleteCmd.String("config", "", "Path to jails.json (required)")
		nameFlag := deleteCmd.String("name", "", "Jail name (optional, deletes only specified jail)")
		forceFlag := deleteCmd.Bool("f", false, "Force delete without confirmation")

		deleteCmd.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage: jailconf-builder delete -template <template_file> -config <config_file> [-name <jail_name>] [-f]\n\n")
			fmt.Fprintf(os.Stderr, "Options:\n")
			deleteCmd.PrintDefaults()
		}

		if err := deleteCmd.Parse(os.Args[2:]); err != nil {
			os.Exit(1)
		}

		if *templateFlag == "" || *configFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: -template and -config are required\n\n")
			deleteCmd.Usage()
			os.Exit(1)
		}

		if err := runDelete(*templateFlag, *configFlag, *nameFlag, *forceFlag); err != nil {
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
  preview     Preview generated jail.conf without creating
  create      Create jails from template and config
  delete      Delete jails specified in config
  dl-base     Download FreeBSD base system

Options:
  version     Show version
  help        Show this help message

For more information on each command, use:
  jailconf-builder <command> -h

`)
}
