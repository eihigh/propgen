package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"unicode/utf8"

	gen "github.com/moznion/gowrtr/generator"
)

func isPrivate(field string) bool {
	runes := []rune(field)
	initial := runes[0]
	if 'a' <= initial && initial <= 'z' {
		return true
	}
	return false
}

func receiverName(typeName string) string {
	runes := []rune(typeName)
	initial := runes[0]
	return strings.ToLower(string(initial))
}

type field struct {
	name, typ string
}

var (
	targets = map[string][]field{} // typeName => fields
)

func addField(typeName, name, typ string) {
	if isPrivate(name) {
		if targets[typeName] == nil {
			targets[typeName] = []field{}
		}
		targets[typeName] = append(targets[typeName], field{name, typ})
	}
}

func astNodeString(src string, node ast.Node) string {
	return src[node.Pos()-1 : node.End()-1]
}

func main() {
	src := `
package entity

type Body struct {
	pos, vec   complex128
	angle, dir float64
	t time.Time
}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		panic(err)
	}
	// ast.Print(fset, f)
	pkg := f.Name.Name

	// get target type and fields
	ast.Inspect(f, func(n ast.Node) bool {

		switch x := n.(type) {
		case *ast.TypeSpec:
			if st, ok := x.Type.(*ast.StructType); ok {

				structName := x.Name.Name
				for _, field := range st.Fields.List {

					typ := astNodeString(src, field.Type)
					for _, nameObj := range field.Names {

						name := nameObj.Name
						addField(structName, name, typ)
					}
				}
			}

		}
		return true
	})

	for typeName, t := range targets {
		for _, field := range t {
			fmt.Println(typeName, field.name, field.typ)
		}
	}

	g := gen.NewRoot(
		gen.NewComment(" auto generatod by propgen."),
		gen.NewPackage(pkg),
	)

	for typeName, t := range targets {
		for _, field := range t {

			r, _ := utf8.DecodeRuneInString(typeName)
			receiver := strings.ToLower(string(r))
			getter := field.name
			i, size := utf8.DecodeRuneInString(getter)
			getter = strings.ToUpper(string(i)) + getter[size:]
			setter := "Set" + getter
			accessor := receiver + "." + field.name
			ptrTypeName := "*" + typeName

			g = g.AddStatements(
				gen.NewNewline(),
				gen.NewFunc(
					gen.NewFuncReceiver(receiver, ptrTypeName),
					gen.NewFuncSignature(getter).
						AddReturnTypes(field.typ),
				).
					AddStatements(
						gen.NewReturnStatement(accessor),
					),

				gen.NewNewline(),
				gen.NewFunc(
					gen.NewFuncReceiver(receiver, ptrTypeName),
					gen.NewFuncSignature(setter).
						AddFuncParameters(
							gen.NewFuncParameter(field.name, field.typ),
						),
				).
					AddStatements(
						gen.NewRawStatement(accessor+" = "+field.name),
					),
			)
		}
	}

	result, err := g.Generate(0)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
