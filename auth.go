package forwarder

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// SetKeyFile changes the authentication to key-based and uses the specified file.
// Leaving the file empty defaults to the default linux private key locations: `~/.ssh/id_rsa`, `~/.ssh/id_dsa`,
// `~/.ssh/id_ecdsa`, `~/.ssh/id_ecdsa_sk`, `~/.ssh/id_ed25519` and `~/.ssh/id_ed25519_sk`.
func (tun *ForwardConfig) SetKeyFile(file string) {
	tun.authKeyFile = file
}

// SetPort changes the port where the SSH connection will be made.
func (tun *ForwardConfig) SetPort(port int) {
	tun.Server.port = port
}

func (tun *ForwardConfig) readPrivateKey(keyFile string) (ssh.AuthMethod, error) {
	buf, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("reading SSH key file %s: %w", keyFile, err)
	}

	key, err := tun.parsePrivateKey(buf)
	if err != nil {
		return nil, fmt.Errorf("parsing SSH key file %s: %w", keyFile, err)
	}

	return key, nil
}

func (tun *ForwardConfig) parsePrivateKey(buf []byte) (ssh.AuthMethod, error) {
	var key ssh.Signer
	var err error
	key, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, fmt.Errorf("error parsing key: %w", err)
	}

	return ssh.PublicKeys(key), nil
}

func (tun *ForwardConfig) getSSHAuthMethodForKeyFile() (ssh.AuthMethod, error) {
	if tun.authKeyFile != "" {
		return tun.readPrivateKey(tun.authKeyFile)
	}

	return nil, fmt.Errorf("could not read  SSH key %v", tun.authKeyFile)
}

func (tun *ForwardConfig) getSSHAuthMethod() (ssh.AuthMethod, error) {
	return tun.getSSHAuthMethodForKeyFile()
}
