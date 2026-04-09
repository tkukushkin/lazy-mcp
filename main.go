package main

import (
	"fmt"
	"os"

	"github.com/tkukushkin/lazy-mcp/internal/lazymcp"
)

func main() {
	args := os.Args[1:]

	if len(args) > 0 && args[0] == "clear-cache" {
		dir := lazymcp.CacheDir()
		if err := os.RemoveAll(dir); err != nil {
			fmt.Fprintf(os.Stderr, "lazy-mcp: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Removed %s\n", dir)
		return
	}

	sepIdx := -1
	for i, a := range args {
		if a == "--" {
			sepIdx = i
			break
		}
	}
	if sepIdx == -1 || sepIdx+1 >= len(args) {
		fmt.Fprintln(os.Stderr, "Usage: lazy-mcp -- <command> [args...]")
		fmt.Fprintln(os.Stderr, "       lazy-mcp clear-cache")
		os.Exit(1)
	}
	command := args[sepIdx+1:]

	cache := lazymcp.NewCache(command)
	proxy := lazymcp.NewProxy(command, cache)
	if err := proxy.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "lazy-mcp: %v\n", err)
		os.Exit(1)
	}
}
