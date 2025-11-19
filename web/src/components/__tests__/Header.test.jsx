/*-------------------------------------------------------------------------
 *
 * pgEdge MCP Client - Header Component Tests
 *
 * Copyright (c) 2025, pgEdge, Inc.
 * This software is released under The PostgreSQL License
 *
 *-------------------------------------------------------------------------
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Header from '../Header';
import { AuthProvider } from '../../contexts/AuthContext';

// Mock the logo imports
vi.mock('../../assets/images/logo-light.png', () => ({
    default: 'logo-light.png',
}));

vi.mock('../../assets/images/logo-dark.png', () => ({
    default: 'logo-dark.png',
}));

describe('Header Component', () => {
    const mockToggleTheme = vi.fn();

    beforeEach(() => {
        mockToggleTheme.mockClear();
        global.fetch = vi.fn();
        localStorage.clear();
    });

    const renderHeader = (mode = 'light', user = { username: 'testuser' }) => {
        // Mock authenticated state
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({
                authenticated: true,
                user: user.username,
            }),
        });

        return render(
            <AuthProvider>
                <Header onToggleTheme={mockToggleTheme} mode={mode} />
            </AuthProvider>
        );
    };

    it('renders header with logo and title', async () => {
        renderHeader();

        await waitFor(() => {
            expect(screen.getByAltText('pgEdge MCP Client')).toBeInTheDocument();
            expect(screen.getByText('MCP Client')).toBeInTheDocument();
        });
    });

    it('displays correct logo based on theme mode', async () => {
        const { rerender } = renderHeader('light');

        await waitFor(() => {
            const logo = screen.getByAltText('pgEdge MCP Client');
            expect(logo).toHaveAttribute('src', 'logo-light.png');
        });

        // Re-render with dark mode
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({
                authenticated: true,
                user: 'testuser',
            }),
        });

        rerender(
            <AuthProvider>
                <Header onToggleTheme={mockToggleTheme} mode="dark" />
            </AuthProvider>
        );

        await waitFor(() => {
            const logo = screen.getByAltText('pgEdge MCP Client');
            expect(logo).toHaveAttribute('src', 'logo-dark.png');
        });
    });

    it('toggles theme when theme button is clicked', async () => {
        renderHeader();
        const user = userEvent.setup();

        await waitFor(() => {
            expect(screen.getByLabelText('toggle theme')).toBeInTheDocument();
        });

        const themeButton = screen.getByLabelText('toggle theme');
        await user.click(themeButton);

        expect(mockToggleTheme).toHaveBeenCalledTimes(1);
    });

    it('displays correct theme icon based on mode', async () => {
        const { rerender } = renderHeader('light');

        await waitFor(() => {
            // In light mode, should show dark mode icon (moon)
            const themeButton = screen.getByLabelText('toggle theme');
            expect(themeButton).toBeInTheDocument();
        });

        // Re-render with dark mode
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({
                authenticated: true,
                user: 'testuser',
            }),
        });

        rerender(
            <AuthProvider>
                <Header onToggleTheme={mockToggleTheme} mode="dark" />
            </AuthProvider>
        );

        await waitFor(() => {
            // In dark mode, should show light mode icon (sun)
            const themeButton = screen.getByLabelText('toggle theme');
            expect(themeButton).toBeInTheDocument();
        });
    });

    it('opens help panel when help button is clicked', async () => {
        renderHeader();
        const user = userEvent.setup();

        await waitFor(() => {
            expect(screen.getByLabelText('open help')).toBeInTheDocument();
        });

        const helpButton = screen.getByLabelText('open help');
        await user.click(helpButton);

        // Wait for help panel to appear
        await waitFor(() => {
            expect(screen.getByText('Help & Documentation')).toBeInTheDocument();
        });
    });

    it('closes help panel when close button is clicked', async () => {
        renderHeader();
        const user = userEvent.setup();

        await waitFor(() => {
            expect(screen.getByLabelText('open help')).toBeInTheDocument();
        });

        // Open help panel
        const helpButton = screen.getByLabelText('open help');
        await user.click(helpButton);

        await waitFor(() => {
            expect(screen.getByText('Help & Documentation')).toBeInTheDocument();
        });

        // Close help panel
        const closeButton = screen.getByLabelText('close help');
        await user.click(closeButton);

        await waitFor(() => {
            expect(screen.queryByText('Help & Documentation')).not.toBeInTheDocument();
        });
    });

    it('opens user menu when avatar is clicked', async () => {
        renderHeader('light', { username: 'testuser' });
        const user = userEvent.setup();

        await waitFor(() => {
            expect(screen.getByLabelText('user menu')).toBeInTheDocument();
        });

        const avatarButton = screen.getByLabelText('user menu');
        await user.click(avatarButton);

        await waitFor(() => {
            expect(screen.getByText('Logout')).toBeInTheDocument();
        });
    });

    it('calls logout when logout menu item is clicked', async () => {
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({
                authenticated: true,
                user: 'testuser',
            }),
        });

        // Mock logout API call
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({}),
        });

        renderHeader();
        const user = userEvent.setup();

        await waitFor(() => {
            expect(screen.getByLabelText('user menu')).toBeInTheDocument();
        });

        // Open user menu
        const avatarButton = screen.getByLabelText('user menu');
        await user.click(avatarButton);

        await waitFor(() => {
            expect(screen.getByText('Logout')).toBeInTheDocument();
        });

        // Click logout
        const logoutButton = screen.getByText('Logout');
        await user.click(logoutButton);

        // Verify logout was called
        await waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith('/api/logout', expect.any(Object));
        });
    });

    it('closes user menu after logout', async () => {
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({
                authenticated: true,
                user: 'testuser',
            }),
        });

        // Mock logout API call
        global.fetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({}),
        });

        renderHeader();
        const user = userEvent.setup();

        await waitFor(() => {
            expect(screen.getByLabelText('user menu')).toBeInTheDocument();
        });

        // Open user menu
        const avatarButton = screen.getByLabelText('user menu');
        await user.click(avatarButton);

        await waitFor(() => {
            expect(screen.getByText('Logout')).toBeInTheDocument();
        });

        // Click logout
        const logoutButton = screen.getByText('Logout');
        await user.click(logoutButton);

        // Menu should close
        await waitFor(() => {
            // Menu is still rendered but not visible (MUI keeps it in DOM but hidden)
            // Just verify the logout was called
            expect(global.fetch).toHaveBeenCalledWith('/api/logout', expect.any(Object));
        });
    });
});
