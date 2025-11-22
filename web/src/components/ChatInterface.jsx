/*-------------------------------------------------------------------------
 *
 * pgEdge MCP Client - Chat Interface (Refactored)
 *
 * Portions copyright (c) 2025, pgEdge, Inc.
 * This software is released under The PostgreSQL License
 *
 *-------------------------------------------------------------------------
 */

import React, { useState, useEffect, useCallback } from 'react';
import { Box, Paper, Alert } from '@mui/material';
import { useAuth } from '../contexts/AuthContext';
import { useLocalStorageBoolean } from '../hooks/useLocalStorage';
import { useQueryHistory } from '../hooks/useQueryHistory';
import { useMCPClient } from '../hooks/useMCPClient';
import { useLLMProviders } from '../hooks/useLLMProviders';
import MessageList from './MessageList';
import MessageInput from './MessageInput';
import ProviderSelector from './ProviderSelector';
import PromptPopover from './PromptPopover';

const MAX_AGENTIC_LOOPS = 10;

/**
 * Compacts message history to prevent token overflow by keeping only
 * the most recent messages while preserving the initial context
 * @param {Array} messages - The full message history
 * @returns {Array} - Compacted message history
 */
const compactMessages = (messages) => {
    const MAX_RECENT_MESSAGES = 10;

    // If we have fewer messages than the limit, return all
    if (messages.length <= MAX_RECENT_MESSAGES) {
        return messages;
    }

    // Strategy: Keep the first user message and the last N messages
    // This preserves the original query context while maintaining recent conversation flow
    const compacted = [];

    // Keep the first user message (original query)
    if (messages.length > 0 && messages[0].role === 'user') {
        compacted.push(messages[0]);
    }

    // Keep the last N messages
    const startIdx = Math.max(1, messages.length - MAX_RECENT_MESSAGES);
    compacted.push(...messages.slice(startIdx));

    console.log(`[Compaction] Reduced messages from ${messages.length} to ${compacted.length} (kept first + last ${MAX_RECENT_MESSAGES})`);

    return compacted;
};

