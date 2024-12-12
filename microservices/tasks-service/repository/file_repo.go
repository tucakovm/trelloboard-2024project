package repository

import (
	"fmt"
	"github.com/colinmarc/hdfs/v2"
	"log"
)

type HDFSRepository struct {
	Client *hdfs.Client
	logger *log.Logger
}

func NewHDFSRepository(logger *log.Logger, namenodeURL string) (*HDFSRepository, error) {
	client, err := hdfs.New(namenodeURL)
	if err != nil {
		return nil, fmt.Errorf("error creating HDFS client: %v", err)
	}
	return &HDFSRepository{
		Client: client,
		logger: logger,
	}, nil
}
func (repo *HDFSRepository) Close() {
	if repo.Client != nil {
		repo.Client.Close()
	}
}

func (repo *HDFSRepository) UploadFile(taskID, fileName string, content []byte) error {
	// Define the HDFS path (modify according to your project structure)
	hdfsPath := fmt.Sprintf("/tasks/%s/%s", taskID, fileName)

	// Open a file in HDFS for writing (this creates the file if it doesn't exist)
	file, err := repo.Client.Create(hdfsPath)
	if err != nil {
		return fmt.Errorf("failed to create file on HDFS at %s: %v", hdfsPath, err)
	}
	defer file.Close()

	// Write content to the file
	_, err = file.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write content to file on HDFS at %s: %v", hdfsPath, err)
	}

	// Flush the file to ensure all data is written
	err = file.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush file to HDFS at %s: %v", hdfsPath, err)
	}

	repo.logger.Printf("Successfully uploaded file %s to HDFS at %s\n", fileName, hdfsPath)
	return nil
}

func (repo *HDFSRepository) DownloadFile(taskID string, fileName string) ([]byte, error) {
	// Check if repo.client is nil
	if repo.Client == nil {
		repo.logger.Println("Error: repo.client is nil")
		return nil, fmt.Errorf("HDFS client is not initialized")
	}

	hdfsPath := "/tasks/" + taskID + "/" + fileName

	// Open file in HDFS
	file, err := repo.Client.Open(hdfsPath)
	if err != nil {
		repo.logger.Println("Error opening file in HDFS:", err)
		return nil, err
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil {
		repo.logger.Println("Error reading file in HDFS:", err)
		return nil, err
	}

	return buffer[:n], nil
}

func (repo *HDFSRepository) DeleteFile(taskID string, fileName string) error {
	// Check if repo.client is nil
	if repo.Client == nil {
		repo.logger.Println("Error: repo.client is nil")
		return fmt.Errorf("HDFS client is not initialized")
	}

	hdfsPath := "/tasks/" + taskID + "/" + fileName

	err := repo.Client.Remove(hdfsPath)
	if err != nil {
		repo.logger.Println("Error deleting file in HDFS:", err)
		return err
	}

	return nil
}
