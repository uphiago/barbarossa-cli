package docker

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"testing"
)

func TestNowStreamLogsDemultiplexesDockerFramesIntoLines(t *testing.T) {
	var stream bytes.Buffer
	writeLogFrame(t, &stream, 1, "first line\npartial")
	writeLogFrame(t, &stream, 2, " line\nthird line\n")

	lines := make(chan string)
	go NowStreamLogs(context.Background(), io.NopCloser(&stream), lines)

	var got []string
	for line := range lines {
		got = append(got, line)
	}

	want := []string{"first line", "partial line", "third line"}
	if len(got) != len(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func writeLogFrame(t *testing.T, dst io.Writer, stream byte, payload string) {
	t.Helper()

	header := make([]byte, 8)
	header[0] = stream
	binary.BigEndian.PutUint32(header[4:], uint32(len(payload)))
	if _, err := dst.Write(header); err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(dst, payload); err != nil {
		t.Fatal(err)
	}
}
