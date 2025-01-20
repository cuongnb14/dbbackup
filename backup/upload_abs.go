package backup

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func UploadToABS(filePath string) {
	// Load environment variables
	accountName := os.Getenv("ABS_ACCOUNT_NAME")
	accountKey := os.Getenv("ABS_ACCESS_KEY")
	absContainer := os.Getenv("ABS_CONTAINER")

	if accountName == "" || accountKey == "" || absContainer == "" {
		log.Fatal("Missing required environment variables: ABS_ACCOUNT_NAME, ABS_ACCESS_KEY, or ABS_CONTAINER")
	}

	// Create a shared key credential
	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatalf("Failed to create shared key credential: %v", err)
	}

	// Create a blob service client
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		log.Fatalf("Failed to create Azure Blob Storage client: %v", err)
	}

	// Read the file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Generate blob name (using filePath's name for example)
	blobName := "databases/" + filePath

	// Upload the file
	_, err = client.UploadBuffer(context.Background(), absContainer, blobName, fileData, nil)
	if err != nil {
		log.Fatalf("Upload buffer to Azure Blob Storage failed: %v", err)
	}

	log.Printf("File uploaded successfully to Azure Blob Storage: %s", blobName)
}