const ChatInterface = () => {
    const { sessionToken, forceLogout } = useAuth();

    // State management using custom hooks
    // Initialize messages with fromPreviousSession flag for loaded messages
    const [messages, setMessages] = useState(() => {
        try {
            const savedMessages = localStorage.getItem('chat-messages');
            if (savedMessages) {
                const parsed = JSON.parse(savedMessages);
                // Mark all loaded messages as from previous session and ensure content is a string
                return parsed.map(msg => ({
                    ...msg,
                    content: typeof msg.content === 'string' ? msg.content : JSON.stringify(msg.content),
                    fromPreviousSession: true
                }));
            }
        } catch (error) {
            console.error('Error loading chat messages:', error);
        }
        return [];
    });

    const [showActivity, setShowActivity] = useLocalStorageBoolean('show-activity', true);
    const [renderMarkdown, setRenderMarkdown] = useLocalStorageBoolean('render-markdown', true);
    const [debug, setDebug] = useLocalStorageBoolean('debug', false);

    const [input, setInput] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    // Prompt popover state
    const [promptPopoverAnchor, setPromptPopoverAnchor] = useState(null);
    const [executingPrompt, setExecutingPrompt] = useState(false);

    // Custom hooks for functionality
    const queryHistory = useQueryHistory();
    const { mcpClient, tools, prompts, refreshTools, refreshPrompts } = useMCPClient(sessionToken);
    const llmProviders = useLLMProviders(sessionToken);

    // Log prompts when they're available (for debugging)
    useEffect(() => {
        if (prompts.length > 0) {
            console.log('MCP prompts available:', prompts);
        }
    }, [prompts]);

    // Save messages to localStorage when they change
    useEffect(() => {
        try {
            // Don't save if messages array is empty
            if (messages.length > 0) {
                // Remove the fromPreviousSession flag before saving
                const messagesToSave = messages.map(({ fromPreviousSession, ...msg }) => msg);
                localStorage.setItem('chat-messages', JSON.stringify(messagesToSave));
            }
        } catch (error) {
            console.error('Error saving chat messages:', error);
        }
    }, [messages]);

    // Handle message sending
    const handleSend = useCallback(async () => {
        if (!input.trim() || loading || !mcpClient) return;

        const userMessage = {
            role: 'user',
            content: input.trim(),
            timestamp: new Date().toISOString(),
        };

        // Add to history
        queryHistory.addToHistory(userMessage.content);

        // Create thinking message placeholder
        const thinkingMessage = {
            role: 'assistant',
            content: '',
            timestamp: new Date().toISOString(),
            provider: llmProviders.selectedProvider,
            model: llmProviders.selectedModel,
            activity: [],
            isThinking: true,
        };

        setMessages(prev => [...prev, userMessage, thinkingMessage]);
        setInput('');
        setLoading(true);
        setError('');

        try {
            // Build conversation history
            const conversationMessages = [];

            // Add all previous messages
            for (const msg of messages) {
                if (msg.role === 'user') {
                    conversationMessages.push({
                        role: 'user',
                        content: msg.content
                    });
                } else if (msg.role === 'assistant' && msg.content) {
                    conversationMessages.push({
                        role: 'assistant',
                        content: msg.content
                    });
                }
            }

            // Add current user message
            conversationMessages.push({
                role: 'user',
                content: userMessage.content
            });

            const activity = [];
            let loopCount = 0;

            // Agentic loop
            while (loopCount < MAX_AGENTIC_LOOPS) {
                loopCount++;

                // Compact message history to prevent token overflow
                const compactedMessages = compactMessages(conversationMessages);

                // Call LLM with compacted history
                const llmResponse = await fetch('/api/llm/chat', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${sessionToken}`,
                    },
                    credentials: 'include',
                    body: JSON.stringify({
                        messages: compactedMessages,
                        tools: tools,
                        provider: llmProviders.selectedProvider,
                        model: llmProviders.selectedModel,
                        debug: debug,
                    }),
                });

                // Handle session invalidation
                if (llmResponse.status === 401) {
                    console.log('Session invalidated, logging out...');
                    forceLogout();
                    setError('Your session has expired. Please log in again.');
                    // Remove thinking message (keep user message for context)
                    setMessages(prev => prev.slice(0, -1));
                    return;
                }

                if (!llmResponse.ok) {
                    const errorText = await llmResponse.text();
                    throw new Error(`LLM request failed: ${llmResponse.status} ${errorText}`);
                }

                const llmData = await llmResponse.json();
                console.log('LLM response:', llmData);
                console.log('Loop iteration:', loopCount, 'Stop reason:', llmData.stop_reason);
                if (llmData.stop_reason === 'tool_use') {
                    const toolUseCount = llmData.content.filter(c => c.type === 'tool_use').length;
                    console.log('Number of tool_use blocks in this response:', toolUseCount);
                }

                // Check stop reason
                if (llmData.stop_reason === 'end_turn' || loopCount >= MAX_AGENTIC_LOOPS) {
                    // Final response - extract text content
                    let textContent = '';
                    const contentArray = Array.isArray(llmData.content) ? llmData.content : [llmData.content];

                    for (const content of contentArray) {
                        if (content && content.type === 'text') {
                            const text = typeof content.text === 'string' ? content.text : String(content.text || '');
                            textContent += text;
                        }
                    }

                    const finalContent = textContent || 'No response received';

                    // Replace thinking message with final response
                    console.log('Final activity array:', activity);
                    console.log('Total tool uses tracked:', activity.length);
                    setMessages(prev => {
                        const newMessages = prev.slice(0, -1);
                        return [...newMessages, {
                            role: 'assistant',
                            content: finalContent,
                            timestamp: new Date().toISOString(),
                            provider: llmProviders.selectedProvider,
                            model: llmProviders.selectedModel,
                            activity: activity,
                            tokenUsage: llmData.token_usage,
                        }];
                    });
                    break;
                }

                // Handle tool use
                if (llmData.stop_reason === 'tool_use') {
                    const toolUses = llmData.content.filter(c => c.type === 'tool_use');

                    if (toolUses.length === 0) {
                        throw new Error('LLM indicated tool_use but no tool_use blocks found');
                    }

                    // Execute tools
                    const toolResults = [];
                    for (const toolUse of toolUses) {
                        console.log('Executing tool:', toolUse.name, 'with args:', toolUse.input);

                        // Update activity
                        activity.push({
                            type: 'tool',
                            name: toolUse.name,
                            timestamp: new Date().toISOString(),
                        });

                        // Update thinking message with new activity
                        setMessages(prev => {
                            const newMessages = [...prev];
                            if (newMessages.length > 0 && newMessages[newMessages.length - 1].isThinking) {
                                // Create a new message object instead of mutating
                                newMessages[newMessages.length - 1] = {
                                    ...newMessages[newMessages.length - 1],
                                    activity: [...activity]
                                };
                            }
                            return newMessages;
                        });

                        try {
                            // Execute tool via MCP
                            const result = await mcpClient.callTool(toolUse.name, toolUse.input);
                            console.log('Tool result:', result);

                            toolResults.push({
                                type: 'tool_result',
                                tool_use_id: toolUse.id,
                                content: result.content,
                            });

                            // Refresh tools if manage_connections was called
                            if (toolUse.name === 'manage_connections' && !result.isError) {
                                await refreshTools();
                            }
                        } catch (toolError) {
                            console.error('Tool execution error:', toolError);
                            toolResults.push({
                                type: 'tool_result',
                                tool_use_id: toolUse.id,
                                content: `Error: ${toolError.message}`,
                                is_error: true,
                            });
                        }
                    }

                    // Add assistant message with tool uses
                    conversationMessages.push({
                        role: 'assistant',
                        content: llmData.content,
                    });

                    // Add user message with tool results
                    conversationMessages.push({
                        role: 'user',
                        content: toolResults,
                    });

                    // Continue loop
                    continue;
                }

                // Unknown stop reason
                throw new Error(`Unexpected stop reason: ${llmData.stop_reason}`);
            }

            if (loopCount >= MAX_AGENTIC_LOOPS) {
                throw new Error('Maximum tool execution loops reached');
            }
        } catch (err) {
            console.error('Chat error:', err);

            // Remove thinking message (keep user message for context)
            setMessages(prev => prev.slice(0, -1));

            // Network errors
            if (err.name === 'TypeError' && err.message.includes('fetch')) {
                setError('Cannot connect to server. Please check that the server is running.');
            } else {
                setError(err.message || 'Failed to send message');
            }
        } finally {
            setLoading(false);
        }
    }, [input, loading, mcpClient, messages, sessionToken, tools, llmProviders.selectedProvider, llmProviders.selectedModel, queryHistory, forceLogout, refreshTools]);

    // Handle keyboard shortcuts
    const handleKeyDown = useCallback((e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSend();
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            const newInput = queryHistory.navigateUp(input);
            setInput(newInput);
        } else if (e.key === 'ArrowDown') {
            e.preventDefault();
            const newInput = queryHistory.navigateDown(input);
            setInput(newInput);
        }
    }, [input, queryHistory, handleSend]);

    // Handle input change
    const handleInputChange = useCallback((e) => {
        setInput(e.target.value);
        // Reset history navigation when user types
        if (queryHistory.isNavigating) {
            queryHistory.resetNavigation();
        }
    }, [queryHistory]);

    // Handle clear conversation
    const handleClear = useCallback(() => {
        if (!window.confirm('Clear conversation history?')) return;

        setMessages([]);
        queryHistory.clearHistory();
        setError('');
    }, [queryHistory]);

    // Handle prompt selection
    const handlePromptClick = useCallback((event) => {
        setPromptPopoverAnchor(event.currentTarget);
    }, []);

    // Handle prompt execution
    const handlePromptExecute = useCallback(async (promptName, args) => {
        if (!mcpClient || loading) return;

        setExecutingPrompt(true);

        try {
            // Get the prompt with arguments from MCP server
            const promptResult = await mcpClient.getPrompt(promptName, args);

            // Add a system message to indicate prompt execution
            const systemMessage = {
                role: 'system',
                content: `Executing prompt: ${promptName}`,
                timestamp: new Date().toISOString(),
            };
            setMessages(prev => [...prev, systemMessage]);

            // Build conversation history (exclude system messages)
            const conversationMessages = [];
            for (const msg of messages) {
                if (msg.role === 'user') {
                    conversationMessages.push({
                        role: 'user',
                        content: msg.content
                    });
                } else if (msg.role === 'assistant' && msg.content) {
                    conversationMessages.push({
                        role: 'assistant',
                        content: msg.content
                    });
                }
            }
            // Add prompt messages to conversation
            if (promptResult.messages) {
                for (const msg of promptResult.messages) {
                    if (msg.role === 'user') {
                        const userMsg = {
                            role: 'user',
                            content: msg.content.text,
                            timestamp: new Date().toISOString(),
                            fromPrompt: true,
                        };
                        setMessages(prev => [...prev, userMsg]);
                        // Add to conversation history (only role and content)
                        conversationMessages.push({
                            role: 'user',
                            content: msg.content.text
                        });
                    }
                }
            }

            // Create thinking message placeholder
            const thinkingMessage = {
                role: 'assistant',
                content: '',
                timestamp: new Date().toISOString(),
                provider: llmProviders.selectedProvider,
                model: llmProviders.selectedModel,
                isThinking: true,
                activity: [],
            };
            setMessages(prev => [...prev, thinkingMessage]);

            // Trigger the agentic loop with the prompt messages
            setLoading(true);

            // Start agentic loop (similar to handleSend but using prompt messages)
            let loopCount = 0;
            const activity = [];

            while (loopCount < MAX_AGENTIC_LOOPS) {
                loopCount++;

                // Compact message history to prevent token overflow
                const compactedMessages = compactMessages(conversationMessages);

                // Make LLM request with compacted history
                const response = await fetch('/api/llm/chat', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${sessionToken}`,
                    },
                    body: JSON.stringify({
                        messages: compactedMessages,
                        tools: tools,
                        provider: llmProviders.selectedProvider,
                        model: llmProviders.selectedModel,
                    }),
                });

                if (!response.ok) {
                    if (response.status === 401) {
                        forceLogout();
                        throw new Error('Session expired. Please login again.');
                    }
                    const errorText = await response.text();
                    throw new Error(`Server error: ${errorText}`);
                }

                const llmData = await response.json();

                // Handle end_turn
                if (llmData.stop_reason === 'end_turn') {
                    const finalContent = llmData.content
                        .filter(c => c.type === 'text')
                        .map(c => c.text)
                        .join('\n');

                    // Replace thinking message with actual response
                    setMessages(prev => {
                        const newMessages = prev.slice(0, -1);
                        return [...newMessages, {
                            role: 'assistant',
                            content: finalContent,
                            timestamp: new Date().toISOString(),
                            provider: llmProviders.selectedProvider,
                            model: llmProviders.selectedModel,
                            activity: activity,
                            tokenUsage: llmData.token_usage,
                        }];
                    });
                    break;
                }

                // Handle tool use
                if (llmData.stop_reason === 'tool_use') {
                    const toolUses = llmData.content.filter(c => c.type === 'tool_use');

                    if (toolUses.length === 0) {
                        throw new Error('LLM indicated tool_use but no tool_use blocks found');
                    }

                    // Execute tools
                    const toolResults = [];
                    for (const toolUse of toolUses) {
                        // Update activity
                        activity.push({
                            type: 'tool',
                            name: toolUse.name,
                            timestamp: new Date().toISOString(),
                        });

                        // Update thinking message with new activity
                        setMessages(prev => {
                            const newMessages = [...prev];
                            if (newMessages.length > 0 && newMessages[newMessages.length - 1].isThinking) {
                                newMessages[newMessages.length - 1] = {
                                    ...newMessages[newMessages.length - 1],
                                    activity: [...activity]
                                };
                            }
                            return newMessages;
                        });

                        try {
                            // Execute tool via MCP
                            const result = await mcpClient.callTool(toolUse.name, toolUse.input);

                            toolResults.push({
                                type: 'tool_result',
                                tool_use_id: toolUse.id,
                                content: result.content,
                            });

                            // Refresh tools if manage_connections was called
                            if (toolUse.name === 'manage_connections' && !result.isError) {
                                await refreshTools();
                            }
                        } catch (toolError) {
                            console.error('Tool execution error:', toolError);
                            toolResults.push({
                                type: 'tool_result',
                                tool_use_id: toolUse.id,
                                content: `Error: ${toolError.message}`,
                                is_error: true,
                            });
                        }
                    }

                    // Add assistant message with tool uses
                    conversationMessages.push({
                        role: 'assistant',
                        content: llmData.content,
                    });

                    // Add user message with tool results
                    conversationMessages.push({
                        role: 'user',
                        content: toolResults,
                    });

                    // Continue loop
                    continue;
                }

                // Unknown stop reason
                throw new Error(`Unexpected stop reason: ${llmData.stop_reason}`);
            }

            if (loopCount >= MAX_AGENTIC_LOOPS) {
                throw new Error('Maximum tool execution loops reached');
            }
        } catch (err) {
            console.error('Prompt execution error:', err);

            // Remove thinking message if present
            setMessages(prev => {
                if (prev.length > 0 && prev[prev.length - 1].isThinking) {
                    return prev.slice(0, -1);
                }
                return prev;
            });

            // Network errors
            if (err.name === 'TypeError' && err.message.includes('fetch')) {
                setError('Cannot connect to server. Please check that the server is running.');
            } else {
                setError(err.message || 'Failed to execute prompt');
            }
        } finally {
            setExecutingPrompt(false);
            setLoading(false);
        }
    }, [mcpClient, loading, messages, sessionToken, tools, llmProviders.selectedProvider, llmProviders.selectedModel, forceLogout, refreshTools]);

    return (
        <Box
            sx={{
                display: 'flex',
                flexDirection: 'column',
                flex: 1,
                minHeight: 0,
            }}
        >
            {/* Messages */}
            <MessageList
                messages={messages}
                showActivity={showActivity}
                renderMarkdown={renderMarkdown}
                debug={debug}
                onClear={handleClear}
            />

            {/* Error Display */}
            {(error || llmProviders.error) && (
                <Alert
                    severity="error"
                    sx={{ mb: 1 }}
                    onClose={() => {
                        setError('');
                        // Note: Can't clear llmProviders.error as it's from the hook
                    }}
                >
                    {error || llmProviders.error}
                </Alert>
            )}

            {/* Input Area */}
            <Paper elevation={2} sx={{ p: 2 }}>
                <MessageInput
                    value={input}
                    onChange={handleInputChange}
                    onSend={handleSend}
                    onKeyDown={handleKeyDown}
                    disabled={loading}
                    onPromptClick={handlePromptClick}
                    hasPrompts={prompts && prompts.length > 0}
                />

                <ProviderSelector
                    providers={llmProviders.providers}
                    selectedProvider={llmProviders.selectedProvider}
                    onProviderChange={llmProviders.setSelectedProvider}
                    models={llmProviders.models}
                    selectedModel={llmProviders.selectedModel}
                    onModelChange={llmProviders.setSelectedModel}
                    showActivity={showActivity}
                    onActivityChange={setShowActivity}
                    renderMarkdown={renderMarkdown}
                    onMarkdownChange={setRenderMarkdown}
                    debug={debug}
                    onDebugChange={setDebug}
                    disabled={loading}
                    loadingModels={llmProviders.loadingModels}
                />
            </Paper>

            {/* Prompt Popover */}
            <PromptPopover
                anchorEl={promptPopoverAnchor}
                open={Boolean(promptPopoverAnchor)}
                onClose={() => setPromptPopoverAnchor(null)}
                prompts={prompts}
                onExecute={handlePromptExecute}
                executing={executingPrompt}
            />
        </Box>
    );
};

export default ChatInterface;
