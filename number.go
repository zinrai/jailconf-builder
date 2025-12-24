package main

import (
	"os"
	"sort"
	"strconv"
	"strings"
)

// getNextAvailableNumber returns the next available epair number
// by scanning files in /etc/jail.conf.d/.
// It fills gaps in numbering if any, otherwise returns max+1.
func getNextAvailableNumber() (int, error) {
	files, err := os.ReadDir(JailConfDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}

	usedNumbers := make([]int, 0)

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if !strings.HasSuffix(name, ".conf") {
			continue
		}

		// Filename format: {num}-{jailname}.conf
		parts := strings.SplitN(name, "-", 2)
		if len(parts) != 2 {
			continue
		}

		num, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		usedNumbers = append(usedNumbers, num)
	}

	if len(usedNumbers) == 0 {
		return 1, nil
	}

	// Sort and find gaps
	sort.Ints(usedNumbers)

	// Find the first gap
	for i, num := range usedNumbers {
		expected := i + 1
		if num != expected {
			return expected, nil
		}
	}

	// No gaps found, return max+1
	return usedNumbers[len(usedNumbers)-1] + 1, nil
}

// jailExists checks if a jail with the given name already exists
func jailExists(name string) bool {
	files, err := os.ReadDir(JailConfDir)
	if err != nil {
		return false
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".conf") {
			continue
		}

		// Filename format: {num}-{jailname}.conf
		parts := strings.SplitN(fileName, "-", 2)
		if len(parts) != 2 {
			continue
		}

		jailName := strings.TrimSuffix(parts[1], ".conf")
		if jailName == name {
			return true
		}
	}

	return false
}

// findJailConf searches for a jail config file by jail name
func findJailConf(name string) (string, error) {
	files, err := os.ReadDir(JailConfDir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".conf") {
			continue
		}

		// Filename format: {num}-{jailname}.conf
		parts := strings.SplitN(fileName, "-", 2)
		if len(parts) != 2 {
			continue
		}

		jailName := strings.TrimSuffix(parts[1], ".conf")
		if jailName == name {
			return JailConfDir + "/" + fileName, nil
		}
	}

	return "", os.ErrNotExist
}
