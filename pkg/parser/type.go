package parser

import (
	"unicode"

	"k8s.io/gengo/types"
)

type Type struct {
	*types.Type
}

func (t *Type) realType() *Type {
	if t.Type.Kind == "Pointer" || t.Type.Kind == "Slice" {
		return &Type{t.Type.Elem}
	}
	return t
}

func (t *Type) Package() string {
	return t.realType().Type.Name.Package
}

func (t *Type) Name() string {
	return t.realType().Type.Name.Name
}

func (t *Type) LowerName() string {
	return lowerName(t.Name())
}

func (t *Type) key() objectKey {
	return objectKey{
		Package: t.Package(),
		Name:    t.Name(),
	}
}

func lowerName(s string) string {
	name := []rune(s)
	for pos, r := range name {
		if !unicode.IsUpper(r) {
			break
		}
		if pos > 0 && len(name) > pos+1 && unicode.IsLower(name[pos+1]) {
			continue
		}
		name[pos] = unicode.ToLower(r)
	}
	return string(name)
}
