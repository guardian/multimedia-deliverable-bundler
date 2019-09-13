package vidispine

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

type VSFileReader struct {
	storageId string
	fileId    string
	bytesRead int64
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
func NewVSFileReader(communicator *VidispineCommunicator, fileData *VSFileDocument) (*VSFileReader, error) {
	rtn := VSFileReader{fileData.StorageId, fileData.Id, 0, fileData, communicator}
	return &rtn, nil
}

func (r *VSFileReader) nextBlockSize(cap int) int {
	potentialSize := int64(cap)

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
	bytesToRead := r.nextBlockSize(cap(p))
	if bytesToRead == 0 {
		log.Printf("Download completed")
		return 0, io.EOF
	}
	log.Printf("Reading %d bytes, total read %d / %d...", bytesToRead, r.bytesRead, r.fileData.Size)

	fmt.Printf("Bytes=%d-%d\n", r.bytesRead, r.bytesRead+int64(bytesToRead)-1)

	headers := map[string]string{
		"Range": fmt.Sprintf("Bytes=%d-%d", r.bytesRead, r.bytesRead+int64(bytesToRead)-1),
	}
	query := map[string]string{}
	matrix := map[string]string{}

	urlpath := fmt.Sprintf("/API/storage/%s/file/%s/data", r.storageId, r.fileId)
	response, vsErr := r.comm.MakeRequestRaw("GET", urlpath, matrix, query, headers, nil)

	if vsErr != nil {
		log.Print("Could not get chunk from server: ", vsErr)
		return 0, vsErr
	}

	buf, readErr := ioutil.ReadAll(response.Body)

	if len(buf) == 0 {
		log.Print(response.Header)
		log.Print(response.Body)
		log.Print(response.StatusCode)
		panic("Zero bytes read")
	}
	log.Printf("Dest capacity %d, buffer length %d", cap(p), len(buf))
	if readErr != nil {
		log.Print("Could not read request body: ", readErr)
		return 0, readErr
	} else {
		copy(p, buf)
		r.bytesRead += int64(bytesToRead)
		return bytesToRead, nil
	}
}

func BufferedCopy(dst io.Writer, src io.Reader, bufsize int) (int, error) {
	totalRead := 0
	for {
		buf := make([]byte, bufsize)
		countRead, readErr := src.Read(buf)

		if readErr != nil {
			if readErr == io.EOF {
				return totalRead, nil
			} else {
				return totalRead, readErr
			}
		}

		if countRead == 0 {
			return totalRead, nil
		}

		writeBuf := buf[:countRead]
		_, writeErr := dst.Write(writeBuf)
		if writeErr != nil {
			return totalRead, writeErr
		}
	}
}
