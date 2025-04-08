package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/paulsmith/mcp-go/mcp"
)

// List of allowed safe (read-only) Tailscale subcommands
var safeCommands = map[string]bool{
	"netcheck":  true,
	"ip":        true,
	"dns":       true,
	"status":    true,
	"metrics":   true,
	"ping":      true,
	"version":   true,
	"exit-node": true,
	"whois":     true,
}

func main() {
	server := mcp.NewMCPServer("tailscale", "0.0.1")

	addTailscaleTools(server)
	addTailscalePrompts(server)

	ctx := context.Background()
	if err := server.ConnectStdio(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "Error connecting to MCP:", err)
		os.Exit(1)
	}

	server.LogInfo(ctx, "Tailscale MCP server started", "main")

	select {}
}

func addTailscaleTools(server *mcp.MCPServer) {
	// Tool for running Tailscale command
	runCommandSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"command": {
				"type": "string",
				"description": "The tailscale subcommand to run (e.g., status, ip, netcheck)"
			},
			"args": {
				"type": "string",
				"description": "Optional arguments for the command"
			}
		},
		"required": ["command"]
	}`)

	server.Tool("tailscale", "Run a tailscale command", runCommandSchema,
		func(ctx context.Context, args map[string]interface{}) (string, error) {
			command := args["command"].(string)

			// Validate the command is in our safe list
			if !safeCommands[command] {
				return "", fmt.Errorf("unsafe command '%s' not allowed - only read-only commands permitted", command)
			}

			// Build the command
			cmdArgs := []string{command}

			// Add optional arguments if provided
			if argsStr, ok := args["args"].(string); ok && argsStr != "" {
				cmdArgs = append(cmdArgs, strings.Fields(argsStr)...)
			}

			// Execute the command
			cmd := exec.Command("tailscale", cmdArgs...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("command failed: %v\nOutput: %s", err, string(output))
			}

			return string(output), nil
		})

	// Tool for getting Tailscale IP addresses
	server.Tool("get-ip", "Get Tailscale IP addresses", json.RawMessage(`{
		"type": "object",
		"properties": {}
	}`),
		func(ctx context.Context, args map[string]interface{}) (string, error) {
			cmd := exec.Command("tailscale", "ip")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("failed to get Tailscale IP: %v", err)
			}
			return string(output), nil
		})

	// Tool for checking Tailscale status
	server.Tool("get-status", "Get Tailscale status", json.RawMessage(`{
  "type": "object",
  "properties": {
    "active": {
      "type": "boolean",
      "description": "Filter output to only peers with active sessions"
    },
    "json": {
      "type": "boolean",
      "description": "output in JSON format"
    }
  }
}`),
		func(ctx context.Context, args map[string]interface{}) (string, error) {
			showActive := false
			if active, ok := args["active"].(bool); ok {
				showActive = active
			}
			outputJson := false
			if json, ok := args["json"].(bool); ok {
				outputJson = json
			}

			cmdArgs := []string{"status"}
			if showActive {
				cmdArgs = append(cmdArgs, "--active")
			}
			if outputJson {
				cmdArgs = append(cmdArgs, "--json")
			}

			cmd := exec.Command("tailscale", cmdArgs...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("failed to get Tailscale status: %v", err)
			}
			return string(output), nil
		})

	// Tool for checking network connectivity
	server.Tool("network-check", "Check Tailscale network connectivity", json.RawMessage(`{
		"type": "object",
		"properties": {}
	}`),
		func(ctx context.Context, args map[string]interface{}) (string, error) {
			cmd := exec.Command("tailscale", "netcheck")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("network check failed: %v", err)
			}
			return string(output), nil
		})

	// Tool for querying Tailscale exit nodes
	server.Tool("list-exit-nodes", "List available Tailscale exit nodes", json.RawMessage(`{
		"type": "object",
		"properties": {}
	}`),
		func(ctx context.Context, args map[string]interface{}) (string, error) {
			cmd := exec.Command("tailscale", "exit-node", "list")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("failed to list exit nodes: %v", err)
			}
			return string(output), nil
		})

	// Tool for looking up information about a Tailscale IP
	server.Tool("ip-lookup", "Look up information about a Tailscale IP", json.RawMessage(`{
		"type": "object",
		"properties": {
			"ip": {
				"type": "string",
				"description": "The Tailscale IP address to look up"
			}
		},
		"required": ["ip"]
	}`),
		func(ctx context.Context, args map[string]interface{}) (string, error) {
			ip := args["ip"].(string)

			cmd := exec.Command("tailscale", "whois", ip)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("failed to look up IP %s: %v", ip, err)
			}
			return string(output), nil
		})

	// Tool for pinging a Tailscale host
	server.Tool("ping-host", "Ping a Tailscale host", json.RawMessage(`{
		"type": "object",
		"properties": {
			"host": {
				"type": "string",
				"description": "The Tailscale host to ping (hostname or IP)"
			},
			"count": {
				"type": "integer",
				"description": "Number of pings to send (default: 1)"
			}
		},
		"required": ["host"]
	}`),
		func(ctx context.Context, args map[string]interface{}) (string, error) {
			host := args["host"].(string)

			cmdArgs := []string{"ping", host}

			// Add count if provided
			if count, ok := args["count"].(float64); ok && count > 0 {
				cmdArgs = append(cmdArgs, "-c", fmt.Sprintf("%d", int(count)))
			}

			cmd := exec.Command("tailscale", cmdArgs...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("ping failed: %v\nOutput: %s", err, string(output))
			}
			return string(output), nil
		})

	// Tool for DNS diagnostics
	server.Tool("dns-status", "Get DNS diagnostic information", json.RawMessage(`{
		"type": "object",
		"properties": {}
	}`),
		func(ctx context.Context, args map[string]interface{}) (string, error) {
			cmd := exec.Command("tailscale", "dns", "status")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("failed to get DNS status: %v", err)
			}
			return string(output), nil
		})
}

func addTailscalePrompts(server *mcp.MCPServer) {
	// Prompt for network connectivity diagnosis
	server.Prompt("diagnose-network", "Diagnose Tailscale network connectivity issues",
		[]mcp.PromptArgument{},
		func(ctx context.Context, args map[string]interface{}) ([]mcp.PromptMessage, error) {
			// Run netcheck command to gather data
			cmd := exec.Command("tailscale", "netcheck")
			netcheckOutput, err := cmd.CombinedOutput()
			if err != nil {
				return nil, fmt.Errorf("failed to run netcheck: %v", err)
			}

			// Get status data as well
			cmd = exec.Command("tailscale", "status")
			statusOutput, err := cmd.CombinedOutput()
			if err != nil {
				return nil, fmt.Errorf("failed to get status: %v", err)
			}

			promptText := fmt.Sprintf(`Please analyze the following Tailscale network diagnostic information and provide insights:

