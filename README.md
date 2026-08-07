# MCP OAuth Demo

English | [简体中文](README.zh-CN.md)

A minimal OAuth-protected MCP server for testing an OAuth 2.1 authorization
server against an MCP client (such as MCP Inspector).
All authorization server details are supplied via required/optional
environment variables, so this can point at any standards-compliant
authorization server.

The MCP server exposes three tools (`add_note`, `list_notes`, `clear_notes`)
backed by an in-memory, per-user note store. Notes are isolated by the OAuth
access token's `sub` claim.

## What it proves

1. Client (e.g. MCP Inspector) calls the MCP server without a token.
2. Server responds `401` with `WWW-Authenticate` pointing at its
   `/.well-known/oauth-protected-resource` metadata.
3. Client reads the `authorization_servers` field from that metadata to find
   the authorization server, then discovers the authorization server's own
   OAuth metadata at `/.well-known/oauth-authorization-server` (RFC 8414) or
   `/.well-known/openid-configuration` (OIDC Discovery).
4. Client redirects the user to the authorization server to sign in and
   approve the requested scopes.
5. Authorization server issues an access token audienced to this MCP
   server's resource URL.
6. This server verifies the token's signature, issuer, audience, expiry,
   subject and scopes before running a tool.

## Configuration

Copy `.env.example` to `.env` and fill in the required values:

```sh
cp .env.example .env
```

| Variable                | Required | Description                                                                      |
| ----------------------- | -------- | -------------------------------------------------------------------------------- |
| `OAUTH_ISSUER_URL`      | yes      | Issuer URL of your authorization server (e.g. your auth-center instance).        |
| `OAUTH_SIGNING_ALGS`    | no       | Allowed signing algorithms. Default `RS256`.                                     |
| `OAUTH_REQUIRED_SCOPES` | no       | Scopes required to call the tools. Default `notes:read notes:write`.             |
| `MCP_RESOURCE_URL`      | no       | This server's resource/audience identifier. Default `http://localhost:8080/mcp`. |
| `OAUTH_RESOURCE_NAME`   | no       | Human-readable name in the protected-resource metadata.                          |
| `MCP_SERVER_NAME`       | no       | MCP implementation name reported over the protocol.                              |
| `MCP_SERVER_VERSION`    | no       | MCP implementation version reported over the protocol.                           |
| `LISTEN_ADDR`           | no       | Address the server binds to. Default `:8080`.                                    |

The server refuses to start if `OAUTH_ISSUER_URL` is unset — there is
intentionally no default authorization server baked in. On startup it runs
OIDC discovery against `{OAUTH_ISSUER_URL}/.well-known/openid-configuration`
and uses the `jwks_uri` it returns to verify token signatures, so there's no
separate JWKS URL to configure. This requires your authorization server to
serve that discovery document with an `iss` matching `OAUTH_ISSUER_URL`
exactly.

On your authorization server, register an API/resource whose identifier
exactly matches `MCP_RESOURCE_URL`, with scopes matching
`OAUTH_REQUIRED_SCOPES`.

## Running against MCP Inspector

1. Set up an OAuth 2.1-capable authorization server (e.g. Auth0). Register
   the API/resource as described in [Configuration](#configuration) above,
   then register an application and note its client ID and client secret —
   MCP Inspector needs these to drive the OAuth flow if your authorization
   server doesn't support Dynamic Client Registration.

2. Start this server via Docker Compose:

   ```sh
   docker compose up --build mcp
   ```

3. Start MCP Inspector:

   ```sh
   npx @modelcontextprotocol/inspector
   ```

4. Open <http://localhost:6274/>. Under **Transport Type**, choose
   `Streamable HTTP` and set the URL to
   `http://localhost:${MCP_PORT:-3068}/mcp`. Under **OAuth 2.0 Flow**, enter
   the client ID/secret from step 1. Click Connect — Inspector should receive
   the `401` challenge and walk you through discovery and the authorization
   redirect.

Notes exist only in memory and disappear when the `mcp` container restarts.
