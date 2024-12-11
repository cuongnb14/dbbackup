package backup

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func UploadToABS(filePath string) {
	absUrl := os.Getenv("ABS_URL")
	absContainer := os.Getenv("ABS_CONTAINER")
	absSAS := os.Getenv("ABS_SAS")

	sasURL := fmt.Sprintf("%s?%s", absUrl, absSAS)

	// Read the file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Create a BlobClient
	client, err := azblob.NewClientWithNoCredential(sasURL, nil)
	if err != nil {
		log.Fatal("failed to create abs client: ", "err", err)
	}
	// Upload the file
	_, err = client.UploadBuffer(context.Background(), absContainer, "databases"+filePath, fileData, nil)
	if err != nil {
		log.Fatal("upload buffer to abs failed: ", "err", err)
	}

	log.Printf("File uploaded successfully!")
}
