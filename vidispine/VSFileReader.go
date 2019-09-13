package vidispine

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
)

type VSFileReader struct {
	storageId string
	fileId    string
	bytesRead int64
	blockSize int64
	fileData  *VSFileDocument
	comm      *VidispineCommunicator
}

func VSFileInfo(communicator *VidispineCommunicator, storageId string, fileId string) (*VSFileDocument, error) {
	var fileData VSFileDocument
	requestUrl := fmt.Sprintf("/API/storage/%s/file/%s", storageId, fileId)
	headers := map[string]string{
		"Accept": "application/xml",
	}
	result, vsErr := communicator.MakeRequest("GET", requestUrl, map[string]string{}, map[string]string{}, headers, nil)

	if vsErr != nil {
		log.Printf("Could not request file information for %s on %s: %s", fileId, storageId, vsErr)
		return nil, vsErr
	}

	parseErr := xml.Unmarshal(result, &fileData)
	if parseErr != nil {
		log.Print("Could not decode server response: ", vsErr)
		return nil, vsErr
	}
	return &fileData, nil
}

/**
create a new VSFileReader for the given storageId and fileId
*/
func NewVSFileReader(communicator *VidispineCommunicator, storageId string, fileId string, blockSize int64) (*VSFileReader, error) {
	fileData, vsErr := VSFileInfo(communicator, storageId, fileId)

	if vsErr != nil {
		return nil, vsErr
	}

	rtn := VSFileReader{storageId, fileId, 0, blockSize, fileData, communicator}
	return &rtn, nil
}

func (r *VSFileReader) nextBlockSize() int {
	potentialSize := r.blockSize

	if r.fileData.Size == -1 {
		panic("VS reports file size as -1, can't determine size!")
	}

	if r.fileData.Size < r.bytesRead+potentialSize {
		return int(r.fileData.Size - r.bytesRead)
	} else {
		return int(potentialSize)
	}
}

func (r *VSFileReader) Read(p []byte) (int, error) {
	bytesToRead := r.nextBlockSize()
	log.Printf("Reading %d bytes...", bytesToRead)
	headers := map[string]string{
		"Range": fmt.Sprintf("Bytes=%d-%d", r.bytesRead, bytesToRead),
	}
	query := map[string]string{}
	matrix := map[string]string{}

	response, vsErr := r.comm.MakeRequestRaw("GET", "/API/storage/%s/file/%s/data", matrix, query, headers, nil)

	if vsErr != nil {
		log.Print("Could not get chunk from server: ", vsErr)
		return 0, vsErr
	}

	p, readErr := ioutil.ReadAll(response.Body)

	if readErr != nil {
		log.Print("Could not read request body: ", readErr)
		return 0, readErr
	} else {
		r.bytesRead += int64(bytesToRead)
		return bytesToRead, nil
	}
}
