# Web Client Architecture Migration

## Overview

This document outlines the architectural changes needed to migrate the web client from using the Express backend's REST APIs to communicating directly with the MCP server via JSON-RPC, matching the CLI's architecture.

## Current vs Desired Architecture

### Current Architecture (Development Mode)

```
┌─────────────┐
│   Browser   │
│  (Port 3000)│
└──────┬──────┘
       │ REST APIs (/api/*)
       ▼
┌─────────────┐     ┌────────────┐
│   Express   │────▶│ MCP Server │
│   Backend   │     │ (Port 8080)│
│  (Port 3001)│◀────│            │
└─────────────┘     └────────────┘
   JSON-RPC

Problem: Web client depends on Express backend for:
- Authentication (/api/login, /api/session, /api/logout)
- LLM provider/model selection (/api/llm/providers, /api/llm/models)
- Chat agentic loop (/api/chat - server-side implementation)
- System info (/api/mcp/system-info)
- Conversation history (Express session management)
```

### Desired Architecture (Matches CLI)

```
┌─────────────┐
│   Browser   │
│  (Port 3000)│
└──────┬──────┘
       │ JSON-RPC (/mcp/v1)
       │ + Bearer token auth
       ▼
┌────────────┐
│ MCP Server │
│ (Port 8080)│
└────────────┘

Benefits:
- Web client talks directly to MCP server (like CLI does)
- No dependency on Express backend
- Agentic loop implemented client-side (in React)
- Session token stored in browser (localStorage)
- Simpler deployment (2 containers instead of 3 services)
```

## CLI Implementation Reference

The CLI demonstrates the correct pattern. Key files:

### Authentication via JSON-RPC Tool

