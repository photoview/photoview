import { vi } from 'vitest'
import { render, fireEvent, waitFor } from '@testing-library/react'
import MergeFaceGroupsModal, { MergeFaceGroupsModalState } from './MergeFaceGroupsModal'
import { MockedProvider } from '@apollo/client/testing'
import { MY_FACES_QUERY } from '../PeoplePage'
import { COMBINE_FACES_MUTATION } from './MergeFaceGroupsModal'

// Mock Modal component to prevent import issues
vi.mock('../../../primitives/Modal', () => ({
    __esModule: true,
    default: ({ children, open, title, description, actions, onClose }: any) =>
        open ? (
            <div data-testid="modal">
                <div data-testid="modal-title">{title}</div>
                <div data-testid="modal-description">{description}</div>
                <div data-testid="modal-content">{children}</div>
                <div data-testid="modal-actions">
                    {actions?.map((action: any) => (
                        <button key={action.key} onClick={action.onClick}>
                            {action.label}
                        </button>
                    ))}
                </div>
            </div>
        ) : null,
}))

// Mock useTranslation to prevent i18n issues
vi.mock('react-i18next', () => ({
    useTranslation: () => ({
        t: (key: string, fallback?: string) => fallback || key,
    }),
}))

// Mock useNavigate before any imports that use it
const navigate = vi.fn()
vi.mock('react-router-dom', async () => {
    const actual: object = await vi.importActual('react-router-dom')
    return { ...actual, useNavigate: () => navigate }
})

// Mock IntersectionObserver for tests
beforeAll(() => {
    global.IntersectionObserver = class {
        constructor() { }
        observe() { }
        unobserve() { }
        disconnect() { }
    } as any
})

// Helper function to convert face group ID to test ID
function idToTestId(id: string): string {
    return `facegroup-${id}`
}

// Mock SelectFaceGroupTable to simplify selection logic
vi.mock('./SelectFaceGroupTable', () => ({
    __esModule: true,
    default: ({ faceGroups, selectedFaceGroups, toggleSelectedFaceGroup, title }: any) => (
        <div>
            <div>{title}</div>
            {faceGroups.map((fg: any) => (
                <button
                    key={fg.id}
                    data-testid={idToTestId(fg.id)}
                    onClick={() => toggleSelectedFaceGroup(fg)}
                    style={{ fontWeight: selectedFaceGroups.has(fg) ? 'bold' : 'normal' }}
                >
                    {fg.label}
                </button>
            ))}
        </div>
    ),
}))

const mockFaceGroups = ["Alice", "Bob", "Charlie", "David", "Felix"].map((name, index) => { return { __typename: 'FaceGroup', id: index.toString(), label: name, imageFaceCount: 0, imageFaces: [] } })

const myFacesMock = {
    request: { query: MY_FACES_QUERY },
    result: { data: { myFaceGroups: mockFaceGroups } },
}

function getCombineFacesMock(destinationID: string, sourceIDs: string[]) {
    return {
        request: {
            query: COMBINE_FACES_MUTATION,
            variables: {
                destID: destinationID,
                srcIDs: sourceIDs,
            },
        },
        result: {
            data: {
                combineFaceGroups: {
                    id: destinationID,
                    __typename: 'FaceGroup',
                },
            },
        },
    }
}

// Tests the merging of source face groups into a destination face group
// Logic extracted to be used in several tests
async function testMerge(destinationID: string, sourceIDs: string[]) {
    const setState = vi.fn()

    const destinationTestID: string = idToTestId(destinationID)
    const sourceTestIDs: string[] = sourceIDs.map(idToTestId)

    const combineFacesMock = getCombineFacesMock(destinationID, sourceIDs)

    // Render modal in SelectDestination state
    const { getByText, getByTestId, getByRole, rerender } = render(
        <MockedProvider mocks={[myFacesMock, combineFacesMock]} addTypename={false}>
            <MergeFaceGroupsModal
                state={MergeFaceGroupsModalState.SelectDestination}
                setState={setState}
                refetchQueries={[]}
            />
        </MockedProvider>
    )

    // Wait for face groups to load and select destination
    await waitFor(() => getByTestId(destinationTestID))
    fireEvent.click(getByTestId(destinationTestID))

    // Click "Next" to go to SelectSources
    fireEvent.click(getByText(/Next/i))
    expect(setState).toHaveBeenCalledWith(MergeFaceGroupsModalState.SelectSources)

    // Rerender in SelectSources state
    rerender(
        <MockedProvider mocks={[myFacesMock, combineFacesMock]} addTypename={false}>
            <MergeFaceGroupsModal
                state={MergeFaceGroupsModalState.SelectSources}
                setState={setState}
                refetchQueries={[]}
            />
        </MockedProvider>
    )

    // Wait for source face groups to load
    await waitFor(() => getByTestId(sourceTestIDs[0]))

    // Select multiple source face groups
    for (const testID of sourceTestIDs)
        fireEvent.click(getByTestId(testID))

    // Click the Merge button
    fireEvent.click(getByRole('button', { name: /Merge/i }))

    // Check that the modal closes and redirects to the destination face group
    await waitFor(() => {
        expect(setState).toHaveBeenCalledWith(MergeFaceGroupsModalState.Closed)
        expect(navigate).toHaveBeenCalledWith(`/people/${destinationID}`)
    })
}

test('merges a single source face group into a destination face group', async () => {
    testMerge(mockFaceGroups[0].id, [mockFaceGroups[1].id])
})

test('merges all source face groups into a destination face group', async () => {
    testMerge(mockFaceGroups[0].id, mockFaceGroups.slice(1).map(fg => fg.id))
})

test('merges multiple source face groups into a destination face group', async () => {
    testMerge(mockFaceGroups[0].id, [mockFaceGroups[1], mockFaceGroups[2]].map(fg => fg.id))
})
