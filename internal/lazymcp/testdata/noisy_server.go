//go:build ignore

// noisy_server is a mock MCP server that sends unsolicited notifications
// before each response, to test that readResponse forwards non-matching messages.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type message struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method,omitempty"`
	Params  *json.RawMessage `json:"params,omitempty"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var msg message
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}

		switch msg.Method {
		case "initialize":
			// Send an unsolicited notification before the response
			sendNotification("tools/listChanged")
			respond(msg.ID, map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    "noisy-server",
					"version": "0.1.0",
				},
			})
		case "notifications/initialized":
			// no response
		case "tools/list":
			sendNotification("resources/listChanged")
			respond(msg.ID, map[string]any{
				"tools": []any{
					map[string]any{
						"name":        "echo",
						"description": "Echo tool",
						"inputSchema": map[string]any{
							"type":       "object",
							"properties": map[string]any{"text": map[string]any{"type": "string"}},
						},
					},
				},
			})
		case "tools/call":
			var params struct {
				Arguments struct {
					Text string `json:"text"`
				} `json:"arguments"`
			}
			if msg.Params != nil {
				json.Unmarshal(*msg.Params, &params)
			}
			respond(msg.ID, map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": fmt.Sprintf("echo: %s", params.Arguments.Text)},
				},
			})
		}
	}
}

func sendNotification(method string) {
	notif := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
	}
	data, _ := json.Marshal(notif)
	fmt.Fprintf(os.Stdout, "%s\n", data)
}

func respond(id *json.RawMessage, result any) {
	resp := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stdout, "%s\n", data)
}
