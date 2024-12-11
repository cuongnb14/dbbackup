package backup

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func EncryptFile(filePath, passphrase string) error {
	// GPG: echo "<passphrase>" | gpg --batch --yes --passphrase-fd 0 --symmetric --cipher-algo AES256 <filePath>
	cmd := exec.Command(
		"gpg", "--batch", "--yes", "--passphrase-fd", "0", "--symmetric", "--cipher-algo", "AES256",
		"--output", filePath+".gpg",
		filePath,
	)
	cmd.Stdin = strings.NewReader(passphrase)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to encrypt file: %v", err)
	}

	err = os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to remove original file after encryption: %v", err)
	}

	return nil
}
