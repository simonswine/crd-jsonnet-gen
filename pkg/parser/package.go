package parser

import (
	"k8s.io/gengo/types"
)

const (
	packageDocCommentForceIncludes = "// +gencrdrefdocs:force"
)

type Package struct {
	*types.Package
}

// APIGropu extracts the "//+groupName" meta-comment from the specified
// package's godoc, or returns empty string if it cannot be found.
func (pkg *Package) APIGroup() string {
	m := types.ExtractCommentTags("+", pkg.DocComments)
	v := m["groupName"]
	if len(v) == 1 {
		return v[0]
	}
	return ""
}
