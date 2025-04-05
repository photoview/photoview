import { describe, it, expect } from 'vitest'

// Since paginateCache isn't currently exported, we'll need to recreate it for testing
// Consider exporting it from apolloClient.ts to make it directly testable
type PaginateCacheType = {
    keyArgs: string[]
    merge: (existing: any, incoming: any, context: any) => any
}

const paginateCache = (keyArgs: string[]) => ({
    keyArgs,
    merge(existing: any[], incoming: any[], { args, fieldName }: any) {
        const merged = existing ? existing.slice(0) : []
        if (args?.paginate) {
            const { offset = 0 } = args.paginate as { offset: number }
            for (let i = 0; i < incoming.length; ++i) {
                merged[offset + i] = incoming[i]
            }
        } else {
            throw new Error(`Paginate argument is missing for query: ${fieldName}`)
        }
        return merged
    },
} as PaginateCacheType)

describe('paginateCache', () => {
    it('should merge incoming data with existing data using offset', () => {
        const keyArgs = ['testKey']
        const paginateFn = paginateCache(keyArgs)

        // Simulate existing data
        const existing = ['item1', 'item2']

        // Simulate incoming data
        const incoming = ['item3', 'item4']

        // Simulate Apollo cache merge context with paginate args
        const context = {
            args: { paginate: { offset: 2 } },
            fieldName: 'testField'
        }

        const result = paginateFn.merge(existing, incoming, context)

        // Expect merged array with items at correct positions
        expect(result).toEqual(['item1', 'item2', 'item3', 'item4'])
    })

    it('should handle empty existing data', () => {
        const keyArgs = ['testKey']
        const paginateFn = paginateCache(keyArgs)

        // No existing data
        const existing = null

        // Simulate incoming data
        const incoming = ['item1', 'item2']

        // Simulate Apollo cache merge context with paginate args
        const context = {
            args: { paginate: { offset: 0 } },
            fieldName: 'testField'
        }

        const result = paginateFn.merge(existing, incoming, context)

        // Expect only incoming items
        expect(result).toEqual(['item1', 'item2'])
    })

    it('should throw an error when paginate argument is missing', () => {
        const keyArgs = ['testKey']
        const paginateFn = paginateCache(keyArgs)

        // Simulate existing and incoming data
        const existing = ['item1', 'item2']
        const incoming = ['item3', 'item4']

        // Simulate context without paginate args
        const context = {
            args: {},
            fieldName: 'testField'
        }

        // Expect an error to be thrown
        expect(() => {
            paginateFn.merge(existing, incoming, context)
        }).toThrow('Paginate argument is missing for query: testField')
    })
})
