package vidispine

import (
	"encoding/xml"
	"testing"
)

var sample_data = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<FileDocument xmlns="http://xml.vidispine.com/schema/vidispine">
    <id>VX-1</id>
    <path>Multimedia_Documentaries/Yusuf_Test_5_June/yusuf_parkar_Yusuf_Test_5_June_WITH_MEDIA/KP-27179 RICH.prproj</path>
    <state>LOST</state>
    <size>598445</size>
    <hash>b7ac192e46360c0e21fd33df65db214f97bb66be</hash>
    <timestamp>2019-09-11T11:45:00.883+01:00</timestamp>
    <refreshFlag>1</refreshFlag>
    <storage>VX-2</storage>
    <metadata>
        <field>
            <key>created</key>
            <value>1507050044879</value>
        </field>
        <field>
            <key>mtime</key>
            <value>1507050044879</value>
        </field>
    </metadata>
</FileDocument>`

func TestUnmarshalDoc(t *testing.T) {
	var test VSFileDocument

	sample_data_bytes := []byte(sample_data)

	err := xml.Unmarshal(sample_data_bytes, &test)

	if err != nil {
		t.Error("Could not unmarshal xml: ", err)
	}

	if len(test.Metadata) != 2 {
		t.Error("Expecting 2 metadata keys, got ", len(test.Metadata))
	}

	if test.Id != "VX-1" {
		t.Error("Did not get expected ID")
	}

	if test.Path != "Multimedia_Documentaries/Yusuf_Test_5_June/yusuf_parkar_Yusuf_Test_5_June_WITH_MEDIA/KP-27179 RICH.prproj" {
		t.Error("Did not get expected path")
	}

	if test.State != "LOST" {
		t.Error("Did not get expected state")
	}

	if test.Size != 598445 {
		t.Error("Did not get expected size")
	}

	if test.Hash != "b7ac192e46360c0e21fd33df65db214f97bb66be" {
		t.Error("Did not get expected hash")
	}
}
