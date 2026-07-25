package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	mobyclient "github.com/moby/moby/client"
	"github.com/moby/moby/api/types/container"
)

// ─── Types ───────────────────────────────────────────────────

// ContainerInfo holds summary info for a Docker container.
type ContainerInfo struct {
	ID     string
	Name   string
	State  string // "running", "exited", "paused"
	Status string // human-readable: "Up 3 hours", "Exited (0)"
	Image  string
}

// Stats holds resource usage for a container.
type Stats struct {
	CPUPercent float64
	MemoryMB   float64
	MemoryPct  float64
	PIDs       uint64
}

// Client wraps the Docker SDK client.
type Client struct {
	inner *mobyclient.Client
}

// NewClient creates a new Docker client connected via the default socket.
func NewClient() (*Client, error) {
	return NewClientWithHost("")
}

// NewClientWithHost creates a new Docker client connected to the given host.
// If host is empty it uses DOCKER_HOST env var or the default socket.
func NewClientWithHost(host string) (*Client, error) {
	opts := []mobyclient.Opt{mobyclient.WithAPIVersionNegotiation()}
	if host != "" {
		opts = append(opts, mobyclient.WithHost(host))
	} else {
		opts = append(opts, mobyclient.FromEnv)
	}
	cli, err := mobyclient.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	return &Client{inner: cli}, nil
}

// Close closes the underlying Docker client connection.
func (c *Client) Close() error {
	return c.inner.Close()
}

// Ping checks if the Docker daemon is reachable.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.inner.Ping(ctx, mobyclient.PingOptions{})
	return err
}

// ListContainers returns all containers (running and stopped).
func (c *Client) ListContainers(ctx context.Context) ([]ContainerInfo, error) {
	result, err := c.inner.ContainerList(ctx, mobyclient.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	infos := make([]ContainerInfo, 0, len(result.Items))
	for _, ctr := range result.Items {
		name := ctr.Names[0]
		if strings.HasPrefix(name, "/") {
			name = name[1:]
		}
		infos = append(infos, ContainerInfo{
			ID:     ctr.ID[:12],
			Name:   name,
			State:  string(ctr.State),
			Status: ctr.Status,
			Image:  ctr.Image,
		})
	}
	return infos, nil
}

// ContainerStats fetches a one-shot CPU/RAM snapshot for a container.
// name can be container name or ID prefix.
func (c *Client) ContainerStats(ctx context.Context, name string) (*Stats, error) {
	result, err := c.inner.ContainerStats(ctx, name, mobyclient.ContainerStatsOptions{})
	if err != nil {
		return nil, fmt.Errorf("container stats %q: %w", name, err)
	}
	defer result.Body.Close()

	var v container.StatsResponse
	if err := json.NewDecoder(result.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("decode stats %q: %w", name, err)
	}

	return parseStats(&v), nil
}

// ContainerLogsResult wraps the moby log stream.
type ContainerLogsResult = mobyclient.ContainerLogsResult

// ContainerLogs streams container stdout+stderr from the last `tail` lines.
func (c *Client) ContainerLogs(ctx context.Context, name string, tail int) (ContainerLogsResult, error) {
	opts := mobyclient.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       fmt.Sprintf("%d", tail),
	}
	return c.inner.ContainerLogs(ctx, name, opts)
}

// ─── internal helpers ───────────────────────────────────────

func parseStats(v *container.StatsResponse) *Stats {
	cpuDelta := float64(v.CPUStats.CPUUsage.TotalUsage) - float64(v.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(v.CPUStats.SystemUsage) - float64(v.PreCPUStats.SystemUsage)
	cpuPct := 0.0
	if sysDelta > 0 && cpuDelta > 0 {
		numCores := float64(v.CPUStats.OnlineCPUs)
		if numCores == 0 {
			numCores = float64(len(v.CPUStats.CPUUsage.PercpuUsage))
		}
		if numCores > 0 {
			cpuPct = math.Round((cpuDelta/sysDelta)*numCores*100*10) / 10
		}
	}

	memUsed := float64(v.MemoryStats.Usage - v.MemoryStats.Stats["cache"])
	memLimit := float64(v.MemoryStats.Limit)
	memMB := math.Round(memUsed / 1024 / 1024 * 10) / 10
	memPct := 0.0
	if memLimit > 0 {
		memPct = math.Round(memUsed/memLimit*100*10) / 10
	}

	return &Stats{
		CPUPercent: cpuPct,
		MemoryMB:   memMB,
		MemoryPct:  memPct,
		PIDs:       v.PidsStats.Current,
	}
}

// NowStreamLogs reads from a Docker log stream and sends lines over a channel.
func NowStreamLogs(ctx context.Context, stream io.ReadCloser, lines chan<- string) {
	defer close(lines)
	defer stream.Close()

	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		n, err := stream.Read(buf)
		if n > 0 {
			data := buf[:n]
			if len(data) > 8 {
				data = data[8:]
			}
			text := strings.TrimRight(string(data), "\n\r")
			if text != "" {
				select {
				case lines <- text:
				case <-ctx.Done():
					return
				}
			}
		}
		if err != nil {
			return
		}
	}
}

// WatchActivity polls container stats periodically and sends activity messages.
func WatchActivity(ctx context.Context, cli *Client, interval time.Duration, activities chan<- string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			containers, err := cli.ListContainers(ctx)
			if err != nil {
				continue
			}
			for _, ctr := range containers {
				state := "stopped"
				if ctr.State == "running" {
					state = "running"
				}
				msg := fmt.Sprintf("%s => %s (%s)", ctr.Name, state, ctr.Status)
				select {
				case activities <- msg:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
