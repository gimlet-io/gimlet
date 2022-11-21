package dx

// BranchDeletedEvent contains all metadata about the deleted branch
type BranchDeletedEvent struct {
	Manifests []*Manifest
	Branch    string
	Repo      string
}
