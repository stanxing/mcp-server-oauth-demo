# MCP OAuth Demo

[English](README.md) | 简体中文

一个受 OAuth 保护的极简 MCP server,用于针对 MCP 客户端(例如 MCP
Inspector)测试 OAuth 2.1 授权服务器。所有授权服务器相关信息都通过必填/可选的
环境变量提供,因此可以对接任意符合标准的授权服务器。

该 MCP server 暴露了三个工具(`add_note`、`list_notes`、`clear_notes`),
由一个内存中的、按用户隔离的笔记存储支撑。笔记按 OAuth access token 的 `sub`
claim 隔离。

## 这个 Demo 验证了什么

1. 客户端(例如 MCP Inspector)在没有 token 的情况下调用 MCP server。
2. Server 返回 `401`,并通过 `WWW-Authenticate` 指向其
   `/.well-known/oauth-protected-resource` 元数据。
3. 客户端从该元数据的 `authorization_servers` 字段中找到授权服务器,再向
   授权服务器的 `/.well-known/oauth-authorization-server`(RFC 8414)或
   `/.well-known/openid-configuration`(OIDC Discovery)路径发起请求,发现
   其自身的 OAuth 元数据。
4. 客户端将用户重定向到授权服务器进行登录,并同意请求的 scopes。
5. 授权服务器签发一个受众(audience)指向该 MCP server 资源的 access
   token。
6. 该 server 在执行工具之前会校验 token 的签名、issuer、audience、过期时间、
   subject 以及 scopes。

## 配置

将 `.env.example` 复制为 `.env` 并填写必填项:

```sh
cp .env.example .env
```

| 变量名                  | 是否必填 | 说明                                                                    |
| ----------------------- | -------- | ----------------------------------------------------------------------- |
| `OAUTH_ISSUER_URL`      | 是       | 你的授权服务器的 Issuer URL(例如你的 auth-center 实例)。                |
| `OAUTH_SIGNING_ALGS`    | 否       | 允许的签名算法。默认 `RS256`。                                          |
| `OAUTH_REQUIRED_SCOPES` | 否       | 调用工具所需的 scopes。默认 `notes:read notes:write`。                  |
| `MCP_RESOURCE_URL`      | 否       | 该 server 的 resource/audience 标识。默认 `http://localhost:8080/mcp`。 |
| `OAUTH_RESOURCE_NAME`   | 否       | 在 protected-resource 元数据中展示的可读名称。                          |
| `MCP_SERVER_NAME`       | 否       | 通过 MCP 协议上报的实现名称。                                           |
| `MCP_SERVER_VERSION`    | 否       | 通过 MCP 协议上报的实现版本号。                                         |
| `LISTEN_ADDR`           | 否       | Server 监听的地址。默认 `:8080`。                                       |

如果没有设置 `OAUTH_ISSUER_URL`,server 会拒绝启动 —— 这里刻意没有内置默认的
授权服务器。启动时它会对
`{OAUTH_ISSUER_URL}/.well-known/openid-configuration` 发起 OIDC discovery,
并使用返回结果中的 `jwks_uri` 来校验 token 签名,因此不需要单独配置 JWKS
URL。这要求你的授权服务器提供的 discovery 文档中的 `iss` 与
`OAUTH_ISSUER_URL` 完全一致。

需要在你的授权服务器上注册一个 API/resource,其标识需与 `MCP_RESOURCE_URL`
完全一致,scopes 与 `OAUTH_REQUIRED_SCOPES` 匹配。

## 针对 MCP Inspector 运行

1. 准备一个支持 OAuth 2.1 的授权服务器(例如 Auth0)。按照上面
   [配置](#配置) 一节注册 API/resource,然后再注册一个应用并记下它的
   client ID 和 client secret —— 如果你的授权服务器不支持 Dynamic Client
   Registration,MCP Inspector 需要这两个值来走通 OAuth 流程。

2. 通过 Docker Compose 启动该 server:

   ```sh
   docker compose up --build mcp
   ```

3. 启动 MCP Inspector:

   ```sh
   npx @modelcontextprotocol/inspector
   ```

4. 打开 <http://localhost:6274/>。在 **Transport Type** 中选择
   `Streamable HTTP`,并将 URL 设置为
   `http://localhost:${MCP_PORT:-3068}/mcp`。在 **OAuth 2.0 Flow** 部分
   填入第 1 步得到的 client ID/secret。点击 Connect —— Inspector 应该会收到
   `401` 挑战,并引导你完成 discovery 与授权跳转。

笔记只存在于内存中,`mcp` 容器重启后就会消失。
