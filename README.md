# Tailscale Model Context Protocol (MCP) Server

A simple [MCP](https://modelcontextprotocol.io/introduction) server that
provides read-only access to your [Tailscale](https://tailscale.com/) network
directly from Claude Desktop and other MCP-compatible clients.

> [!CAUTION]
> You might not want to do this! This server exposes your Tailscale network to
> an external application. It invokes the `tailscale` binary on your system on
> your behalf, assembling an argument list through string concatenation, and
> executing as your logged-in Tailscale account. While its operation is
> intended to be read-only and therefore "secure", you should be aware of the
> potential risks involved in exposing any part of your network to third-party
> applications. [Especially when interacting with
> LLMs](https://simonwillison.net/search/?tag=prompt-injection).

## Description

This server allows, for example, Claude to interact with your Tailscale network
by exposing read-only commands as tools and prompts. It enables you to:

- Check your Tailscale status and connected devices
- Get network diagnostics
- View your Tailscale IP addresses
- List available exit nodes
- Ping Tailscale hosts
- Look up information about Tailscale IPs

## Requirements

- [Go](https://golang.org/doc/install) (for building from source)
- [Tailscale CLI](https://tailscale.com/download) must be installed and
  accessible in your `$PATH`
- An MCP-compatible client like Claude Desktop

## Installation

### Pre-built Binary

Coming soon

### Building from Source

```bash
go install github.com/paulsmith/tailscale-mcp@latest
```

## Using with Claude Desktop

1. First, make sure you have [Claude Desktop](https://claude.ai/download)
   installed and updated to the latest version

2. Open your Claude Desktop configuration file:
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`

3. Add the Tailscale MCP server configuration:

```json
{
  "mcpServers": {
    "tailscale": {
      "command": "/path/to/tailscale-mcp"
    }
  }
}
```

4. Replace `/path/to/tailscale-mcp` with the actual path to the binary

5. Restart Claude Desktop

## Available Tools

The server exposes the following tools:

- **tailscale**: Run any "safe" Tailscale command
- **get-ip**: Get your Tailscale IP addresses
- **get-status**: Get information about your Tailscale network
- **network-check**: Check Tailscale network connectivity
- **list-exit-nodes**: List available Tailscale exit nodes
- **ip-lookup**: Look up information about a Tailscale IP
- **ping-host**: Ping a Tailscale host
- **dns-status**: Get DNS diagnostic information

## The List of Safe Subcommands

The list of subcommands that are considered "safe" and can be run through the
`tailscale` tool that this MCP server uses are:

- `netcheck`
- `ip`
- `dns status`
- `status`
- `metrics`
- `ping <peer>`
- `version`
- `exit-node list`
- `whois <ip>`

## Available Prompts

The server also includes several prompts to help with common tasks:

- **diagnose-network**: Analyze Tailscale network connectivity issues
- **analyze-peers**: Get a summary of devices in your tailnet
- **exit-node-recommendations**: Get recommendations for exit nodes

## Example Usage

Once connected to Claude Desktop, you can ask questions like:

- "What's my Tailscale IP address?"
- "Show me all the devices connected to my Tailscale network"
- "Can you check if my Tailscale network connection is working properly?"
- "Ping my device called 'laptop'"
- "Are there any exit nodes available in my network?"
- "What DNS settings is Tailscale using?"

### Example Chat

> Describe my tailnet and note anything out of the ordinary.

```markdown
I'll check your Tailscale network status and provide you with information about
your tailnet. Let me run a few commands to gather the details.

First, let's check the overall status of your tailnet:

View result from get-status

Now, let's get the Tailscale IP addresses:

View result from get-ip

Let's check the network connectivity status:

View result from network-check

Let's check for available exit nodes in your tailnet:

View result from list-exit-nodes

Let's also check the DNS status:

View result from dns-status

Based on all the information gathered, here's a description of your tailnet:

Your tailnet is named "orca-arctic.ts.net" and currently has 11 devices
registered, with 6 devices currently online and 5 offline. Your current device
is "venus" running macOS with the IP address 100.86.102.37.

Here's what I've observed about your tailnet:

- You have one device ("bunny") configured as an exit node. This allows other devices to route their traffic through it.
- There's a version mismatch between your client (1.80.3) and the tailscaled server (1.82.0), which could be worth updating to ensure compatibility and security.
- You have a device named "hello.ts.net" with a different user account (hello@) compared to your other devices (paulsmith@). This might be unusual if you're not expecting another user in your tailnet.
- Your network connectivity looks good with both IPv4 and IPv6 support, and you're using UPnP for port mapping.
- Your nearest DERP (Designated Encrypted Relay for Packets) server is in Ashburn with a 71.9ms latency, which is quite good.
- MagicDNS is enabled tailnet-wide, allowing you to reach devices by their hostname (e.g., venus.orca-arctic.ts.net).
- You're using NextDNS as your DNS resolver, which is a custom configuration.

Notable items that might need attention:

- The client/server version mismatch mentioned above
- The device "hello.ts.net" with a different user account (hello@) that might need verification if it's not expected
- Five of your devices are currently offline, which is normal if they're not in use but worth checking if you expect them to be online

Would you like me to provide more information about any specific aspect of your tailnet?
```

## Security Notes

- This server allows read-only access to your Tailscale network
- Only subcommands of the `tailscale` CLI in the "safe" whitelist are permitted
- No configuration changes can be made (in theory!)
- All commands are executed with your user's permissions
- You are exposing your Tailscale network to an LLM, however indirectly via the
  MCP RPC, which is subject to security risks such as prompt injection -
  educate yourself

## Troubleshooting

### Server not appearing in Claude Desktop

Check the following:

1. Make sure the path to the binary in your configuration is correct
2. Verify Tailscale CLI is installed and accessible in your `$PATH`
3. Check Claude Desktop logs for errors:
   - macOS: `~/Library/Logs/Claude/mcp*.log`
   - Windows: `%APPDATA%\Claude\logs\mcp*.log`

### Command errors

If commands are failing, try:

1. Running the command directly using the Tailscale CLI to verify it works
2. Check that your Tailscale is correctly configured and connected
3. Ensure the command is in the allowed safe list

## Development

This server is built using the MCP Go SDK. If you want to extend or modify it:

1. Clone the repository
2. Make your changes
3. Build using `go build`
4. Test with Claude Desktop or other MCP clients

## Contributing

Contributions are welcome! Please feel free to open an issue and/or submit a
pull request.

## License

[MIT License](COPYING)
