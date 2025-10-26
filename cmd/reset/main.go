package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

const resetComment = "// generate:reset"

const src = `/* Тестовый пакет */
package main

// generate:reset
type ResetableStruct struct {
	i     int
	str   string
	strP  *string
	s     []int
	m     map[string]string
	child *ResetableStruct
}

// non-resetable struct
type NonResetableStruct struct {
	i     int
	str   string
	strP  *string
	s     []int
	m     map[string]string
	child *ResetableStruct
}
	
func main() {
	// create a resetable struct
	rs := &ResetableStruct{
		i:     1,
		str:   "test",
		strP:  &str,
		s:     []int{1, 2, 3},
		m:     map[string]string{"test": "test"},
		child: &ResetableStruct{i: 2, str: "test2", strP: &str2, s: []int{4, 5, 6}, m: map[string]string{"test2": "test2"}, child: nil},
	}
}
`

func main() {

	// parse the file
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, ``, src, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// find all the lines of declaration of the structures
	structIdxs := make(map[int]*ast.TypeSpec)
	ast.Inspect(f, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		// if the node is not a type spec -> return
		if !ok {
			return true
		}
		// if the type is not a struct -> return
		if _, ok := ts.Type.(*ast.StructType); !ok {
			return true
		}
		// add the line index to the map
		line := fset.Position(ts.Pos()).Line
		structIdxs[line] = ts
		return true
	})

	// find the reset comment and check if the next line is a struct
	for _, gr := range f.Comments {
		// if the comment list is empty -> continue
		if len(gr.List) == 0 {
			continue
		}
		// get the last comment in the group
		last := gr.List[len(gr.List)-1]
		// if the comment is not the reset comment -> continue
		if last.Text != resetComment {
			continue
		}
		// get the line index of the comment
		cl := fset.Position(last.Slash).Line
		// get the struct index from the map by the line index
		if ts, ok := structIdxs[cl+1]; ok {
			fmt.Printf("FOUND struct at line %d: %s\n", cl+1, ts.Name.Name)
			// iterate over the fields and generate
			fields := ts.Type.(*ast.StructType).Fields.List
			for _, field := range fields {
				fmt.Printf("Field: %s\n", field.Names[0].Name)
				
				switch field.Type.(type) {

			}
		}
	}

}
