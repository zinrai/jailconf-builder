package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// runInit executes the init subcommand
func runInit() error {
	// 1. Check if /etc/jail.conf.d exists
	if _, err := os.Stat(JailConfDir); os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist", JailConfDir)
	}
	fmt.Printf("Confirmed directory exists: %s\n", JailConfDir)

	// 2. Add include directive to /etc/jail.conf
	if err := ensureInclude(); err != nil {
		return fmt.Errorf("failed to update %s: %w", MainJailConf, err)
	}

	// 3. Create directories
	dirs := []string{JailRootDir, BaseDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		fmt.Printf("Created directory: %s\n", dir)
	}

	fmt.Println("jailconf-builder initialized successfully.")
	return nil
}

// ensureInclude adds the include directive to /etc/jail.conf
func ensureInclude() error {
	// Create file if it doesn't exist
	if _, err := os.Stat(MainJailConf); os.IsNotExist(err) {
		content := "# FreeBSD jail configuration\n\n" + IncludeLine + "\n"
		if err := os.WriteFile(MainJailConf, []byte(content), 0644); err != nil {
			return err
		}
		fmt.Printf("Created %s with include directive.\n", MainJailConf)
		return nil
	}

	// Check if include directive already exists
	content, err := os.ReadFile(MainJailConf)
	if err != nil {
		return err
	}

	if strings.Contains(string(content), IncludeLine) {
		fmt.Printf("%s already contains include directive.\n", MainJailConf)
		return nil
	}

	// Append to end of file
	f, err := os.OpenFile(MainJailConf, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString("\n" + IncludeLine + "\n"); err != nil {
		return err
	}

	fmt.Printf("Added include directive to %s.\n", MainJailConf)
	return nil
}

// runCreate executes the create subcommand
func runCreate(name, version string) error {
	// 1. Check if base.txz exists
	basePath := filepath.Join(BaseDir, version, "base.txz")
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return fmt.Errorf("base.txz for version %s not found at %s\nPlease run 'jailconf-builder dl-base' first", version, basePath)
	}

	// 2. Check if jail already exists
	if jailExists(name) {
		return fmt.Errorf("jail '%s' already exists", name)
	}

	// 3. Get next available number
	num, err := getNextAvailableNumber()
	if err != nil {
		return fmt.Errorf("failed to get next available number: %w", err)
	}

	// 4. Create Jail instance
	jail := NewJail(name, num)

	// 5. Create jail root directory
	if err := os.MkdirAll(jail.RootPath(), 0755); err != nil {
		return fmt.Errorf("failed to create jail directory: %w", err)
	}

	// 6. Extract base.txz
	fmt.Printf("Extracting base system to %s...\n", jail.RootPath())
	cmd := exec.Command("tar", "-xf", basePath, "-C", jail.RootPath())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Clean up on failure
		os.RemoveAll(jail.RootPath())
		return fmt.Errorf("failed to extract base system: %w", err)
	}

	// 7. Generate jail.conf
	if err := jail.WriteConf(); err != nil {
		// Clean up on failure
		os.RemoveAll(jail.RootPath())
		return fmt.Errorf("failed to write jail.conf: %w", err)
	}

	fmt.Printf("Jail '%s' created successfully.\n", name)
	fmt.Printf("  Config: %s\n", jail.ConfPath())
	fmt.Printf("  Root:   %s\n", jail.RootPath())
	fmt.Printf("  IP:     %s\n", jail.IPAddr)
	fmt.Printf("\nTo start the jail:\n")
	fmt.Printf("  sudo jail -c %s\n", name)
	fmt.Printf("  # or\n")
	fmt.Printf("  sudo service jail start %s\n", name)

	return nil
}

// runDelete executes the delete subcommand
func runDelete(name string, force bool) error {
	// 1. Find config file
	confPath, err := findJailConf(name)
	if err != nil {
		return fmt.Errorf("jail '%s' not found", name)
	}

	jailPath := filepath.Join(JailRootDir, name)

	// 2. Confirmation prompt
	if !force {
		fmt.Printf("Are you sure you want to delete jail '%s'? This action cannot be undone. [y/N]: ", name)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Jail deletion cancelled.")
			return nil
		}
	}

	// 3. Delete config file
	if err := os.Remove(confPath); err != nil {
		return fmt.Errorf("failed to delete config file: %w", err)
	}
	fmt.Printf("Deleted: %s\n", confPath)

	// 4. Delete jail directory
	if _, err := os.Stat(jailPath); err == nil {
		// Remove schg flags to allow deletion
		chflagsCmd := exec.Command("chflags", "-R", "noschg,nouchg", jailPath)
		if err := chflagsCmd.Run(); err != nil {
			fmt.Printf("Warning: failed to remove flags: %v\n", err)
		}

		if err := os.RemoveAll(jailPath); err != nil {
			return fmt.Errorf("failed to delete jail directory: %w", err)
		}
		fmt.Printf("Deleted: %s\n", jailPath)
	}

	fmt.Printf("Jail '%s' has been successfully deleted.\n", name)
	return nil
}

// runDlBase executes the dl-base subcommand
func runDlBase(source string) error {
	// 1. Extract version from URL
	version, err := extractVersionFromURL(source)
	if err != nil {
		return fmt.Errorf("failed to extract version from URL: %w", err)
	}

	// 2. Download (check status before creating directory)
	fmt.Printf("Downloading base.txz from %s...\n", source)
	resp, err := http.Get(source)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// 3. Create destination directory (only after successful response)
	destDir := filepath.Join(BaseDir, version)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	destFile := filepath.Join(destDir, "base.txz")

	// 4. Save file
	out, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		// Clean up on failure
		os.Remove(destFile)
		os.Remove(destDir)
		return fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Printf("Base system for FreeBSD %s downloaded successfully to %s\n", version, destFile)
	return nil
}

// extractVersionFromURL extracts the FreeBSD version from a download URL
func extractVersionFromURL(url string) (string, error) {
	// URL format: https://download.freebsd.org/.../14.1-RELEASE/base.txz
	path := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")

	dir := filepath.Dir(path)
	version := filepath.Base(dir)

	if !strings.HasSuffix(version, "-RELEASE") {
		return "", fmt.Errorf("invalid version format: %s (expected *-RELEASE)", version)
	}

	return version, nil
}
