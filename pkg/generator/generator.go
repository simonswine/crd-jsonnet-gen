package generator

import (
	"os"
	"regexp"
	"sort"

	"github.com/fatih/structtag"
	"github.com/go-logr/logr"
	"github.com/google/go-jsonnet/ast"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/astext"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/printer"

	"github.com/simonswine/crd-jsonnet-gen/pkg/parser"
)

var log logr.Logger

func WithLogger(l logr.Logger) {
	log = l
}

type DataSource interface {
	PackageToAPIGroup(string) []string
	Types() []*parser.Type
}

func createField(s string) (interface{}, error) {
	return astext.CreateField(s)
}

// reFieldId matches a field id that doesn't need enclosing in quotes
var reFieldId = regexp.MustCompile(`^[_A-Za-z]+[_A-Za-z0-9]*$`)

func setIdentifier(f *ast.ObjectField, id string) {
	// handle fields which valid
	if reFieldId.MatchString(id) {
		idObj := ast.Identifier(id)
		f.Id = &idObj
		return
	}
	f.Expr1 = &ast.LiteralString{
		Value: id,
		Kind:  ast.StringSingle,
	}
}

func Generate(s DataSource) error {
	types := s.Types()
	sort.Slice(types, func(i, j int) bool { return types[i].Name() < types[j].Name() })

	objectsPerPackage := make(map[string]ast.ObjectFields)

	for pos := range types {
		t := types[pos]
		pkg := t.Package()
		if _, ok := objectsPerPackage[pkg]; !ok {
			objectsPerPackage[pkg] = []ast.ObjectField{}
		}

		obj := typeToJsonnetObject(t)

		field := ast.ObjectField{
			Kind:  ast.ObjectFieldStr,
			Expr2: obj,
		}
		setIdentifier(&field, t.LowerName())

		objectsPerPackage[pkg] = append(
			objectsPerPackage[pkg],
			field,
		)
	}

	// ensure packages and their objects are sorted
	var packageNames []string
	for pkg := range objectsPerPackage {
		packageNames = append(packageNames, pkg)
	}
	sort.Strings(packageNames)

	var obj = &ast.Object{
		Fields: ast.ObjectFields{},
	}

	for _, pkg := range packageNames {
		objs := objectsPerPackage[pkg]
		field := ast.ObjectField{
			Kind: ast.ObjectFieldStr,
			Expr2: &ast.Object{
				Fields: objs,
			},
		}
		setIdentifier(&field, pkg)

		obj.Fields = append(
			obj.Fields,
			field,
		)
	}

	if err := printer.Fprint(os.Stdout, obj); err != nil {
		return err
	}

	return nil
}

func typeToJsonnetObject(t *parser.Type) *ast.Object {
	obj := &ast.Object{}

	for _, m := range t.Members {
		if m.Type.IsPrimitive() && m.Type.Name.Name == "string" {
			tags, err := structtag.Parse(string(m.Tags))
			if err != nil {
				panic(err)
			}

			jsonTag, err := tags.Get("json")
			if err != nil {
				panic(err)
			}

			field := ast.ObjectField{
				Kind:  ast.ObjectFieldStr,
				Expr2: &ast.Object{},
			}
			setIdentifier(&field, "with"+m.Name)
			obj.Fields = append(obj.Fields, field)

			log.V(3).Info("my fields", "name", m.Name, "member", m.Type.Name.String(), "tags", jsonTag)
		}
	}

	return obj
}
