package scanner

import (
	"fmt"
	"os/exec"
	"strings"
)

type Scanner struct {
	dbPath string
}

func New(dbPath string) *Scanner {
	return &Scanner{dbPath: dbPath}
}

func (s *Scanner) ScanPackageManagers(managers []string) (map[string]int, error) {
	results := make(map[string]int)

	for _, manager := range managers {
		count, err := s.scanManager(manager)
		if err != nil {
			// Don't fail entire scan if one manager fails
			results[manager] = 0
			continue
		}
		results[manager] = count
	}

	return results, nil
}

func (s *Scanner) scanManager(manager string) (int, error) {
	switch manager {
	case "homebrew":
		return s.scanHomebrew()
	case "npm":
		return s.scanNpm()
	case "pip":
		return s.scanPip()
	case "cargo":
		return s.scanCargo()
	case "go":
		return s.scanGo()
	default:
		return 0, fmt.Errorf("unknown package manager: %s", manager)
	}
}

func (s *Scanner) scanHomebrew() (int, error) {
	cmd := exec.Command("brew", "list", "--formula")
	output, err := cmd.Output()
	if err != nil {
		return 0, nil // Homebrew might not be installed
	}

	packages := strings.Split(strings.TrimSpace(string(output)), "\n")
	return len(packages), nil
}

func (s *Scanner) scanNpm() (int, error) {
	cmd := exec.Command("npm", "list", "-g", "--depth=0", "--json")
	output, err := cmd.Output()
	if err != nil {
		return 0, nil
	}

	// Simple count - in real implementation, parse JSON
	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		if strings.Contains(line, `"dependencies"`) {
			count++
		}
	}

	return count, nil
}

func (s *Scanner) scanPip() (int, error) {
	cmd := exec.Command("pip3", "list", "--format=json")
	output, err := cmd.Output()
	if err != nil {
		return 0, nil
	}

	// Simple count
	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		if strings.Contains(line, `"name"`) {
			count++
		}
	}

	return count, nil
}

func (s *Scanner) scanCargo() (int, error) {
	// Cargo doesn't have a global list command
	// Would need to scan ~/.cargo/bin
	return 0, nil
}

func (s *Scanner) scanGo() (int, error) {
	// Go doesn't maintain a global registry
	// Would need to scan ~/go/bin
	return 0, nil
}
