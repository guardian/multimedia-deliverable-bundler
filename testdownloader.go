package main

import (
	"flag"
	"github.com/guardian/deliverable_bundler/vidispine"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

func main() {
	var storageId string
	var fileId string
	var output string
	var server string
	var proto string
	var port int
	var user string
	var passfile string

	flag.StringVar(&storageId, "storage-id", "", "Vidispine storage ID to read from")
	flag.StringVar(&fileId, "file-id", "", "Vidispine file ID to read")
	flag.StringVar(&output, "output", "test.dat", "Filename to write data to")
	flag.StringVar(&proto, "proto", "http", "Protocol to communicate with Vidsispine. Must be either http or https.")
	flag.IntVar(&port, "port", 8080, "Port to communicate with Vidispine")
	flag.StringVar(&server, "server", "localhost", "Hostname to communicate with Vidispine")
	flag.StringVar(&user, "user", "admin", "Username to communicate with Vidispine")
	flag.StringVar(&passfile, "passfile", ".vspass", "file that contains password to authenticate")
	flag.Parse()

	if storageId == "" || fileId == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	fileContent, readErr := ioutil.ReadFile(passfile)
	replacer := regexp.MustCompile("\\s+")
	passwdContent := replacer.ReplaceAll(fileContent, []byte(""))

	if readErr != nil {
		log.Fatal("Could not open ", passfile, ": ", readErr)
	}

	comm := vidispine.VidispineCommunicator{proto, server, int16(port), user, string(passwdContent)}

	fileData, vsLookupErr := vidispine.VSFileInfo(&comm, storageId, fileId)

	if vsLookupErr != nil {
		log.Fatal("Could not look up file")
	}

	log.Print("Found file ", fileData.Path, " with size ", fileData.Size, " and hash ", fileData.Hash)

	reader, err := vidispine.NewVSFileReader(&comm, storageId, fileId) //default to 2Mb block size
	if err != nil {
		log.Fatal("Could not set up file reader: ", err.Error())
	}

	fp, openErr := os.Create(output)
	if openErr != nil {
		log.Fatal("Could not open output file '", output, "' ", openErr.Error())
	}

	defer fp.Close()
	log.Print("Copying data into ", output, "....\n")

	_, copyErr := vidispine.BufferedCopy(fp, reader, 40*1024*1024)

	if copyErr != nil {
		log.Fatal("Could not copy data: ", copyErr.Error())
	}

	log.Print("All done!")
}
