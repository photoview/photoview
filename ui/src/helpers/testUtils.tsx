import { MockedProvider, MockedResponse } from '@apollo/client/testing'
import { DefaultOptions, FetchPolicy, WatchQueryFetchPolicy } from '@apollo/client/core'
import { render } from '@testing-library/react'
import { MessageProvider } from '../components/messages/MessageState'
import { MemoryRouter, Route, Routes, Location } from 'react-router-dom'
import React from 'react'

interface RenderWithProvidersOptions {
    mocks?: MockedResponse[]
    initialEntries?: (string | Partial<Location>)[]
    route?: React.ReactElement
    path?: string
    apolloOptions?: {
        defaultOptions?: {
            watchQuery?: { fetchPolicy?: WatchQueryFetchPolicy }
            query?: { fetchPolicy?: FetchPolicy }
        }
    }
}

export function renderWithProviders(
    ui: React.ReactNode,
    {
        mocks = [],
        initialEntries = ['/'],
        route,
        path,
        apolloOptions = {},
    }: RenderWithProvidersOptions = {}
) {
    return render(
        <MockedProvider
            mocks={mocks}
            addTypename={false}
            defaultOptions={apolloOptions.defaultOptions}
        >
            <MemoryRouter initialEntries={initialEntries}>
                {route && path ? (
                    <Routes>
                        <Route path={path} element={route} />
                    </Routes>
                ) : (
                    <MessageProvider>{ui}</MessageProvider>
                )}
            </MemoryRouter>
        </MockedProvider>
    )
}
