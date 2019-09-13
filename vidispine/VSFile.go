package vidispine

type VSFileDocument struct {
	Id          string            `xml:"id"`
	Path        string            `xml:"path"`
	State       string            `xml:"state"`
	Size        int64             `xml:"size"`
	Hash        string            `xml:"hash"`
	Timestamp   string            `xml:"timestamp"`
	RefreshFlag int8              `xml:"refreshFlag"`
	StorageId   string            `xml:"storage"`
	Metadata    []GenericMetadata `xml:"metadata>field"`
}