**File**: [internal/chat/client.go:166-232](../internal/chat/client.go#L166-L232)

```go
// Authenticate with username/password via authenticate_user tool
func (c *Client) authenticateUser(ctx context.Context, username, password string) (string, error) {
    // Create temporary HTTP client without token
    tempClient := NewHTTPClient(baseURL, "")

    // Call authenticate_user tool via JSON-RPC
    args := map[string]interface{}{
        "username": username,
        "password": password,
    }

    response, err := tempClient.CallTool(ctx, "authenticate_user", args)

    // Parse response to extract session token
    var authResult struct {
        Success      bool   `json:"success"`
        SessionToken string `json:"session_token"`
        ExpiresAt    string `json:"expires_at"`
        Message      string `json:"message"`
    }

    json.Unmarshal([]byte(response.Content[0].Text), &authResult)
    return authResult.SessionToken, nil
}
```

### JSON-RPC HTTP Client

**File**: [internal/chat/mcp_client.go:233-377](../internal/chat/mcp_client.go#L233-L377)

```go
// Send JSON-RPC request to MCP server
func (c *httpClient) sendRequest(ctx context.Context, method string, params interface{}, result interface{}) error {
    // Build JSON-RPC request
    req := mcp.JSONRPCRequest{
        JSONRPC: "2.0",
        ID:      c.requestID,
        Method:  method,
        Params:  params,
    }

    // Create HTTP request
    httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewBuffer(reqData))
    httpReq.Header.Set("Content-Type", "application/json")

    // Add Bearer token if available
    if c.token != "" {
        httpReq.Header.Set("Authorization", "Bearer "+c.token)
    }

    // Send request and parse JSON-RPC response
    resp, _ := c.client.Do(httpReq)
    var jsonResp mcp.JSONRPCResponse
    json.NewDecoder(resp.Body).Decode(&jsonResp)

    if jsonResp.Error != nil {
        return fmt.Errorf("RPC error %d: %s", jsonResp.Error.Code, jsonResp.Error.Message)
    }

    // Extract result
    json.Unmarshal(jsonResp.Result, result)
    return nil
}
```

### Client-Side Agentic Loop

**File**: [internal/chat/client.go:379-490](../internal/chat/client.go#L379-L490)

```go
// Process query through agentic loop
func (c *Client) processQuery(ctx context.Context, query string) error {
    // Add user message to conversation history
    c.messages = append(c.messages, Message{
        Role:    "user",
        Content: query,
    })

    // Agentic loop (max 10 iterations)
    for iteration := 0; iteration < 10; iteration++ {
        // Get response from LLM with tools
        response, err := c.llm.Chat(ctx, c.messages, c.tools)

        // Check if LLM wants to use tools
        if response.StopReason == "tool_use" {
            // Extract tool uses
            var toolUses []ToolUse
            for _, item := range response.Content {
                if toolUse, ok := item.(ToolUse); ok {
                    toolUses = append(toolUses, toolUse)
                }
            }

            // Add assistant's message to history
            c.messages = append(c.messages, Message{
                Role:    "assistant",
                Content: response.Content,
            })

            // Execute all tool calls via JSON-RPC
            toolResults := []ToolResult{}
            for _, toolUse := range toolUses {
                result, err := c.mcp.CallTool(ctx, toolUse.Name, toolUse.Input)

                toolResults = append(toolResults, ToolResult{
                    Type:      "tool_result",
                    ToolUseID: toolUse.ID,
                    Content:   result.Content,
                    IsError:   result.IsError,
                })
            }

            // Add tool results to conversation
            c.messages = append(c.messages, Message{
                Role:    "user",
                Content: toolResults,
            })

            // Continue loop to get final response
            continue
        }

        // Got final response - display it
        c.ui.PrintAssistantResponse(finalText)

        // Add to history
        c.messages = append(c.messages, Message{
            Role:    "assistant",
            Content: finalText,
        })

        return nil
    }
}
```

## Required Web Client Changes

### 1. Create JSON-RPC Client Module

**New File**: `web/src/lib/mcp-client.js`

```javascript
/**
 * MCP Client - JSON-RPC communication with MCP server
 */
export class MCPClient {
    constructor(baseURL, token) {
        this.baseURL = baseURL;
        this.token = token;
        this.requestID = 0;
    }

    /**
     * Send JSON-RPC request to MCP server
     */
    async sendRequest(method, params = null) {
        this.requestID++;

        const request = {
            jsonrpc: '2.0',
            id: this.requestID,
            method: method,
            params: params || {}
        };

        const response = await fetch(this.baseURL, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                ...(this.token && { 'Authorization': `Bearer ${this.token}` })
            },
            body: JSON.stringify(request)
        });

        if (!response.ok) {
            throw new Error(`HTTP error ${response.status}: ${response.statusText}`);
        }

        const jsonResp = await response.json();

        if (jsonResp.error) {
            throw new Error(`RPC error ${jsonResp.error.code}: ${jsonResp.error.message}`);
        }

        return jsonResp.result;
    }

    /**
     * Initialize MCP connection
     */
    async initialize() {
        return await this.sendRequest('initialize', {
            protocolVersion: '2024-11-05',
            capabilities: {},
            clientInfo: {
                name: 'pgedge-mcp-web',
                version: '1.0.0'
            }
        });
    }

    /**
     * List available tools
     */
    async listTools() {
        const result = await this.sendRequest('tools/list');
        return result.tools;
    }

    /**
     * Call a tool
     */
    async callTool(name, args) {
        return await this.sendRequest('tools/call', {
            name: name,
            arguments: args
        });
    }

    /**
     * List available resources
     */
    async listResources() {
        const result = await this.sendRequest('resources/list');
        return result.resources;
    }

    /**
     * Read a resource
     */
    async readResource(uri) {
        return await this.sendRequest('resources/read', {
            uri: uri
        });
    }

    /**
     * Authenticate user and get session token
     */
    static async authenticate(baseURL, username, password) {
        // Create temporary client without token
        const tempClient = new MCPClient(baseURL, null);

        // Call authenticate_user tool
        const response = await tempClient.callTool('authenticate_user', {
            username: username,
            password: password
        });

        // Parse result
        if (!response.content || response.content.length === 0) {
            throw new Error('Invalid credentials');
        }

        const authResult = JSON.parse(response.content[0].text);

        if (!authResult.success || !authResult.session_token) {
            throw new Error(authResult.message || 'Authentication failed');
        }

        return authResult.session_token;
    }
}
```

### 2. Update AuthContext to Use JSON-RPC

**File**: `web/src/contexts/AuthContext.jsx`

```jsx
import { MCPClient } from '../lib/mcp-client';

const MCP_SERVER_URL = '/mcp/v1';  // Proxied through nginx

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [sessionToken, setSessionToken] = useState(() => {
        // Load session token from localStorage
        return localStorage.getItem('mcp-session-token');
    });
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        checkAuth();
    }, []);

    const checkAuth = async () => {
        try {
            if (!sessionToken) {
                setLoading(false);
                return;
            }

            // Validate session by trying to list tools
            const client = new MCPClient(MCP_SERVER_URL, sessionToken);
            await client.listTools();

            // Session is valid - extract username from token if needed
            // For now, just mark as authenticated
            setUser({ authenticated: true });
        } catch (error) {
            console.error('Auth check failed:', error);
            // Invalid session - clear it
            setSessionToken(null);
            localStorage.removeItem('mcp-session-token');
            setUser(null);
        } finally {
            setLoading(false);
        }
    };

    const login = async (username, password) => {
        // Authenticate via JSON-RPC
        const token = await MCPClient.authenticate(MCP_SERVER_URL, username, password);

        // Store session token
        setSessionToken(token);
        localStorage.setItem('mcp-session-token', token);
        setUser({ username });
    };

    const logout = () => {
        setSessionToken(null);
        localStorage.removeItem('mcp-session-token');
        setUser(null);
    };

    return (
        <AuthContext.Provider value={{ user, sessionToken, loading, login, logout }}>
            {children}
        </AuthContext.Provider>
    );
};
```

### 3. Create LLM Client Module

**New File**: `web/src/lib/llm-client.js`

```javascript
/**
 * LLM Client - Communicates with various LLM providers
 * Supports Anthropic, OpenAI, and Ollama
 */
export class LLMClient {
    constructor(provider, apiKey, model, ollamaURL = null) {
        this.provider = provider;
        this.apiKey = apiKey;
        this.model = model;
        this.ollamaURL = ollamaURL;
    }

    /**
     * Send chat request to LLM with tool support
     */
    async chat(messages, tools) {
        if (this.provider === 'anthropic') {
            return await this.chatAnthropic(messages, tools);
        } else if (this.provider === 'openai') {
            return await this.chatOpenAI(messages, tools);
        } else if (this.provider === 'ollama') {
            return await this.chatOllama(messages, tools);
        } else {
            throw new Error(`Unsupported provider: ${this.provider}`);
        }
    }

    async chatAnthropic(messages, tools) {
        // Implementation for Anthropic's API
        // ...
    }

    async chatOpenAI(messages, tools) {
        // Implementation for OpenAI's API
        // ...
    }

    async chatOllama(messages, tools) {
        // Implementation for Ollama's API
        // ...
    }
}
```

### 4. Update ChatInterface for Client-Side Agentic Loop

**File**: `web/src/components/ChatInterface.jsx`

Major changes needed:

```jsx
import { MCPClient } from '../lib/mcp-client';
import { LLMClient } from '../lib/llm-client';

const ChatInterface = () => {
    const { sessionToken } = useAuth();
    const [tools, setTools] = useState([]);
    const mcpClient = useMemo(() =>
        new MCPClient('/mcp/v1', sessionToken),
        [sessionToken]
    );

    // Initialize MCP and fetch tools
    useEffect(() => {
        const init = async () => {
            await mcpClient.initialize();
            const toolsList = await mcpClient.listTools();
            setTools(toolsList);
        };
        init();
    }, [mcpClient]);

    // Client-side agentic loop
    const processQuery = async (userMessage) => {
        // Add user message to conversation
        const newMessages = [...messages, {
            role: 'user',
            content: userMessage
        }];
        setMessages(newMessages);

        // Create LLM client (API keys would come from config)
        const llmClient = new LLMClient(selectedProvider, apiKey, selectedModel);

        // Agentic loop
        let conversationHistory = [...newMessages];
        for (let iteration = 0; iteration < 10; iteration++) {
            // Get LLM response
            const response = await llmClient.chat(conversationHistory, tools);

            if (response.stopReason === 'tool_use') {
                // Extract tool uses
                const toolUses = response.content.filter(item => item.type === 'tool_use');

                // Add assistant message with tool uses
                conversationHistory.push({
                    role: 'assistant',
                    content: response.content
                });

                // Execute tools via MCP JSON-RPC
                const toolResults = [];
                for (const toolUse of toolUses) {
                    const result = await mcpClient.callTool(toolUse.name, toolUse.input);
                    toolResults.push({
                        type: 'tool_result',
                        tool_use_id: toolUse.id,
                        content: result.content,
                        is_error: result.isError
                    });
                }

                // Add tool results
                conversationHistory.push({
                    role: 'user',
                    content: toolResults
                });

                // Continue loop
                continue;
            }

            // Got final response
            const finalText = response.content
                .filter(item => item.type === 'text')
                .map(item => item.text)
                .join('\n');

            conversationHistory.push({
                role: 'assistant',
                content: finalText
            });

            setMessages(conversationHistory);
            break;
        }
    };
};
```

### 5. Update StatusBanner to Use JSON-RPC

**File**: `web/src/components/StatusBanner.jsx`

```jsx
import { MCPClient } from '../lib/mcp-client';

const StatusBanner = () => {
    const { sessionToken } = useAuth();
    const [systemInfo, setSystemInfo] = useState(null);

    useEffect(() => {
        const fetchSystemInfo = async () => {
            try {
                const client = new MCPClient('/mcp/v1', sessionToken);
                const resource = await client.readResource('pg://system_info');

                // Parse system info from resource content
                const info = JSON.parse(resource.contents[0].text);
                setSystemInfo(info);
            } catch (error) {
                console.error('Failed to fetch system info:', error);
            }
        };

        fetchSystemInfo();
    }, [sessionToken]);
};
```

## Configuration Changes

### LLM API Keys in Web Client

Since the agentic loop moves to the client side, LLM API keys need to be configured. Options:

1. **Environment variables (server-side)**: Pass keys to web client via config endpoint
2. **User input**: Prompt users to enter their own API keys
3. **Proxy through MCP server**: Add LLM proxy endpoints to MCP server (keeps keys server-side)

**Recommended**: Option 3 - Add `/api/llm/chat` endpoint to MCP server that proxies to LLM providers, keeping API keys secure server-side while allowing client-side agentic loop.

### Nginx Configuration

Update nginx to proxy `/mcp/v1` instead of `/api/*`:

```nginx
location /mcp/v1 {
    proxy_pass http://mcp-server:8080/mcp/v1;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header Authorization $http_authorization;
}
```

## Migration Steps

1. **Create JSON-RPC client library** (`web/src/lib/mcp-client.js`)
2. **Create LLM client library** (`web/src/lib/llm-client.js`) or proxy endpoint
3. **Update AuthContext** to use `authenticate_user` tool via JSON-RPC
4. **Update ChatInterface** to implement client-side agentic loop
5. **Update StatusBanner** to read `pg://system_info` resource via JSON-RPC
6. **Remove REST API dependencies** from all components
7. **Test authentication flow** with JSON-RPC
8. **Test chat agentic loop** with tool calls
9. **Update Docker deployment** to remove Express backend
10. **Update documentation**

## Benefits of Migration

1. **Architectural Consistency**: Web client matches CLI architecture
2. **Simpler Deployment**: 2 containers (MCP + Web) instead of 3 services
3. **No Express Dependency**: Express only used for development/debugging
4. **Better Security**: Direct token-based auth with MCP server
5. **More Flexibility**: Client-side agentic loop allows for UI enhancements
6. **Easier Maintenance**: Single authentication mechanism for all clients

## Testing Checklist

- [ ] Login via `authenticate_user` tool returns session token
- [ ] Session token stored in localStorage
- [ ] Session token sent in Authorization header for all requests
- [ ] Tools can be listed via `tools/list`
- [ ] Resources can be listed via `resources/list`
- [ ] System info can be read via `resources/read`
- [ ] Chat sends message to LLM with tools
- [ ] LLM tool_use triggers MCP `tools/call` via JSON-RPC
- [ ] Tool results added to conversation history
- [ ] Final response displayed to user
- [ ] Session validation on app load
- [ ] Logout clears session token
- [ ] Invalid session redirects to login

## Files to Modify

### New Files

- `web/src/lib/mcp-client.js` - JSON-RPC client
- `web/src/lib/llm-client.js` - LLM client (or proxy endpoint)
- `docs/web-client-architecture-migration.md` - This document

### Modified Files

- `web/src/contexts/AuthContext.jsx` - Use JSON-RPC auth
- `web/src/components/ChatInterface.jsx` - Client-side agentic loop
- `web/src/components/StatusBanner.jsx` - Read resource via JSON-RPC
- `web/src/components/Login.jsx` - May need updates
- `docker/nginx.conf` - Proxy `/mcp/v1` instead of `/api/*`
- `docker/docker-compose.yml` - Remove Express backend service
- `docker/Dockerfile.web` - Simplified build

### Files to Remove (Eventually)

- `web/server.js` - Express backend (only used for development/debugging)
- `web/lib/chat-agent.js` - Server-side agentic loop
- `web/lib/config-loader.js` - Server-side config

## Notes

- Express backend (`web/server.js`) should remain available for development/debugging
- The `start_web_client.sh` script should continue to work for local development
- Docker deployment should use the new JSON-RPC architecture
- All REST API endpoints in the Go server (`internal/mcp/rest_api.go`) can be removed once migration is complete
