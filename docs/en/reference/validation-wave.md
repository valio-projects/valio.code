# Boundary validation wave

Local repository discovery now checks path existence, directory type and symlink
resolution before invoking Git. Upload, MCP and authentication configuration
share bounded URL/token syntax checks; repository remote URLs reject malformed
hosts, ports and traversal. CLI upload configuration is validated before capture
or spool side effects. Upload receipts reject oversized or trailing JSON; Git
and upload cancellation preserve the caller's context error.

Tests cover missing paths, file-versus-directory mistakes, malformed URLs,
credential-bearing URLs, invalid ports, control characters, token bounds and
existing transport behavior. Paths on agent machines are never checked as local
paths by the server: Project roots remain repository-relative catalog selections.
URL validation proves syntax, not remote reachability or authentication; AI
provider probes test those separately.
