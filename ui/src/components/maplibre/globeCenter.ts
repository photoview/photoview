export type Coord = { lng: number; lat: number }

export type GlobeView = { center: Coord; zoom: number }

/**
 * Find the coordinate that, when used as the globe center,
 * maximizes how many markers fall on the visible hemisphere.
 * Avoids the naive centroid approach which can land in the middle of nowhere
 * when markers are spread across continents.
 */
export function findBestView(coords: Coord[]): GlobeView {
  if (coords.length === 0) return { center: { lng: 65, lat: 20 }, zoom: 2 }
  if (coords.length === 1) return { center: coords[0], zoom: 12 }

  const vecs = coords.map(toVec)

  // Find the input point that sees the most other points on its hemisphere
  let bestScore = 0
  let bestIdx = 0
  for (let i = 0; i < vecs.length; i++) {
    const score = scoreCenter(vecs[i], vecs)
    if (score > bestScore) {
      bestScore = score
      bestIdx = i
    }
  }

  let candidate = vecs[bestIdx]

  // If some points are invisible and they're not a small minority (outliers),
  // try to find a compromise center that sees more points
  if (bestScore < vecs.length) {
    const invisible = vecs.filter(p => dot(candidate, p) <= Number.EPSILON)
    if (invisible.length >= bestScore) {
      for (const inv of invisible) {
        const mid = {
          x: candidate.x + inv.x,
          y: candidate.y + inv.y,
          z: candidate.z + inv.z,
        }
        const len = Math.sqrt(mid.x ** 2 + mid.y ** 2 + mid.z ** 2)

        let comp
        if (len < 1e-6) {
          // Antipodal -> use perpendicular direction
          comp = findPerpendicular(candidate)
        } else {
          comp = { x: mid.x / len, y: mid.y / len, z: mid.z / len }
        }

        const compScore = scoreCenter(comp, vecs)
        if (compScore > bestScore) {
          bestScore = compScore
          candidate = comp
        }
      }
    }
  }

  // Compute centroid of only the points visible from the chosen center,
  // which excludes outliers on the opposite hemisphere
  const visible = vecs.filter(p => dot(candidate, p) > Number.EPSILON)
  const centroid =
    visible.length > 0 ? getSphericalCentroid(visible) : candidate

  // Compute zoom from the center to ALL points (outliers widen the view)
  const center = toCoord(centroid)
  const centerVec = toVec(center)
  const zoom = computeZoom(centerVec, vecs)
  return { center, zoom }
}

type DirectionVector = { x: number; y: number; z: number }

function dot(a: DirectionVector, b: DirectionVector): number {
  return a.x * b.x + a.y * b.y + a.z * b.z
}

function findPerpendicular(v: DirectionVector): DirectionVector {
  const ax = Math.abs(v.x),
    ay = Math.abs(v.y),
    az = Math.abs(v.z)
  // Pick the axis where v has the smallest component
  const ref =
    az <= ax && az <= ay
      ? { x: 0, y: 0, z: 1 }
      : ay <= ax
      ? { x: 0, y: 1, z: 0 }
      : { x: 1, y: 0, z: 0 }
  // ref × v gives a vector perpendicular to v
  const cross = {
    x: ref.y * v.z - ref.z * v.y,
    y: ref.z * v.x - ref.x * v.z,
    z: ref.x * v.y - ref.y * v.x,
  }
  const len = Math.sqrt(cross.x ** 2 + cross.y ** 2 + cross.z ** 2)
  return { x: cross.x / len, y: cross.y / len, z: cross.z / len }
}

function toVec(coord: Coord): DirectionVector {
  const lambda = (coord.lng * Math.PI) / 180
  const phi = (coord.lat * Math.PI) / 180
  return {
    x: Math.cos(phi) * Math.cos(lambda),
    y: Math.cos(phi) * Math.sin(lambda),
    z: Math.sin(phi),
  }
}

function toCoord(vec: DirectionVector): Coord {
  const lng = Math.atan2(vec.y, vec.x)
  const hyp = Math.sqrt(vec.x * vec.x + vec.y * vec.y)
  const lat = Math.atan2(vec.z, hyp)

  return {
    lng: (lng * 180) / Math.PI,
    lat: (lat * 180) / Math.PI,
  }
}

function scoreCenter(
  centerVec: DirectionVector,
  points: DirectionVector[]
): number {
  let count = 0
  for (const p of points) {
    if (dot(centerVec, p) > Number.EPSILON) count++
  }
  return count
}

function computeZoom(
  centerVec: DirectionVector,
  points: DirectionVector[]
): number {
  let maxAngle = 0
  for (const p of points) {
    const d = Math.min(1, Math.max(-1, dot(centerVec, p)))
    const angle = Math.acos(d)
    if (angle > maxAngle) maxAngle = angle
  }

  // maxAngle is in radians: 0 = all same spot, PI = opposite side of globe
  // Map to zoom: ~180° spread -> zoom 1, ~1° spread -> zoom 15
  const degrees = (maxAngle * 180) / Math.PI
  if (degrees < 0.01) return 15
  // log scale: zoom = log2(360 / degrees) clamped to [1, 15]
  const zoom = Math.log2(360 / degrees)
  return Math.max(1, Math.min(15, zoom))
}

function getSphericalCentroid(coords: DirectionVector[]): DirectionVector {
  const avg = { x: 0, y: 0, z: 0 }

  for (const { x, y, z } of coords) {
    avg.x += x
    avg.y += y
    avg.z += z
  }

  avg.x /= coords.length
  avg.y /= coords.length
  avg.z /= coords.length
  return avg
}

export function extractCoordsFromGeoJson(
  geojson: GeoJSON.FeatureCollection
): Coord[] {
  const coords: Coord[] = []
  for (const feature of geojson.features) {
    if (feature.geometry.type === 'Point') {
      const [lng, lat] = feature.geometry.coordinates
      coords.push({ lng, lat })
    }
  }
  return coords
}
