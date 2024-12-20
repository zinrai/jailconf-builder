package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

const (
	JailconfBuilderRoot = "/opt/jailconf-builder"
	BaseDir             = JailconfBuilderRoot + "/base"
	JailConfDir         = JailconfBuilderRoot + "/jail.conf.d"
	JailsDir            = JailconfBuilderRoot + "/jails"
)

type Jail struct {
	Name    string
	IPAddr  string
	Gateway string
	Version string
	If      int
}

var rootCmd = &cobra.Command{
	Use:   "jailconf-builder",
	Short: "jailconf-builder - FreeBSD Standard Jail Manager",
	Long:  `A CLI tool to manage FreeBSD jails using the standard jail.conf.`,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new jail",
	Run:   createJail,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all jails",
	Run:   listJails,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a jail",
	Run:   deleteJail,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize jailconf-builder environment",
	Run:   initVanillaJail,
}

var dlBaseCmd = &cobra.Command{
	Use:   "dl-base",
	Short: "Download FreeBSD base system for jails",
	Run:   downloadBase,
}

func init() {
	rootCmd.AddCommand(createCmd, listCmd, deleteCmd, initCmd, dlBaseCmd)
	createCmd.Flags().StringP("version", "v", "", "FreeBSD version for the jail")
	createCmd.Flags().StringP("ip", "i", "", "IP address for the jail")
	createCmd.Flags().StringP("gw", "g", "", "Default gateway for the jail")
	createCmd.MarkFlagRequired("version")
	createCmd.MarkFlagRequired("ip")
	createCmd.MarkFlagRequired("gw")
	dlBaseCmd.Flags().StringP("source", "s", "", "URL to base.txz")
	deleteCmd.Flags().BoolP("force", "f", false, "Force delete without confirmation")
}

func checkInitialized() error {
	dirsToCheck := []string{JailconfBuilderRoot, BaseDir, JailConfDir, JailsDir}
	for _, dir := range dirsToCheck {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("directory %s does not exist. Please run 'jailconf-builder init' first", dir)
		}
	}
	return nil
}

func getNextAvailableNumber() (int, error) {
	files, err := os.ReadDir(JailConfDir)
	if err != nil {
		return 0, fmt.Errorf("error reading jail.conf.d directory: %v", err)
	}

	usedNumbers := make([]int, 0)
	maxNumber := 0

	for _, file := range files {
		parts := strings.SplitN(file.Name(), "-", 2)
		if len(parts) == 2 {
			if num, err := strconv.Atoi(parts[0]); err == nil {
				usedNumbers = append(usedNumbers, num)
				if num > maxNumber {
					maxNumber = num
				}
			}
		}
	}

	if len(usedNumbers) == 0 {
		return 1, nil
	}

	sort.Ints(usedNumbers)

	allNumbers := make([]int, maxNumber)
	for i := range allNumbers {
		allNumbers[i] = i + 1
	}

	for _, num := range allNumbers {
		if !contains(usedNumbers, num) {
			return num, nil
		}
	}

	return maxNumber + 1, nil
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func checkJailNameExists(name string) bool {
	files, err := os.ReadDir(JailConfDir)
	if err != nil {
		return false
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), "-"+name+".conf") {
			return true
		}
	}

	return false
}

func createJail(cmd *cobra.Command, args []string) {
	if err := checkInitialized(); err != nil {
		fmt.Println(err)
		return
	}

	if len(args) != 1 {
		fmt.Println("Error: Please provide a jail name")
		return
	}
	jailName := args[0]

	if checkJailNameExists(jailName) {
		fmt.Printf("Error: Jail with name '%s' already exists\n", jailName)
		return
	}

	version, _ := cmd.Flags().GetString("version")
	ipAddr, _ := cmd.Flags().GetString("ip")
	gateway, _ := cmd.Flags().GetString("gw")

	nextNum, err := getNextAvailableNumber()
	if err != nil {
		fmt.Printf("Error getting next available number: %v\n", err)
		return
	}

	jail := Jail{
		Name:    jailName,
		IPAddr:  ipAddr,
		Gateway: gateway,
		Version: version,
		If:      nextNum,
	}

	if err := createJailEnvironment(jail); err != nil {
		fmt.Printf("Error creating jail environment: %v\n", err)
		return
	}

	if err := generateJailConf(jail); err != nil {
		fmt.Printf("Error creating jail configuration: %v\n", err)
		return
	}

	fmt.Printf("Jail '%s' created successfully with interface number %d.\n", jailName, nextNum)
}

func createJailEnvironment(jail Jail) error {
	baseFile := filepath.Join(BaseDir, jail.Version, "base.txz")
	if _, err := os.Stat(baseFile); os.IsNotExist(err) {
		return fmt.Errorf("base.txz for version %s not found. Please run 'jailconf-builder dl-base' first", jail.Version)
	}

	jailDirName := fmt.Sprintf("%d-%s", jail.If, jail.Name)
	jailPath := filepath.Join(JailsDir, jailDirName)
	if err := os.MkdirAll(jailPath, 0755); err != nil {
		return err
	}

	fmt.Printf("Extracting base system to %s...\n", jailPath)
	cmd := exec.Command("tar", "-xvf", baseFile, "-C", jailPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error extracting base system: %v", err)
	}

	return nil
}

func generateJailConf(jail Jail) error {
	tmpl := `
{{ .Name }} {
	$if = {{ .If }};
	$ip4_addr = {{ .IPAddr }};
	$gw = {{ .Gateway }};
}
`

	confPath := filepath.Join(JailConfDir, fmt.Sprintf("%d-%s.conf", jail.If, jail.Name))
	f, err := os.Create(confPath)
	if err != nil {
		return err
	}
	defer f.Close()

	t := template.Must(template.New("jail").Parse(tmpl))
	return t.Execute(f, jail)
}

