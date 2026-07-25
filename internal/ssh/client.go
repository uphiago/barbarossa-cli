// Package ssh provides an SSH client interface for connecting to worker
// containers, along with real and mock implementations.
package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

// ─── Types ───────────────────────────────────────────────────

// SessionResult holds the output from a remote command execution.
type SessionResult struct {
	Stdout string
	Stderr string
	Err    error
}

// Client defines the interface for SSH operations on workers.
type Client interface {
	// Connect establishes an SSH connection to the given host as user,
	// authenticating with the key at keyPath.
	Connect(host, user, keyPath string) error

	// Run executes a command on the remote host and returns the result.
	Run(ctx context.Context, cmd string) *SessionResult

	// NewSession creates an interactive session with a PTY.
	NewSession() (io.ReadWriteCloser, error)

	// Close terminates the SSH connection.
	Close() error
}

// ─── Real Client ────────────────────────────────────────────

// RealClient connects to a real SSH server.
type RealClient struct {
	conn *gossh.Client
}

// NewRealClient creates a new RealClient (not yet connected).
func NewRealClient() *RealClient {
	return &RealClient{}
}

// Connect establishes an SSH connection.
func (c *RealClient) Connect(host, user, keyPath string) error {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("ssh read key %q: %w", keyPath, err)
	}

	signer, err := gossh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("ssh parse key %q: %w", keyPath, err)
	}

	config := &gossh.ClientConfig{
		User:            user,
		Auth:            []gossh.AuthMethod{gossh.PublicKeys(signer)},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), // dev/container context
		Timeout:         5 * time.Second,
	}

	addr := net.JoinHostPort(host, "22")
	conn, err := gossh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("ssh dial %s@%s: %w", user, host, err)
	}

	c.conn = conn
	return nil
}

// Run executes a command and collects stdout+stderr.
func (c *RealClient) Run(ctx context.Context, cmd string) *SessionResult {
	if c.conn == nil {
		return &SessionResult{Err: fmt.Errorf("not connected")}
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return &SessionResult{Err: fmt.Errorf("ssh session: %w", err)}
	}
	defer session.Close()

	done := make(chan *SessionResult, 1)
	go func() {
		stdout, err := session.Output(cmd)
		done <- &SessionResult{
			Stdout: string(stdout),
			Err:    err,
		}
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		session.Close()
		return &SessionResult{Err: ctx.Err()}
	}
}

// NewSession creates an interactive PTY session wrapped as ReadWriteCloser.
func (c *RealClient) NewSession() (io.ReadWriteCloser, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return nil, fmt.Errorf("ssh session: %w", err)
	}

	// Request PTY
	modes := gossh.TerminalModes{
		gossh.ECHO:          1,
		gossh.TTY_OP_ISPEED: 14400,
		gossh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", 80, 24, modes); err != nil {
		session.Close()
		return nil, fmt.Errorf("ssh pty: %w", err)
	}

	// Two pipes: one for stdin, one for stdout+stderr
	stdinR, stdinW := io.Pipe()   // we write to stdinW, session reads from stdinR
	stdoutR, stdoutW := io.Pipe() // session writes to stdoutW, we read from stdoutR

	session.Stdin = stdinR
	session.Stdout = stdoutW
	session.Stderr = stdoutW

	return &sshSessionWrapper{
		session: session,
		Reader:  stdoutR,
		Writer:  stdinW,
	}, nil
}

// Close closes the SSH connection.
func (c *RealClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// sshSessionWrapper wraps an SSH session as an io.ReadWriteCloser.
type sshSessionWrapper struct {
	session *gossh.Session
	io.Reader
	io.Writer
}

func (w *sshSessionWrapper) Close() error {
	return w.session.Close()
}

// ─── Mock Client ────────────────────────────────────────────

// MockClient returns predefined results for testing.
type MockClient struct {
	Connected bool
	Output    string
	Err       error
}

// NewMockClient creates a MockClient ready to return mock data.
func NewMockClient() *MockClient {
	return &MockClient{Connected: true}
}

func (c *MockClient) Connect(host, user, keyPath string) error {
	c.Connected = true
	return c.Err
}

func (c *MockClient) Run(ctx context.Context, cmd string) *SessionResult {
	return &SessionResult{
		Stdout: c.Output,
		Err:    c.Err,
	}
}

func (c *MockClient) NewSession() (io.ReadWriteCloser, error) {
	return nil, fmt.Errorf("mock: interactive session not implemented")
}

func (c *MockClient) Close() error {
	c.Connected = false
	return nil
}
