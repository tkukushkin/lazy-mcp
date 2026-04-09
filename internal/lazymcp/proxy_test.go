package lazymcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func buildMockServer(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "mock_server")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	_, thisFile, _, _ := runtime.Caller(0)
	src := filepath.Join(filepath.Dir(thisFile), "testdata", "mock_server.go")
	cmd := exec.Command("go", "build", "-o", bin, src)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build mock server: %v", err)
	}
	return bin
}

func buildProxy(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "lazy-mcp")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, "../..")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build proxy: %v", err)
	}
	return bin
}

func request(id any, method string, params ...string) string {
	msg := map[string]any{"jsonrpc": "2.0", "id": id, "method": method}
	if len(params) > 0 {
		msg["params"] = json.RawMessage(params[0])
	}
	data, _ := json.Marshal(msg)
	return string(data)
}

func notification(method string) string {
	data, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "method": method})
	return string(data)
}

const initParams = `{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}`

func runProxy(t *testing.T, proxyBin, serverBin, cacheDir string, messages []string, numResponses int) []map[string]any {
	t.Helper()
	input := strings.Join(messages, "\n") + "\n"

	cmd := exec.Command(proxyBin, "--", serverBin)
	cmd.Env = append(os.Environ(), fmt.Sprintf("LAZY_MCP_CACHE_DIR=%s", cacheDir))
	cmd.Stdin = strings.NewReader(input)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	scanner := bufio.NewScanner(stdout)
	var responses []map[string]any
	for i := 0; i < numResponses && scanner.Scan(); i++ {
		var resp map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v\nraw: %s", err, scanner.Text())
		}
		responses = append(responses, resp)
	}

	io.Copy(io.Discard, stdout)
	cmd.Wait()
	return responses
}

func TestNoCachePath(t *testing.T) {
	proxyBin := buildProxy(t)
	serverBin := buildMockServer(t)
	cacheDir := t.TempDir()

	messages := []string{
		request(1, "initialize", initParams),
		notification("notifications/initialized"),
		request(2, "tools/list"),
		request(3, "tools/call", `{"name":"echo","arguments":{"text":"hi"}}`),
	}

	responses := runProxy(t, proxyBin, serverBin, cacheDir, messages, 3)

	if len(responses) < 3 {
		t.Fatalf("got %d responses, want 3", len(responses))
	}

	result := responses[0]["result"].(map[string]any)
	serverInfo := result["serverInfo"].(map[string]any)
	if serverInfo["name"] != "mock-server" {
		t.Fatalf("got server name %v", serverInfo["name"])
	}

	result = responses[1]["result"].(map[string]any)
	tools := result["tools"].([]any)
	if len(tools) != 1 {
		t.Fatalf("got %d tools, want 1", len(tools))
	}

	result = responses[2]["result"].(map[string]any)
	content := result["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if text != "echo: hi" {
		t.Fatalf("got %q, want %q", text, "echo: hi")
	}

	files, _ := filepath.Glob(filepath.Join(cacheDir, "*.json"))
	if len(files) != 1 {
		t.Fatalf("got %d cache files, want 1", len(files))
	}
}

func TestCachedPath(t *testing.T) {
	proxyBin := buildProxy(t)
	serverBin := buildMockServer(t)
	cacheDir := t.TempDir()

	// First run: build cache
	messages := []string{
		request(1, "initialize", initParams),
		notification("notifications/initialized"),
		request(2, "tools/list"),
		request(3, "tools/call", `{"name":"echo","arguments":{"text":"x"}}`),
	}
	runProxy(t, proxyBin, serverBin, cacheDir, messages, 3)

	// Second run: only discovery
	messages = []string{
		request(10, "initialize", initParams),
		notification("notifications/initialized"),
		request(11, "tools/list"),
	}
	responses := runProxy(t, proxyBin, serverBin, cacheDir, messages, 2)

	if len(responses) < 2 {
		t.Fatalf("got %d responses, want 2", len(responses))
	}

	id := responses[0]["id"]
	if fmt.Sprintf("%v", id) != "10" {
		t.Fatalf("got id %v, want 10", id)
	}
}

func TestLiveTransition(t *testing.T) {
	proxyBin := buildProxy(t)
	serverBin := buildMockServer(t)
	cacheDir := t.TempDir()

	// First run: build cache
	messages := []string{
		request(1, "initialize", initParams),
		notification("notifications/initialized"),
		request(2, "tools/list"),
		request(3, "tools/call", `{"name":"echo","arguments":{"text":"x"}}`),
	}
	runProxy(t, proxyBin, serverBin, cacheDir, messages, 3)

	// Second run: cache then live
	messages = []string{
		request(10, "initialize", initParams),
		notification("notifications/initialized"),
		request(11, "tools/list"),
		request(12, "tools/call", `{"name":"echo","arguments":{"text":"live"}}`),
	}
	responses := runProxy(t, proxyBin, serverBin, cacheDir, messages, 3)

	if len(responses) < 3 {
		t.Fatalf("got %d responses, want 3", len(responses))
	}

	result := responses[2]["result"].(map[string]any)
	content := result["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if text != "echo: live" {
		t.Fatalf("got %q, want %q", text, "echo: live")
	}
}

func TestBadCommand(t *testing.T) {
	proxyBin := buildProxy(t)
	cacheDir := t.TempDir()

	input := request(1, "initialize", initParams) + "\n"
	cmd := exec.Command(proxyBin, "--", "nonexistent-command-xyz")
	cmd.Env = append(os.Environ(), fmt.Sprintf("LAZY_MCP_CACHE_DIR=%s", cacheDir))
	cmd.Stdin = strings.NewReader(input)
	out, _ := cmd.Output()

	var resp map[string]any
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to parse: %v\nraw: %s", err, out)
	}
	if resp["error"] == nil {
		t.Fatal("expected error response")
	}
}
