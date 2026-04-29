package ssh

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cristiangonsevi/orbit/internal/config"
)

const (
	scpBufferSize = 32 * 1024 // 32KB buffer for SCP transfers
)

// UploadFile copies a local file or directory to a remote destination via SCP over SSH.
func (c *Client) UploadFile(entry config.UploadEntry, verbose bool) error {
	source := entry.Source

	// Expand ~ in source path
	if strings.HasPrefix(source, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home directory: %w", err)
		}
		source = filepath.Join(home, source[2:])
	}

	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("stat source %q: %w", source, err)
	}

	if info.IsDir() {
		return c.uploadDir(source, entry.Destination, verbose)
	}
	return c.uploadFile(source, entry.Destination, info.Mode(), verbose)
}

// uploadFile copies a single file to the remote host using SCP protocol.
func (c *Client) uploadFile(localPath, remotePath string, mode os.FileMode, verbose bool) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("creating SCP session: %w", err)
	}
	defer session.Close()

	// Open local file
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("opening local file %q: %w", localPath, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stating local file %q: %w", localPath, err)
	}
	fileSize := info.Size()

	if verbose {
		fmt.Fprintf(os.Stderr, "[SCP] Uploading %s (%d bytes) → %s\n", localPath, fileSize, remotePath)
	}

	// Get stdin/stdout pipes
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("getting SCP stdin pipe: %w", err)
	}
	defer stdin.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("getting SCP stdout pipe: %w", err)
	}

	session.Stderr = os.Stderr

	// Start SCP in remote-to-local mode (receiving file) with target path
	// scp -t <destination> means "receive a file to destination"
	scpCmd := fmt.Sprintf("scp -t %q", remotePath)
	if err := session.Start(scpCmd); err != nil {
		return fmt.Errorf("starting SCP on remote: %w", err)
	}

	// Wait for the initial 0 (ACK) — read from stdout
	if err := scpWaitAck(stdout); err != nil {
		return fmt.Errorf("waiting for SCP ACK: %w", err)
	}

	// Send file header: C<perms> <size> <filename>\n
	// Use the remote file's base name
	remoteFilename := path.Base(remotePath)
	header := fmt.Sprintf("C%04o %d %s\n", mode.Perm(), fileSize, remoteFilename)
	if verbose {
		fmt.Fprintf(os.Stderr, "[SCP] Sending header: %s", header)
	}
	if _, err := fmt.Fprint(stdin, header); err != nil {
		return fmt.Errorf("sending SCP header: %w", err)
	}

	// Wait for ACK on header — read from stdout
	if err := scpWaitAck(stdout); err != nil {
		return fmt.Errorf("waiting for SCP header ACK: %w", err)
	}

	// Send file contents in chunks
	buf := make([]byte, scpBufferSize)
	var sent int64
	for sent < fileSize {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("reading local file: %w", err)
		}
		if n == 0 {
			break
		}
		if _, err := stdin.Write(buf[:n]); err != nil {
			return fmt.Errorf("writing file data to SCP stream: %w", err)
		}
		sent += int64(n)
	}

	// Send EOF marker (0x00)
	if _, err := stdin.Write([]byte{0}); err != nil {
		return fmt.Errorf("sending SCP EOF marker: %w", err)
	}

	// Wait for final ACK — read from stdout
	if err := scpWaitAck(stdout); err != nil {
		return fmt.Errorf("waiting for SCP final ACK: %w", err)
	}

	// Send close marker
	if _, err := fmt.Fprint(stdin, "E\n"); err != nil {
		return fmt.Errorf("sending SCP close marker: %w", err)
	}

	return session.Wait()
}

// uploadDir recursively uploads a directory to the remote host.
func (c *Client) uploadDir(localPath, remotePath string, verbose bool) error {
	// First create the remote directory
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("creating SSH session for mkdir: %w", err)
	}
	mkdirCmd := fmt.Sprintf("mkdir -p %q", remotePath)
	if err := session.Run(mkdirCmd); err != nil {
		session.Close()
		return fmt.Errorf("creating remote directory %q: %w", remotePath, err)
	}
	session.Close()

	// Walk local directory and upload each file
	entries, err := os.ReadDir(localPath)
	if err != nil {
		return fmt.Errorf("reading local directory %q: %w", localPath, err)
	}

	for _, entry := range entries {
		localEntryPath := filepath.Join(localPath, entry.Name())
		remoteEntryPath := path.Join(remotePath, entry.Name())

		if entry.IsDir() {
			if err := c.uploadDir(localEntryPath, remoteEntryPath, verbose); err != nil {
				return err
			}
		} else {
			info, err := entry.Info()
			if err != nil {
				return fmt.Errorf("getting file info for %q: %w", localEntryPath, err)
			}
			if err := c.uploadFile(localEntryPath, remoteEntryPath, info.Mode(), verbose); err != nil {
				return err
			}
			// SCP protocol requires a small delay between files in a session
			time.Sleep(50 * time.Millisecond)
		}
	}

	return nil
}

// scpWaitAck reads and validates the SCP protocol acknowledgement (0x00 byte).
func scpWaitAck(r io.Reader) error {
	buf := make([]byte, 1)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return fmt.Errorf("reading SCP ACK: %w", err)
	}
	if buf[0] != 0 {
		return fmt.Errorf("unexpected SCP response: got byte 0x%02x, expected 0x00", buf[0])
	}
	return nil
}
