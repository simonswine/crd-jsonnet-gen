package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"unicode"

	"github.com/google/go-jsonnet/ast"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/astext"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/printer"
	"k8s.io/gengo/parser"
	"k8s.io/gengo/types"
)

func newIdentifier(value string) *ast.Identifier {
	id := ast.Identifier(value)
	return &id
}

func newLiteralNumber(in string) *ast.LiteralNumber {
	f, err := strconv.ParseFloat(in, 64)
	if err != nil {
		return &ast.LiteralNumber{OriginalString: in, Value: 0}
	}
	return &ast.LiteralNumber{OriginalString: in, Value: f}
}

// groupName extracts the "//+groupName" meta-comment from the specified
// package's godoc, or returns empty string if it cannot be found.
func groupName(pkg *types.Package) string {
	m := types.ExtractCommentTags("+", pkg.DocComments)
	v := m["groupName"]
	if len(v) == 1 {
		return v[0]
	}
	return ""
}

const (
	docCommentForceIncludes = "// +gencrdrefdocs:force"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatalf("failed: %v", err)
	}
}

func containsString(sl []string, str string) bool {
	for _, s := range sl {
		if str == s {
			return true
		}
	}
	return false
}

func toCamelCase(s string) string {
	runes := []rune(s)
	if len(runes) < 1 {
		return s
	}
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func typeToResourceObj(t *types.Type) (ast.ObjectField, error) {
	kind := t.Name.Name
	kindCamel := toCamelCase(kind)
	obj, err := astext.CreateField(kindCamel)
	if err != nil {
		return ast.ObjectField{}, fmt.Errorf("unable to create resource identifier: %w", err)
	}
	obj.Expr2 = &ast.Object{
		Fields: ast.ObjectFields{
			ast.ObjectField{
				Kind: ast.ObjectLocal,
				Id:   newIdentifier("kind"),
				Expr2: &ast.Object{
					Fields: ast.ObjectFields{
						ast.ObjectField{
							Kind: ast.ObjectFieldID,
							Hide: ast.ObjectFieldInherit,
							Id:   newIdentifier("kind"),
							Expr2: &ast.LiteralString{
								Kind:  ast.StringDouble,
								Value: kind,
							},
						},
					},
				},
			},
			ast.ObjectField{
				Kind: ast.ObjectFieldID,
				Id:   newIdentifier("new"),
				Expr2: &ast.Binary{
					Left:  &ast.Var{Id: *newIdentifier("kind")},
					Right: &ast.Var{Id: *newIdentifier("apiVersion")},
					Op:    ast.BopPlus,
				},
				Method: &ast.Function{},
			},
		},
	}
	return obj.ObjectField, nil
}

func run(paths []string) error {

	if len(paths) == 0 {
		return errors.New("no package paths given as arguments")
	}

	log.Printf("search for CRDs in paths: %v", paths)

	p := parser.New()

	for _, path := range paths {
		if err := p.AddDirRecursive(path); err != nil {
			return err
		}
	}

	scan, err := p.FindTypes()
	if err != nil {
		return err
	}

	var pkgNames []string
	for p := range scan {
		pkg := scan[p]
		if groupName(pkg) != "" && len(pkg.Types) > 0 || containsString(pkg.DocComments, docCommentForceIncludes) {
			log.Printf("package=%v has groupName and has types", p)
			pkgNames = append(pkgNames, p)
		}
	}

	var obj = &ast.Object{
		Fields: ast.ObjectFields{},
	}

	for _, p := range pkgNames {
		pkg := scan[p]

		groupNameStr := groupName(pkg)

		var resourcesObj ast.ObjectFields
		for _, t := range pkg.Types {
			tags := types.ExtractCommentTags("+", t.SecondClosestCommentLines)
			if _, ok := tags["genclient"]; !ok {
				continue
			}
			log.Printf("found object groupName=%s name=%+v kind=%+v", groupNameStr, t.Name, t.Kind)
			if obj, err := typeToResourceObj(t); err != nil {
				return fmt.Errorf("unable to build resource object for %s, %w", t.Name, err)
			} else {
				resourcesObj = append(resourcesObj, obj)
			}
		}

		apiVersionObj, err := astext.CreateField(pkg.Name)
		if err != nil {
			return fmt.Errorf("unable to create apiVersion identifier: %w", err)
		}
		apiVersionObj.Expr2 = &ast.Object{
			Fields: append(ast.ObjectFields{
				ast.ObjectField{
					Kind: ast.ObjectLocal,
					Id:   newIdentifier("apiVersion"),
					Expr2: &ast.Object{
						Fields: ast.ObjectFields{
							ast.ObjectField{
								Kind: ast.ObjectFieldID,
								Hide: ast.ObjectFieldInherit,
								Id:   newIdentifier("apiVersion"),
								Expr2: &ast.Binary{
									Left: &ast.LiteralString{
										Kind:  ast.StringDouble,
										Value: fmt.Sprintf("%%s/%s", pkg.Name),
									},
									Right: &ast.Var{Id: *newIdentifier("apiGroup")},
									Op:    ast.BopPercent,
								},
							},
						},
					},
				},
			}, resourcesObj...),
		}

		apiGroupObj, err := astext.CreateField(groupNameStr)
		if err != nil {
			return fmt.Errorf("unable to create apiGroup identifier: %w", err)
		}
		apiGroupObj.Expr2 = &ast.Object{
			Fields: ast.ObjectFields{
				{
					Kind: ast.ObjectLocal,
					Id:   newIdentifier("apiGroup"),
					Expr2: &ast.LiteralString{
						Kind:  ast.StringDouble,
						Value: groupNameStr,
					},
				},
				apiVersionObj.ObjectField,
			},
		}
		obj.Fields = append(obj.Fields, apiGroupObj.ObjectField)
	}

	if err := printer.Fprint(os.Stdout, obj); err != nil {
		return err
	}

	return nil
}
