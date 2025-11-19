# pgEdge MCP Web UI - Test Suite Summary

## Test Coverage Overview

This document summarizes the comprehensive test suite created for the pgEdge MCP Web UI.

### Components Tested

1. **Header Component** (`src/components/__tests__/Header.test.jsx`)
   - Theme toggling (light/dark mode)
   - Logo display based on theme
   - Help panel opening/closing
   - User avatar and initials display
   - User menu interactions
   - Logout functionality

2. **HelpPanel Component** (`src/components/__tests__/HelpPanel.test.jsx`)
   - Panel open/close states
   - All documentation sections displayed
   - Close button functionality
   - Drawer visibility states
   - Content for all help topics

3. **StatusBanner Component** (`src/components/__tests__/StatusBanner.test.jsx`)
   - System info fetching on mount
   - Connected/disconnected status display
   - PostgreSQL version and connection string display
   - Error handling and display
   - Expand/collapse functionality
   - Detailed system information display
   - Periodic refresh (every 30 seconds)
   - Authentication failure handling
   - Network error handling

4. **ChatInterface Component** (`src/components/__tests__/ChatInterface.test.jsx`)
   - Initial empty state
   - Message input and sending
   - Keyboard shortcuts (Enter to send, Shift+Enter for newline)
   - Provider and model dropdowns
   - Show Activity toggle
   - Render Markdown toggle
   - localStorage persistence for preferences
   - Clear conversation functionality
   - Query history navigation with arrow keys
   - SSE streaming message handling
   - Activity display (tools and resources)
   - Error handling
   - Loading states

5. **App Component** (`src/__tests__/App.test.jsx`)
   - Initial loading state
   - Login screen display when unauthenticated
   - Main app display when authenticated
   - Theme management (light/dark)
   - Theme persistence to localStorage
   - Theme restoration from localStorage
   - Header rendering
   - Auth context provision
   - Auth check failure handling
   - Viewport constraints

6. **AuthContext** (`src/contexts/__tests__/AuthContext.test.jsx`)
   - Initial unauthenticated state
   - Authentication status check on mount
   - Login functionality
   - Login failure handling
   - Logout functionality

7. **Login Component** (`src/components/__tests__/Login.test.jsx`)
   - Form rendering
   - Input validation
   - Submit button disable state during loading
   - Error message display on login failure
   - Navigation on successful login
   - Error clearing on input change

8. **useMenu Hook** (`src/hooks/__tests__/useMenu.test.js`)
   - Initial state
   - Menu open functionality
   - Menu close functionality
   - Multiple open/close cycles
   - AnchorEl management
   - Open state calculation

## Test Framework & Tools

- **Test Runner**: Vitest
- **Testing Library**: @testing-library/react
- **User Interactions**: @testing-library/user-event
- **Assertions**: @testing-library/jest-dom
- **DOM Environment**: happy-dom
- **API Mocking**: MSW (Mock Service Worker) - available but not yet fully utilized

## Test Features

### Unit Tests
- Individual component behavior
- Props handling
- State management
- Event handlers
- Rendering logic

### Integration Tests
- Component interactions
- Context providers
- Custom hooks
- API communication
- LocalStorage persistence
- SessionStorage usage

### User Interaction Tests
- Click events
- Keyboard events
- Form submissions
- Menu interactions
- Drawer opening/closing

## Coverage Goals

The test suite aims for comprehensive coverage of:
- **Components**: All UI components
- **Hooks**: Custom React hooks
- **Contexts**: Auth context and providers
- **User Flows**: Login, chat, theme switching
- **Error States**: API failures, network errors, validation errors
- **Edge Cases**: Empty states, loading states, error states

## Running Tests

```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage report
npm run test:coverage

# Run tests with UI
npm run test:ui
```

## Known Issues / Test Improvements Needed

1. **StatusBanner Tests**: Some timeout issues with fake timers - need to ensure proper cleanup and timer advancement
2. **AuthContext Tests**: Need to ensure fetch mocks are properly set up before component renders
3. **SSE Streaming Tests**: Complex mocking of ReadableStream - may need additional utilities
4. **Material-UI Select Tests**: Dropdown selection testing is limited due to MUI's complex implementation

## Test Maintenance

- Tests use explicit waits (`waitFor`) for async operations
- Fake timers are used for time-based features (intervals, timeouts)
- localStorage and sessionStorage are cleared between tests
- Fetch API is mocked globally in each test suite
- Components are properly unmounted after each test

## Future Improvements

1. Add E2E tests using Playwright or Cypress
2. Add visual regression testing
3. Improve MSW integration for more realistic API mocking
4. Add performance benchmarks
5. Add accessibility (a11y) tests
6. Increase coverage of edge cases
7. Add integration tests for complete user workflows