## Network Check Results:
%s

## Tailscale Status:
%s

What does this information tell us about the Tailscale network connectivity? Are there any issues that need to be addressed?`,
				string(netcheckOutput), string(statusOutput))

			content, _ := json.Marshal(mcp.TextContent{
				Type: "text",
				Text: promptText,
			})

			return []mcp.PromptMessage{
				{Role: "user", Content: content},
			}, nil
		})

	// Prompt for tailnet peer analysis
	server.Prompt("analyze-peers", "Get information about peers in your tailnet",
		[]mcp.PromptArgument{},
		func(ctx context.Context, args map[string]interface{}) ([]mcp.PromptMessage, error) {
			cmd := exec.Command("tailscale", "status")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return nil, fmt.Errorf("failed to get peers: %v", err)
			}

			promptText := fmt.Sprintf(`Please analyze the following peers in my Tailscale network:

%s

Can you provide a summary of the devices in my tailnet, their connection status, and any notable information?`, string(output))

			content, _ := json.Marshal(mcp.TextContent{
				Type: "text",
				Text: promptText,
			})

			return []mcp.PromptMessage{
				{Role: "user", Content: content},
			}, nil
		})

	// Prompt for exit node recommendations
	server.Prompt("exit-node-recommendations", "Get recommendations for exit nodes",
		[]mcp.PromptArgument{},
		func(ctx context.Context, args map[string]interface{}) ([]mcp.PromptMessage, error) {
			cmd := exec.Command("tailscale", "exit-node", "list")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return nil, fmt.Errorf("failed to list exit nodes: %v", err)
			}

			promptText := fmt.Sprintf(`Here are the available exit nodes in my tailnet:

%s

Based on this information, what would you recommend for choosing an exit node? Please explain the factors to consider when selecting an exit node and provide specific recommendations based on the available nodes.`, string(output))

			content, _ := json.Marshal(mcp.TextContent{
				Type: "text",
				Text: promptText,
			})

			return []mcp.PromptMessage{
				{Role: "user", Content: content},
			}, nil
		})
}
