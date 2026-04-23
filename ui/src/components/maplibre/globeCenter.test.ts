import { describe, it, expect } from 'vitest'
import {
  findBestView,
  extractCoordsFromGeoJson,
  type GlobeView,
} from './globeCenter'

function expectView(actual: GlobeView, expected: GlobeView) {
  expect(actual.center.lng).toBeCloseTo(expected.center.lng, 1)
  expect(actual.center.lat).toBeCloseTo(expected.center.lat, 1)
  expect(actual.zoom).toBeCloseTo(expected.zoom, 1)
}

describe('findBestView', () => {
  it('returns default view for empty input', () => {
    expectView(findBestView([]), { center: { lng: 65, lat: 20 }, zoom: 2 })
  })

  it('returns close-up view for a single marker', () => {
    expectView(findBestView([{ lng: 10, lat: 50 }]), {
      center: { lng: 10, lat: 50 },
      zoom: 12,
    })
  })

  it('centers on majority cluster when outlier is at north pole', () => {
    expectView(
      findBestView([
        { lng: 0, lat: 0 },
        { lng: 0, lat: 0 },
        { lng: 0, lat: 0 },
        { lng: 0, lat: 90 },
      ]),
      { center: { lng: 0, lat: 0 }, zoom: 2 }
    )
  })

  it('centers on majority cluster when outlier is at 90 degrees east', () => {
    expectView(
      findBestView([
        { lng: 0, lat: 0 },
        { lng: 0, lat: 0 },
        { lng: 0, lat: 0 },
        { lng: 90, lat: 0 },
      ]),
      { center: { lng: 0, lat: 0 }, zoom: 2 }
    )
  })

  it('centers on Europe for continental cluster', () => {
    expectView(
      findBestView([
        { lng: 2, lat: 48 }, // Paris
        { lng: 13, lat: 52 }, // Berlin
        { lng: -3, lat: 40 }, // Madrid
        { lng: 12, lat: 41 }, // Rome
        { lng: 23, lat: 37 }, // Athens
      ]),
      { center: { lng: 9.56, lat: 43.98 }, zoom: 4.87 }
    )
  })

  it('centers on Europe and zooms out for distant outlier', () => {
    expectView(
      findBestView([
        { lng: 2, lat: 48 }, // Paris
        { lng: 13, lat: 52 }, // Berlin
        { lng: -3, lat: 40 }, // Madrid
        { lng: 12, lat: 41 }, // Rome
        { lng: 23, lat: 37 }, // Athens
        { lng: -157, lat: 21 }, // Hawaii (outlier)
      ]),
      { center: { lng: 9.56, lat: 43.98 }, zoom: 1.66 }
    )
  })

  it('centers on western US and zooms out for distant outlier', () => {
    expectView(
      findBestView([
        { lng: -74, lat: 40 }, // NYC
        { lng: -87, lat: 41 }, // Chicago
        { lng: -118, lat: 34 }, // LA
        { lng: -122, lat: 37 }, // SF
        { lng: 139, lat: 35 }, // Tokyo (outlier)
      ]),
      { center: { lng: -116.55, lat: 48.96 }, zoom: 2.31 }
    )
  })

  it('centers on the correct height on the atlantic for transatlantic markers, in the event of a tie', () => {
    expectView(
      findBestView([
        { lng: 2, lat: 48 }, // Paris
        { lng: 13, lat: 52 }, // Berlin
        { lng: -74, lat: 40 }, // NYC
        { lng: -87, lat: 41 }, // Chicago
      ]),
      { center: { lng: -41.16, lat: 54.51 }, zoom: 3.45 }
    )
  })

  it('picks perpendicular center for antipodal markers', () => {
    expectView(
      findBestView([
        { lng: 0, lat: 0 },
        { lng: 180, lat: 0 },
      ]),
      { center: { lng: 0, lat: 0 }, zoom: 1 }
    )
  })

  it('centers near antimeridian for markers straddling it', () => {
    expectView(
      findBestView([
        { lng: 170, lat: 35 },
        { lng: 175, lat: 40 },
        { lng: -175, lat: 38 },
        { lng: -170, lat: 36 },
      ]),
      { center: { lng: -180, lat: 37.52 }, zoom: 5.41 }
    )
  })

  it('centers on Europe for three-city triangle', () => {
    expectView(
      findBestView([
        { lng: 2, lat: 48 }, // Paris
        { lng: 13, lat: 52 }, // Berlin
        { lng: 12, lat: 41 }, // Rome
      ]),
      { center: { lng: 9.03, lat: 47.11 }, zoom: 5.8 }
    )
  })

  it('centers near north pole for equidistant triangle at low latitude', () => {
    expectView(
      findBestView([
        { lng: 0, lat: 10 },
        { lng: 120, lat: 10 },
        { lng: -120, lat: 10 },
      ]),
      { center: { lng: 0, lat: 90 }, zoom: 2.17 }
    )
  })

  it('centers near north pole for antipodal line at low latitude', () => {
    expectView(
      findBestView([
        { lng: 0, lat: 10 },
        { lng: 180, lat: 10 },
      ]),
      { center: { lng: 90, lat: 90 }, zoom: 2.17 }
    )
  })

  it('zooms way out for global spread', () => {
    expectView(
      findBestView([
        { lng: -74, lat: 40 }, // NYC
        { lng: 139, lat: 35 }, // Tokyo
        { lng: 2, lat: 48 }, // Paris
      ]),
      { center: { lng: -33.87, lat: 80.86 }, zoom: 2.49 }
    )
  })
})

