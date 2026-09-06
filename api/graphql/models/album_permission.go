package models

// rank gives AlbumPermissionLevel (a gqlgen-generated string enum, see
// graphql/models/generated.go) a total order: Delete implies Upload implies
// Read. Plain string comparison on the enum's wire values ("DELETE" <
// "READ" < "UPLOAD" alphabetically) would not match this, hence this
// explicit mapping.
func (l AlbumPermissionLevel) rank() int {
	switch l {
	case AlbumPermissionLevelRead:
		return 0
	case AlbumPermissionLevelUpload:
		return 1
	case AlbumPermissionLevelDelete:
		return 2
	default:
		return -1
	}
}

// AtLeast reports whether l is the same as, or a higher tier than, other.
func (l AlbumPermissionLevel) AtLeast(other AlbumPermissionLevel) bool {
	return l.rank() >= other.rank()
}

// HigherThan reports whether l is a strictly higher tier than other.
func (l AlbumPermissionLevel) HigherThan(other AlbumPermissionLevel) bool {
	return l.rank() > other.rank()
}
