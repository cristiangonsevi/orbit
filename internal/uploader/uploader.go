package uploader

import (
	"fmt"
	"os"

	"github.com/cristiangonsevi/orbit/internal/config"
	sshclient "github.com/cristiangonsevi/orbit/internal/ssh"
)

// UploadFiles uploads all entries from the upload config using the SSH client.
func UploadFiles(client *sshclient.Client, entries []config.UploadEntry, verbose bool) error {
	if len(entries) == 0 {
		return nil
	}

	for i, entry := range entries {
		if verbose {
			fmt.Fprintf(os.Stderr, "[UPLOAD] Transfer %d/%d: %s → %s\n",
				i+1, len(entries), entry.Source, entry.Destination)
		}

		if err := client.UploadFile(entry, verbose); err != nil {
			return fmt.Errorf("uploading %q to %q: %w", entry.Source, entry.Destination, err)
		}
	}

	return nil
}

// DryRunUploads prints what uploads would be performed.
func DryRunUploads(entries []config.UploadEntry) {
	fmt.Println("\n📋 Step 3: Uploads")
	if len(entries) == 0 {
		fmt.Println("  (no uploads configured)")
		return
	}
	for i, entry := range entries {
		fmt.Printf("  %d. %s → %s\n", i+1, entry.Source, entry.Destination)
	}
}
