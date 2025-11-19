/*-------------------------------------------------------------------------
 *
 * pgEdge MCP Client - AuthContext Tests
 *
 * Copyright (c) 2025, pgEdge, Inc.
 * This software is released under The PostgreSQL License
 *
 *-------------------------------------------------------------------------
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { AuthProvider, useAuth } from '../AuthContext';

describe('AuthContext', () => {
  beforeEach(() => {
    global.fetch = vi.fn();
  });

  it('provides initial unauthenticated state', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        authenticated: false,
      }),
    });

    const { result } = renderHook(() => useAuth(), {
      wrapper: AuthProvider,
    });

    expect(result.current.user).toBe(null);
    expect(result.current.loading).toBe(true); // Loading while checking auth status

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.user).toBe(null);
  });

  it('checks authentication status on mount', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        authenticated: true,
        user: 'testuser',
      }),
    });

    const { result } = renderHook(() => useAuth(), {
      wrapper: AuthProvider,
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.user).toBe('testuser');
    expect(global.fetch).toHaveBeenCalledWith('/api/session', expect.any(Object));
  });

  it('handles login successfully', async () => {
    // Mock initial auth status check
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        authenticated: false,
      }),
    });

    const { result } = renderHook(() => useAuth(), {
      wrapper: AuthProvider,
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Mock login response
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        user: 'testuser',
      }),
    });

    await result.current.login('testuser', 'testpass');

    await waitFor(() => {
      expect(result.current.user).toBe('testuser');
    });
  });

  it('throws error on login failure', async () => {
    // Mock initial auth status check
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        authenticated: false,
      }),
    });

    const { result } = renderHook(() => useAuth(), {
      wrapper: AuthProvider,
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Mock login failure
    global.fetch.mockResolvedValueOnce({
      ok: false,
      json: async () => ({
        message: 'Invalid credentials',
      }),
    });

    await expect(result.current.login('testuser', 'wrongpass')).rejects.toThrow(
      'Invalid credentials'
    );

    expect(result.current.user).toBe(null);
  });

  it('handles logout successfully', async () => {
    // Mock initial authenticated state
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        authenticated: true,
        user: 'testuser',
      }),
    });

    const { result } = renderHook(() => useAuth(), {
      wrapper: AuthProvider,
    });

    await waitFor(() => {
      expect(result.current.user).toBe('testuser');
    });

    // Mock logout response
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({}),
    });

    await result.current.logout();

    await waitFor(() => {
      expect(result.current.user).toBe(null);
    });
  });
});
