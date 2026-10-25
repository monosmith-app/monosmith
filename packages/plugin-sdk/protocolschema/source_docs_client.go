package protocolschema

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ConstValue is a string or integer constant read from the protocol source.
type ConstValue struct {
	Name  string
	Value string // the literal's value; an integer is kept as its decimal text
	Doc   string
}

// SourceDocs is what reflection cannot see: the doc comments of the protocol
// package's types and fields, and its constants with a literal value.
type SourceDocs struct {
	types  map[string]string
	fields map[string]map[string]string
	enums  map[string][]ConstValue // typed constants by type name, in declaration order
	consts []ConstValue            // every constant with a literal value, in file order
	groups map[string]string       // the doc of the const block that declares a type's values
}

// ReadSourceDocs parses the non-test Go files of the protocol package in dir.
func ReadSourceDocs(dir string) (*SourceDocs, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	d := &SourceDocs{
		types: map[string]string{}, fields: map[string]map[string]string{},
		enums: map[string][]ConstValue{}, groups: map[string]string{},
	}
	fset := token.NewFileSet()
	for _, e := range entries { // os.ReadDir sorts by name: the order is stable
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, e.Name()), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		for _, decl := range file.Decls {
			if gen, ok := decl.(*ast.GenDecl); ok {
				if err := d.readDecl(gen); err != nil {
					return nil, fmt.Errorf("%s: %w", e.Name(), err)
				}
			}
		}
	}
	return d, nil
}

func (d *SourceDocs) readDecl(gen *ast.GenDecl) error {
	switch gen.Tok {
	case token.TYPE:
		for _, spec := range gen.Specs {
			ts := spec.(*ast.TypeSpec)
			doc := ts.Doc
			if doc == nil && len(gen.Specs) == 1 {
				doc = gen.Doc
			}
			d.types[ts.Name.Name] = Paragraphs(doc.Text())
			if st, ok := ts.Type.(*ast.StructType); ok {
				d.readFields(ts.Name.Name, st)
			}
		}
	case token.CONST:
		for _, spec := range gen.Specs {
			vs := spec.(*ast.ValueSpec)
			if len(vs.Values) != 1 || len(vs.Names) != 1 {
				continue
			}
			value, ok, err := literal(vs.Values[0])
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			doc := vs.Doc
			if doc == nil {
				doc = vs.Comment
			}
			c := ConstValue{Name: vs.Names[0].Name, Value: value, Doc: WithoutSubject(vs.Names[0].Name, Paragraphs(doc.Text()))}
			d.consts = append(d.consts, c)
			if id, typed := vs.Type.(*ast.Ident); typed {
				d.enums[id.Name] = append(d.enums[id.Name], c)
				if gen.Doc != nil {
					d.groups[id.Name] = Paragraphs(gen.Doc.Text())
				}
			}
		}
	}
	return nil
}

func (d *SourceDocs) readFields(typeName string, st *ast.StructType) {
	docs := map[string]string{}
	for _, f := range st.Fields.List {
		doc := f.Doc
		if doc == nil {
			doc = f.Comment
		}
		for _, name := range f.Names {
			docs[name.Name] = Paragraphs(doc.Text())
		}
	}
	d.fields[typeName] = docs
}

// literal evaluates a string or integer literal, with an optional minus sign.
func literal(e ast.Expr) (string, bool, error) {
	switch x := e.(type) {
	case *ast.BasicLit:
		switch x.Kind {
		case token.STRING:
			s, err := strconv.Unquote(x.Value)
			return s, err == nil, err
		case token.INT:
			return x.Value, true, nil
		}
	case *ast.UnaryExpr:
		if x.Op == token.SUB {
			inner, ok, err := literal(x.X)
			if ok && err == nil {
				return "-" + inner, true, nil
			}
			return "", false, err
		}
	}
	return "", false, nil
}

// TypeDoc is the doc comment of a type, as paragraphs.
func (d *SourceDocs) TypeDoc(name string) string { return d.types[name] }

// FieldDoc is the doc comment of a struct field, by its Go name.
func (d *SourceDocs) FieldDoc(typeName, field string) string { return d.fields[typeName][field] }

// Enum lists the typed constants declared for a type, in order.
func (d *SourceDocs) Enum(typeName string) []ConstValue { return d.enums[typeName] }

// EnumGroupDoc is the doc of the const block that declares a type's values.
func (d *SourceDocs) EnumGroupDoc(typeName string) string { return d.groups[typeName] }

// ConstsWithPrefix lists the constants whose name starts with prefix and
// continues with an upper-case letter, sorted by name.
func (d *SourceDocs) ConstsWithPrefix(prefix string) []ConstValue {
	var out []ConstValue
	for _, c := range d.consts {
		rest, found := strings.CutPrefix(c.Name, prefix)
		if found && rest != "" && rest[0] >= 'A' && rest[0] <= 'Z' {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Const finds a constant by name.
func (d *SourceDocs) Const(name string) (ConstValue, bool) {
	for _, c := range d.consts {
		if c.Name == name {
			return c, true
		}
	}
	return ConstValue{}, false
}
