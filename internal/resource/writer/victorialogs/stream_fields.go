package victorialogs

import (
	"net"
	"os"
	"path/filepath"
	"strings"
)

// StreamFields represents the VictoriaLogs stream identification fields
type StreamFields struct {
	AppName  string `json:"app_name"`
	Hostname string `json:"hostname"`
	RemoteIP string `json:"remote_ip"`
}

// DetectStreamFields automatically detects the stream fields for VictoriaLogs
func DetectStreamFields() StreamFields {
	return StreamFields{
		AppName:  detectAppName(),
		Hostname: detectHostname(),
		RemoteIP: detectRemoteIP(),
	}
}

// detectAppName gets the application name from the executable name
func detectAppName() string {
	// Get the executable path
	execPath, err := os.Executable()
	if err != nil {
		// Fallback to program arguments
		if len(os.Args) > 0 {
			return filepath.Base(os.Args[0])
		}
		return "ups-metrics" // Default fallback
	}

	// Extract just the filename without extension
	appName := filepath.Base(execPath)

	// Remove common executable extensions
	if strings.HasSuffix(appName, ".exe") {
		appName = strings.TrimSuffix(appName, ".exe")
	}

	return appName
}

// detectHostname gets the system hostname
func detectHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown-host" // Fallback
	}
	return hostname
}

// detectRemoteIP gets the primary network interface IP address
func detectRemoteIP() string {
	// Try to get the IP by connecting to a remote address (doesn't actually connect)
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Fallback: get first non-loopback interface
		return getFirstNonLoopbackIP()
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// getFirstNonLoopbackIP gets the first non-loopback IP address
func getFirstNonLoopbackIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1" // Fallback to localhost
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip loopback and IPv6 addresses for simplicity
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			return ip.String()
		}
	}

	return "127.0.0.1" // Final fallback
}
