package netrasp

import (
	"context"
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"
)

// sshConnection contains configuration and connection information for SSH.
type sshConnection struct {
	Config  *ssh.ClientConfig
	Host    *host
	reader  io.Reader
	writer  io.Writer
	session *ssh.Session
}

// Dial opens an SSH connection.
func (s *sshConnection) Dial(ctx context.Context) error {

	// Older cisco not using new crypto (Check rfc9142)
	s.Config.MACs = append(s.Config.MACs, "hmac-sha2-256", "hmac-sha2-512")
	s.Config.Ciphers = append(s.Config.Ciphers, "aes128-cbc", "aes192-cbc", "aes256-cbc")
	s.Config.KeyExchanges = append(
		s.Config.KeyExchanges,
		"diffie-hellman-group14-sha256",
		"diffie-hellman-group-exchange-sha256",
		"diffie-hellman-group14-sha1",
		"diffie-hellman-group-exchange-sha1",
		"diffie-hellman-group1-sha1",
	)

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", s.Host.Address, s.Host.Port), s.Config)
	if err != nil {
		return fmt.Errorf("unable to establish connection: %w", err)
	}

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("unable to open new session: %w", err)
	}

	terminalMode := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 28800,
		ssh.TTY_OP_OSPEED: 28800,
	}
	err = session.RequestPty("xterm", 80, 40, terminalMode)
	if err != nil {
		return fmt.Errorf("error requesting pty terminal: %w", err)
	}

	s.reader, err = session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error requesting StdoutPipe: %w", err)
	}
	s.writer, err = session.StdinPipe()
	if err != nil {
		return fmt.Errorf("error requesting StdinPipe: %w", err)
	}

	err = session.Shell()
	if err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	s.session = session

	return nil
}

// GetHost returns information about the connected host.
func (s *sshConnection) GetHost() *host {
	return s.Host
}

// Close disconnects from the device.
func (s *sshConnection) Close(ctx context.Context) error {
	s.session.Close()

	return nil
}

// Send is used to write commands to the device.
func (s *sshConnection) Send(ctx context.Context, command string) error {
	_, err := s.writer.Write([]byte(command + "\n"))
	if err != nil {
		return fmt.Errorf("unable to send command to device: %w", err)
	}

	return nil
}

// Recv is used to read data from the device.
func (s *sshConnection) Recv(ctx context.Context) io.Reader {
	return newContextReader(ctx, s.reader)
}
