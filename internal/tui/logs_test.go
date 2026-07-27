package tui

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"testing"
	"time"

	"github.com/uphiago/barbarossa-cli/internal/config"
	"github.com/uphiago/barbarossa-cli/internal/docker"
)

type fakeLogSource struct {
	streams map[string]string
}

func (f *fakeLogSource) ContainerLogs(_ context.Context, name string, _ int) (docker.ContainerLogsResult, error) {
	var framed bytes.Buffer
	payload := f.streams[name]
	header := make([]byte, 8)
	header[0] = 1
	binary.BigEndian.PutUint32(header[4:], uint32(len(payload)))
	framed.Write(header)
	framed.WriteString(payload)
	return io.NopCloser(&framed), nil
}

func TestLogsModelStreamsDockerLogs(t *testing.T) {
	model := NewLogsModel(&fakeLogSource{
		streams: map[string]string{"charlie": "real docker log\n"},
	}, []string{"charlie"})

	msg := model.Init()()
	entry, ok := msg.(LogEntryMsg)
	if !ok {
		t.Fatalf("got message %T, want LogEntryMsg", msg)
	}
	if entry.Worker != "charlie" || entry.Line != "real docker log" {
		t.Fatalf("got %+v", entry)
	}
}

func TestAppRoutesLogEntriesWhileAnotherTabIsActive(t *testing.T) {
	app := NewApp(nil, &config.Config{
		Containers: config.ContainerConfig{Names: []string{"charlie"}},
	})
	app.activeTab = 0
	app.logModel = NewLogsModel(nil, []string{"charlie"})
	app.logModel.events = make(chan LogEntryMsg)

	app.Update(LogEntryMsg{
		Worker:    "charlie",
		Line:      "background log",
		Timestamp: time.Now(),
	})

	if len(app.logModel.entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(app.logModel.entries))
	}
}
