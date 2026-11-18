import React, { createContext, useState } from 'react'

interface AlbumTreeSearchContextType {
  query: string
  setQuery: (query: string) => void
}

export const AlbumTreeSearchContext = createContext<AlbumTreeSearchContextType>(
  {
    query: '',
    setQuery: () => {
      console.warn(
        'AlbumTreeSearchContext: setQuery was called before initialized'
      )
    },
  }
)
AlbumTreeSearchContext.displayName = 'AlbumTreeSearchContext'

type AlbumTreeSearchProviderProps = {
  children: React.ReactNode
}

export const AlbumTreeSearchProvider = ({
  children,
}: AlbumTreeSearchProviderProps) => {
  const [query, setQuery] = useState('')

  return (
    <AlbumTreeSearchContext.Provider value={{ query, setQuery }}>
      {children}
    </AlbumTreeSearchContext.Provider>
  )
}
