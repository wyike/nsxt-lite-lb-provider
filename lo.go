package main

import (
	"fmt"
	"os/exec"
)

// SetLoopbackIP sets the specified IP address on the lo interface.
func SetLoopbackIP(ip string) error {
	// Command to add an IP address to the lo interface
	cmd := exec.Command("ip", "addr", "replace", ip, "dev", "lo")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set IP %s on lo interface: %v, output: %s", ip, err, string(output))
	}
	return nil
}

// DeleteLoopbackIP deletes the specified IP address from the lo interface.
func DeleteLoopbackIP(ip string) error {
	// Command to remove an IP address from the lo interface
	cmd := exec.Command("ip", "addr", "del", ip, "dev", "lo")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete IP %s from lo interface: %v, output: %s", ip, err, string(output))
	}
	return nil
}
