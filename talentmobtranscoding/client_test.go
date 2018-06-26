package talentmobtranscoding

import (
	"os"
	"testing"
)

func TestTranscodeWithWatermark(t *testing.T) {
	os.Setenv("ADMIN_TOKEN", "B4A4EC33F369D724984C9E38A9EF3")

	if err := TranscodeWithWatermark(2156); err != nil {
		t.Fatal(err)
	}

}

func TestTranscode(t *testing.T) {
	os.Setenv("ADMIN_TOKEN", "B4A4EC33F369D724984C9E38A9EF3")

	if err := Transcode(2156); err != nil {
		t.Fatal(err)
	}
}
