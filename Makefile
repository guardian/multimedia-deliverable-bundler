all: bundler test_downloader

bundler: cmd/bundler/main.go contentlist/ContentList.go vidispine/GenericMetadata.go vidispine/VSCommunicator.go vidispine/VSFile.go vidispine/VSFileReader.go
	cd cmd/bundler; go build

test_downloader: cmd/test_downloader/testdownloader.go contentlist/ContentList.go vidispine/GenericMetadata.go vidispine/VSCommunicator.go vidispine/VSFile.go vidispine/VSFileReader.go
	cd cmd/test_downloader; go build

clean:
	rm -f cmd/bundler/bundler
	rm -f cmd/test_downloader/test_downloader