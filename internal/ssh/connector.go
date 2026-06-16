// Package ssh provides SSH connectivity to target servers.
package ssh

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// Connector manages SSH connections to remote servers.
type Connector struct {
	timeout time.Duration
}

// NewConnector creates an SSH connector with a 10-second dial timeout.
func NewConnector() *Connector {
	return &Connector{timeout: 10 * time.Second}
}

// Connect establishes an SSH connection to the target server.
func (c *Connector) Connect(host string, port int, authType, credential string) (*ssh.Client, error) {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	var auth ssh.AuthMethod
	switch authType {
	case "password":
		auth = ssh.Password(credential)
	case "key":
		keyData, err := os.ReadFile(credential)
		if err != nil {
			// try credential as raw key content
			keyData = []byte(credential)
		}
		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse SSH key: %w", err)
		}
		auth = ssh.PublicKeys(signer)
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", authType)
	}

	config := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{auth, c.keyboardInteractive(credential)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         c.timeout,
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("SSH dial failed: %w", err)
	}
	return client, nil
}

// Run executes a command on the remote server and returns stdout.
func (c *Connector) Run(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("session create failed: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		return stderr.String(), fmt.Errorf("command failed: %w\nstderr: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

// TestConnection tests SSH connectivity and returns the server hostname.
func (c *Connector) TestConnection(host string, port int, authType, credential string) (map[string]string, error) {
	client, err := c.Connect(host, port, authType, credential)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	hostname, _ := c.Run(client, "hostname")
	arch, _ := c.Run(client, "uname -m")
	kernel, _ := c.Run(client, "uname -r")

	return map[string]string{
		"hostname": trim(hostname),
		"arch":     trim(arch),
		"kernel":   trim(kernel),
	}, nil
}

func (c *Connector) keyboardInteractive(password string) ssh.AuthMethod {
	return ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
		answers := make([]string, len(questions))
		for i := range questions {
			answers[i] = password
		}
		return answers, nil
	})
}

func trim(s string) string {
	return string(bytes.TrimRight([]byte(s), "\n\r\t "))
}
