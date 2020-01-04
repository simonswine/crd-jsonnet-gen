package parser

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/gengo/parser"
	"k8s.io/gengo/types"
)

type Parser struct {
	*parser.Builder
	log logr.Logger

	apiGroup map[string]string
}

type objectKey struct {
	Package string
	Name    string
}

func containsString(sl []string, str string) bool {
	for _, s := range sl {
		if str == s {
			return true
		}
	}
	return false
}

func New() *Parser {
	return &Parser{
		Builder:  parser.New(),
		apiGroup: make(map[string]string),
	}
}

func (p *Parser) WithLogger(l logr.Logger) *Parser {
	p.log = l
	return p
}

func walkType(objects map[objectKey]*Type, t *Type) {
	for _, m := range t.Members {
		key := objectKey{
			Package: m.Type.Name.Package,
			Name:    m.Type.Name.Name,
		}
		if _, ok := objects[key]; ok {
			continue
		}
		mType := &Type{m.Type}
		objects[key] = mType
		walkType(objects, mType)
	}
}

func reverseStringSlice(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

func (p *Parser) PackageToAPIGroup(pkgName string) []string {
	apiGroup, ok := p.apiGroup[pkgName]
	if !ok {
		return []string{}
	}
	parts := strings.Split(apiGroup, ".")
	reverseStringSlice(parts)
	return parts
}

func (p *Parser) ExportResources() ([]*Type, error) {

	objectMap := make(map[objectKey]*Type)

	scan, err := p.FindTypes()
	if err != nil {
		return nil, fmt.Errorf("error during finding types: %w", err)
	}

	var pkgNames []string
	for key := range scan {
		pkg := &Package{scan[key]}
		if pkg.APIGroup() != "" && len(pkg.Types) > 0 || containsString(pkg.DocComments, packageDocCommentForceIncludes) {
			p.log.V(3).Info("package has groupName and has types", "package", key)
			pkgNames = append(pkgNames, key)
			p.apiGroup[key] = pkg.APIGroup()
		}
	}

	for _, key := range pkgNames {
		pkg := &Package{scan[key]}
		for _, t := range pkg.Types {
			t := &Type{
				Type: t,
			}
			tags := types.ExtractCommentTags("+", t.SecondClosestCommentLines)
			if _, ok := tags["genclient"]; !ok {
				continue
			}
			p.log.V(3).Info("found object", "groupName", pkg.APIGroup(), "name", t.Type.Name.Name, "kind", t.Kind)
			objectMap[t.key()] = t
		}
	}

	for _, t := range objectMap {
		walkType(objectMap, t)
	}

	var objectList []*Type

	for _, t := range objectMap {
		if t.Package() != "" {
			objectList = append(objectList, t)
		}
	}

	return objectList, nil
}