describe('extractCoordsFromGeoJson', () => {
  it('extracts coords from Point features', () => {
    const geojson: GeoJSON.FeatureCollection = {
      type: 'FeatureCollection',
      features: [
        {
          type: 'Feature',
          geometry: { type: 'Point', coordinates: [10, 50] },
          properties: {},
        },
        {
          type: 'Feature',
          geometry: { type: 'Point', coordinates: [-74, 40] },
          properties: {},
        },
      ],
    }

    expect(extractCoordsFromGeoJson(geojson)).toEqual([
      { lng: 10, lat: 50 },
      { lng: -74, lat: 40 },
    ])
  })

  it('returns empty array for no features', () => {
    const geojson: GeoJSON.FeatureCollection = {
      type: 'FeatureCollection',
      features: [],
    }
    expect(extractCoordsFromGeoJson(geojson)).toEqual([])
  })

  it('skips non-Point geometries', () => {
    const geojson: GeoJSON.FeatureCollection = {
      type: 'FeatureCollection',
      features: [
        {
          type: 'Feature',
          geometry: {
            type: 'LineString',
            coordinates: [
              [0, 0],
              [1, 1],
            ],
          },
          properties: {},
        },
        {
          type: 'Feature',
          geometry: {
            type: 'Polygon',
            coordinates: [
              [
                [0, 0],
                [1, 0],
                [1, 1],
                [0, 0],
              ],
            ],
          },
          properties: {},
        },
      ],
    }
    expect(extractCoordsFromGeoJson(geojson)).toEqual([])
  })

  it('extracts only Points from a mixed-geometry collection', () => {
    const geojson: GeoJSON.FeatureCollection = {
      type: 'FeatureCollection',
      features: [
        {
          type: 'Feature',
          geometry: { type: 'Point', coordinates: [10, 50] },
          properties: {},
        },
        {
          type: 'Feature',
          geometry: {
            type: 'LineString',
            coordinates: [
              [0, 0],
              [1, 1],
            ],
          },
          properties: {},
        },
        {
          type: 'Feature',
          geometry: { type: 'Point', coordinates: [-74, 40] },
          properties: {},
        },
      ],
    }
    expect(extractCoordsFromGeoJson(geojson)).toEqual([
      { lng: 10, lat: 50 },
      { lng: -74, lat: 40 },
    ])
  })
})
