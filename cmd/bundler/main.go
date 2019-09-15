package main

import (
	"archive/zip"
	"github.com/guardian/deliverable_bundler/contentlist"
	"github.com/guardian/deliverable_bundler/vidispine"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"strconv"
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

	vsUri := os.Getenv("vidispine_url")
	vsUser := os.Getenv("vidispine_user")
	vsPass := os.Getenv("vidispine_password")
	vsToken := os.Getenv("vidispine_token")

	if contentListUri == "" || serverToken == "" || outputFile == "" {
		log.Fatal("You need to set content_list, server_token and output_file in the environment")
	}

	vsUriData, uriParseErr := url.Parse(vsUri)
	if uriParseErr != nil {
		log.Fatalf("Could not parse provided Vidispine URI '%s': %s", vsUri, uriParseErr.Error())
	}

	downloadsList, downloadErr := contentlist.DownloadContentList(contentListUri, serverToken)

	if downloadErr != nil {
		log.Fatal("Could not download content list from ", contentListUri)
	}

	writer, initErr := initialiseZip(outputFile)

	if initErr != nil {
		log.Fatal("Could not initialise output writer: ", initErr.Error())
	}

	portPart, _ := strconv.Atoi(vsUriData.Port())

	comm := vidispine.VidispineCommunicator{
		Protocol: vsUriData.Scheme,
		Hostname: vsUriData.Host,
		Port:     int16(portPart),
		User:     vsUser,
		Password: vsPass,
		Token:    vsToken,
	}

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
