package main

import (
	"archive/zip"
	"github.com/guardian/deliverable_bundler/contentlist"
	"github.com/guardian/deliverable_bundler/vidispine"
	"io"
	"log"
	"os"
	"path"
)

func initialiseZip(filename string) (*zip.Writer, error) {
	newFile, openErr := os.Create(filename)
	if openErr != nil {
		return nil, openErr
	}

	zipWriter := zip.NewWriter(newFile)
	return zipWriter, nil
}

func addToZip(zipWriter *zip.Writer, src io.Reader, name string, size int64) error {
	header := zip.FileHeader{
		Name:               name,
		UncompressedSize64: uint64(size),
	}

	dest, writErr := zipWriter.CreateHeader(&header)
	if writErr != nil {
		return writErr
	}

	_, copyErr := vidispine.BufferedCopy(dest, src, 4*1024*1024)

	if copyErr != nil {
		return copyErr
	}

	return nil
}

func main() {
	contentListUri := os.Getenv("content_list")
	serverToken := os.Getenv("server_token")
	outputFile := os.Getenv("output_file")

	if contentListUri == "" || serverToken == "" || outputFile == "" {
		log.Fatal("You need to set content_list, server_token and output_file in the environment")
	}

	downloadsList, downloadErr := contentlist.DownloadContentList(contentListUri, serverToken)

	if downloadErr != nil {
		log.Fatal("Could not download content list from ", contentListUri)
	}

	writer, initErr := initialiseZip(outputFile)

	if initErr != nil {
		log.Fatal("Could not initialise output writer: ", initErr.Error())
	}

	comm := vidispine.VidispineCommunicator{}

	success := true

	for _, item := range downloadsList {
		fileData, vsErr := vidispine.VSFileInfo(&comm, item.StorageId, item.FileId)

		if vsErr != nil {
			log.Printf("Could not open file connection to VS: %s", vsErr.Error())
			success = false
			break
		}

		basename := path.Base(fileData.Path)

		reader, readErr := vidispine.NewVSFileReader(&comm, fileData)
		if readErr != nil {
			log.Printf("Could not read from %s on %s", item.FileId, item.StorageId)
			success = false
			break
		}

		addErr := addToZip(writer, reader, basename, fileData.Size)

		if addErr != nil {
			log.Printf("Could not add stream to zip: %s", addErr.Error())
			success = false
			break
		}
	}

	closeErr := writer.Close()
	if closeErr != nil {
		log.Printf("Could not close zip writer: %s", closeErr.Error())
	}

	if success == false {
		log.Printf("Could not create file")
		os.Remove(outputFile)
		os.Exit(2)
	} else {
		log.Printf("Output file completed at %s", outputFile)
		os.Exit(0)
	}
}
