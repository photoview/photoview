import { MockedProvider } from '@apollo/client/testing'
import { render } from '@testing-library/react'
import { MessageProvider } from '../components/messages/MessageState'
import { MemoryRouter, Route, Routes, Location, Router } from 'react-router-dom'
import React from 'react'
import { History } from 'history'

/**
 * Options for configuring the test environment in renderWithProviders.
 */
interface RenderWithProvidersOptions {
    /** Apollo GraphQL mocks */
    mocks?: any[]
    /** Initial router entries for MemoryRouter */
    initialEntries?: (string | Partial<Location>)[]
    /** Route element to render */
    route?: React.ReactElement
    /** Path for the route */
    path?: string
    /** History object for HistoryRouter */
    history?: History
    /** Apollo client configuration options */
    apolloOptions?: {
        addTypename?: boolean
        defaultOptions?: any
    }
}

/**
 * Renders a component with common providers needed for testing.
 * @param ui - The React component to render
 * @param options - Configuration options for the test environment
 * @param options.mocks - Apollo GraphQL mocks
 * @param options.initialEntries - Initial router entries
 * @param options.route - Route element to render
 * @param options.path - Path for the route
 * @param options.apolloOptions - Apollo client configuration
 * @returns The rendered component with testing utilities
 */
export function renderWithProviders(
    ui: React.ReactNode,
    {
        mocks = [],
        initialEntries = ['/'],
        route,
        path,
        history,
        apolloOptions = {},
    }: RenderWithProvidersOptions = {}
) {
    if ((route && !path) || (!route && path)) {
        throw new Error('Both route and path must be provided together');
    }
    const Wrapper = ({ children }: { children: React.ReactNode }) => (
        <MockedProvider
            mocks={mocks}
            addTypename={apolloOptions.addTypename ?? false}
            defaultOptions={apolloOptions.defaultOptions}
        >
            {history ? (
                <Router location={history.location} navigator={history}>
                    <MessageProvider>{children}</MessageProvider>
                </Router>
            ) : (
                <MemoryRouter initialEntries={initialEntries}>
                    <MessageProvider>{children}</MessageProvider>
                </MemoryRouter>
            )}
        </MockedProvider>
    )
    return {
        ...render(
            <Wrapper>
                {route && path ? (
                    <Routes>
                        <Route path={path} element={route} />
                    </Routes>
                ) : (
                    ui
                )}
            </Wrapper>
        ),
        history,
    }
}
