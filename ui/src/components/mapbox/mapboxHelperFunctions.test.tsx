// IMPORTANT: All mocks must be defined before imports
vi.mock('react-dom/client', () => ({
    createRoot: vi.fn(() => ({
        render: vi.fn(),
        unmount: vi.fn(),
    })),
}));

vi.mock('../../Pages/PlacesPage/MapClusterMarker', () => ({
    default: vi.fn(() => null),
}));

// Now import modules
import React from 'react'
import { vi, describe, test, expect, beforeEach, afterEach } from 'vitest'
import { registerMediaMarkers } from './mapboxHelperFunctions'
import type { PlacesAction } from '../../Pages/PlacesPage/placesReducer'

// Define needed types
type MarkerElement = HTMLDivElement & {
    _root?: {
        render: () => void;
        unmount: () => void;
    };
}

describe('mapboxHelperFunctions', () => {
    let mockDispatch: React.Dispatch<PlacesAction>;
    let consoleWarnSpy: ReturnType<typeof vi.spyOn>;
    let consoleErrorSpy: ReturnType<typeof vi.spyOn>;
    let mockMap: Record<string, any>;
    let mockFeatures: any[];
    let eventHandlers: Record<string, (...args: any[]) => void>;

    beforeEach(() => {
        mockDispatch = vi.fn() as any;
        consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => { });
        consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => { });
        mockFeatures = [];
        eventHandlers = {};

        // Create mock map
        mockMap = {
            on: vi.fn((event, handler) => {
                eventHandlers[event] = handler;
                return mockMap;
            }),
            querySourceFeatures: vi.fn(() => mockFeatures),
        };
    });

    afterEach(() => {
        vi.restoreAllMocks();
        consoleWarnSpy.mockRestore();
        consoleErrorSpy.mockRestore();
    });

    describe('registerMediaMarkers', () => {
        test('registers event handlers on the map', () => {
            // Create simplified mock that doesn't actually use document.createElement
            const mockMapboxgl = {
                Marker: vi.fn(() => ({
                    setLngLat: vi.fn().mockReturnThis(),
                    addTo: vi.fn().mockReturnThis(),
                    remove: vi.fn(),
                    getElement: vi.fn(() => ({
                        // No need to be a real DOM element for this test
                        _root: { unmount: vi.fn() }
                    })),
                })),
            };

            registerMediaMarkers({
                map: mockMap as any,
                mapboxLibrary: mockMapboxgl as any,
                dispatchMarkerMedia: mockDispatch,
            });

            expect(mockMap.on).toHaveBeenCalledTimes(3);
            expect(mockMap.on).toHaveBeenCalledWith('move', expect.any(Function));
            expect(mockMap.on).toHaveBeenCalledWith('moveend', expect.any(Function));
            expect(mockMap.on).toHaveBeenCalledWith('sourcedata', expect.any(Function));
        });

        test('creates markers for valid features', () => {
            // Add a feature
            const feature = {
                geometry: {
                    type: 'Point',
                    coordinates: [0, 0],
                },
                properties: {
                    cluster: false,
                    media_id: 'test-id',
                    thumbnail: JSON.stringify({ url: 'test-url' }),
                },
            };
            mockFeatures.push(feature);

            // Create mock marker
            const mockMarker = {
                setLngLat: vi.fn().mockReturnThis(),
                addTo: vi.fn().mockReturnThis(),
                remove: vi.fn(),
                getElement: vi.fn(() => ({
                    _root: { unmount: vi.fn() }
                })),
            };

            // Create mapboxgl mock
            const mockMapboxgl = {
                Marker: vi.fn(() => mockMarker),
            };

            // FIXED: Use a mocked object directly instead of calling createElement
            vi.spyOn(document, 'createElement').mockImplementation((tagName: string) => {
                if (tagName === 'div') {
                    // Create a mock object matching the shape we need
                    const mockDiv = {
                        tagName: 'DIV',
                        style: {},
                        classList: { add: vi.fn(), remove: vi.fn() }
                    } as unknown as MarkerElement;

                    // Now we can safely add _root property
                    mockDiv._root = { render: vi.fn(), unmount: vi.fn() };
                    return mockDiv;
                }
                // For other elements, return a basic mock
                return {} as any;
            });

            registerMediaMarkers({
                map: mockMap as any,
                mapboxLibrary: mockMapboxgl as any,
                dispatchMarkerMedia: mockDispatch,
            });

            expect(mockMapboxgl.Marker).toHaveBeenCalledTimes(1);
            expect(mockMarker.setLngLat).toHaveBeenCalledWith([0, 0]);
            expect(mockMarker.addTo).toHaveBeenCalledWith(mockMap);
        });

        test('handles features with missing geometry', () => {
            // Add feature without geometry
            const feature = {
                properties: {
                    cluster: false,
                    media_id: 'test-id',
                    thumbnail: JSON.stringify({ url: 'test-url' }),
                },
            };
            mockFeatures.push(feature);

            const mockMapboxgl = {
                Marker: vi.fn(),
            };

            registerMediaMarkers({
                map: mockMap as any,
                mapboxLibrary: mockMapboxgl as any,
                dispatchMarkerMedia: mockDispatch,
            });

            expect(mockMapboxgl.Marker).not.toHaveBeenCalled();
            expect(consoleWarnSpy).toHaveBeenCalledWith(
                'WARN: geojson feature had no geometry',
                { feature }
            );
        });

        test('handles non-Point geometry type', () => {
            // Add feature with LineString geometry
            const feature = {
                geometry: {
                    type: 'LineString',
                    coordinates: [[0, 0], [1, 1]],
                },
                properties: {
                    cluster: false,
                    media_id: 'test-id',
                    thumbnail: JSON.stringify({ url: 'test-url' }),
                },
            };
            mockFeatures.push(feature);

            const mockMapboxgl = {
                Marker: vi.fn(),
            };

            registerMediaMarkers({
                map: mockMap as any,
                mapboxLibrary: mockMapboxgl as any,
                dispatchMarkerMedia: mockDispatch,
            });

            expect(mockMapboxgl.Marker).not.toHaveBeenCalled();
            expect(consoleWarnSpy).toHaveBeenCalledWith(
                'WARN: geojson feature geometry is not a Point',
                { feature }
            );
        });

        test('handles features with missing properties', () => {
            // Add feature without properties
            const feature = {
                geometry: {
                    type: 'Point',
                    coordinates: [0, 0],
                },
                properties: null,
            };
            mockFeatures.push(feature);

            const mockMapboxgl = {
                Marker: vi.fn(),
            };

            registerMediaMarkers({
                map: mockMap as any,
                mapboxLibrary: mockMapboxgl as any,
                dispatchMarkerMedia: mockDispatch,
            });

            expect(mockMapboxgl.Marker).not.toHaveBeenCalled();
            expect(consoleWarnSpy).toHaveBeenCalledWith(
                'WARN: geojson feature had no properties',
                expect.objectContaining({
                    feature
                })
            );
        });
    });
});
