package lazymcp

import (
	"encoding/json"
	"testing"
)

func TestParseMessage(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		msg, err := ParseMessage([]byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if msg.Method != "initialize" {
			t.Fatalf("got method %q, want %q", msg.Method, "initialize")
		}
	})

	t.Run("empty", func(t *testing.T) {
		_, err := ParseMessage([]byte(""))
		if err == nil {
			t.Fatal("expected error for empty input")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := ParseMessage([]byte("not json"))
		if err == nil {
			t.Fatal("expected error for invalid json")
		}
	})

	t.Run("non-object", func(t *testing.T) {
		_, err := ParseMessage([]byte(`[1,2,3]`))
		if err == nil {
			t.Fatal("expected error for non-object json")
		}
	})
}

func TestMakeResponse(t *testing.T) {
	result := json.RawMessage(`{"tools":[]}`)
	data := MakeResponse(1, result)
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatal(err)
	}
	if msg.ID == nil {
		t.Fatal("response must have id")
	}
	if msg.Result == nil {
		t.Fatal("response must have result")
	}
}

func TestMakeErrorResponse(t *testing.T) {
	data := MakeErrorResponse(1, -32603, "server error")
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Error == nil {
		t.Fatal("error response must have error field")
	}
	if msg.Error.Code != -32603 {
		t.Fatalf("got code %d, want -32603", msg.Error.Code)
	}
}

func TestIsDiscoveryMethod(t *testing.T) {
	discovery := []string{
		"initialize",
		"notifications/initialized",
		"tools/list",
		"resources/list",
		"resources/templates/list",
		"prompts/list",
	}
	for _, m := range discovery {
		if !IsDiscoveryMethod(m) {
			t.Errorf("%q should be discovery method", m)
		}
	}

	nonDiscovery := []string{"tools/call", "resources/read", "prompts/get", "ping"}
	for _, m := range nonDiscovery {
		if IsDiscoveryMethod(m) {
			t.Errorf("%q should NOT be discovery method", m)
		}
	}
}

func TestSerializeMessage(t *testing.T) {
	msg := Message{JSONRPC: "2.0", Method: "test"}
	data := SerializeMessage(&msg)
	if data[len(data)-1] != '\n' {
		t.Fatal("serialized message must end with newline")
	}
	newlines := 0
	for _, b := range data {
		if b == '\n' {
			newlines++
		}
	}
	if newlines != 1 {
		t.Fatalf("got %d newlines, want 1", newlines)
	}
}