func listJails(cmd *cobra.Command, args []string) {
	files, err := os.ReadDir(JailConfDir)
	if err != nil {
		fmt.Printf("Error reading jail configurations: %v\n", err)
		return
	}

	fmt.Println("Available jails:")
	for _, file := range files {
		parts := strings.SplitN(file.Name(), "-", 2)
		if len(parts) == 2 {
			jailName := strings.TrimSuffix(parts[1], ".conf")
			fmt.Printf("- %s (interface: epair%s, path: %s)\n", jailName, parts[0], filepath.Join(JailsDir, file.Name()[:len(file.Name())-5]))
		}
	}
}

func deleteJail(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("Error: Please provide a jail name to delete")
		return
	}

	jailName := args[0]
	force, _ := cmd.Flags().GetBool("force")

	// Find the jail configuration file
	files, err := os.ReadDir(JailConfDir)
	if err != nil {
		fmt.Printf("Error reading jail configurations: %v\n", err)
		return
	}

	var confFile string
	var jailDir string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), "-"+jailName+".conf") {
			confFile = file.Name()
			jailDir = file.Name()[:len(file.Name())-5] // Remove .conf
			break
		}
	}

	if confFile == "" {
		fmt.Printf("Error: Jail '%s' not found\n", jailName)
		return
	}

	if !force {
		fmt.Printf("Are you sure you want to delete jail '%s'? This action cannot be undone. [y/N]: ", jailName)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Jail deletion cancelled.")
			return
		}
	}

	// Delete jail configuration file
	confPath := filepath.Join(JailConfDir, confFile)
	if err := os.Remove(confPath); err != nil {
		fmt.Printf("Error deleting jail configuration file: %v\n", err)
		return
	}

	// Remove flags from jail directory
	jailPath := filepath.Join(JailsDir, jailDir)
	chflagsCmd := exec.Command("chflags", "-R", "noschg,nouchg", jailPath)
	if err := chflagsCmd.Run(); err != nil {
		fmt.Printf("Error removing flags from jail directory: %v\n", err)
		return
	}

	// Delete jail directory
	if err := os.RemoveAll(jailPath); err != nil {
		fmt.Printf("Error deleting jail directory: %v\n", err)
		return
	}

	fmt.Printf("Jail '%s' has been successfully deleted.\n", jailName)
}

func initVanillaJail(cmd *cobra.Command, args []string) {
	if err := createMainJailConf(); err != nil {
		fmt.Printf("Error creating main jail.conf: %v\n", err)
		return
	}

	dirsToCreate := []string{JailconfBuilderRoot, BaseDir, JailConfDir, JailsDir}
	for _, dir := range dirsToCreate {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}
		fmt.Printf("Created directory: %s\n", dir)
	}

	fmt.Println("jailconf-builder environment initialized successfully.")
}

func createMainJailConf() error {
	mainConfPath := "/etc/jail.conf"
	if _, err := os.Stat(mainConfPath); err == nil {
		fmt.Println("jail.conf already exists. Skipping creation.")
		return nil
	}

	mainConf := `
exec.prestart  = "ifconfig epair${if} create up > /dev/null";
exec.prestart += "ifconfig bridge0 addm epair${if}a";
exec.start     = "ifconfig lo0 up 127.0.0.1";
exec.start    += "ifconfig epair${if}b up ${ip4_addr}";
exec.start    += "route add default ${gw}";
exec.start    += "sh /etc/rc";
exec.stop      = "sh /etc/rc.shutdown";
exec.poststop  = "ifconfig epair${if}a destroy";
host.hostname  = "${name}.jail";
mount.devfs;
devfs_ruleset  = 5;
vnet;
vnet.interface = "epair${if}b";
path           = "` + JailsDir + `/${if}-${name}";
persist;

.include "` + JailConfDir + `/*";
`

	if err := os.WriteFile(mainConfPath, []byte(mainConf), 0644); err != nil {
		return fmt.Errorf("error creating main jail.conf: %v", err)
	}

	fmt.Println("Main jail.conf created successfully.")
	return nil
}

func extractVersionFromURL(url string) (string, error) {
	path := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	path = strings.TrimPrefix(path, "download.freebsd.org/")

	dir := filepath.Dir(path)
	version := filepath.Base(dir)

	if !strings.HasSuffix(version, "-RELEASE") {
		return "", fmt.Errorf("invalid version format: %s", version)
	}

	return version, nil
}

func downloadBase(cmd *cobra.Command, args []string) {
	if err := checkInitialized(); err != nil {
		fmt.Println(err)
		return
	}

	source, _ := cmd.Flags().GetString("source")

	if source == "" {
		fmt.Println("Please specify the URL to base.txz using the --source flag")
		return
	}

	version, err := extractVersionFromURL(source)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	destDir := filepath.Join(BaseDir, version)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Printf("Error creating destination directory: %v\n", err)
		return
	}

	destFile := filepath.Join(destDir, "base.txz")

	fmt.Printf("Downloading base.txz from %s...\n", source)
	resp, err := http.Get(source)
	if err != nil {
		fmt.Printf("Error downloading base.txz: %v\n", err)
		return
	}
	defer resp.Body.Close()

	out, err := os.Create(destFile)
	if err != nil {
		fmt.Printf("Error creating destination file: %v\n", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Error saving base.txz: %v\n", err)
		return
	}

	fmt.Printf("Base system for FreeBSD %s downloaded successfully to %s\n", version, destFile)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
