// Copyright 2017 aaharu All rights reserved.
// This source code is licensed under the BSD-style license found in
// the LICENSE file in the root directory of this source tree.

package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"sort"
	"strconv"
	"strings"

	"github.com/aaharu/schemarshal/utils"
	"github.com/aaharu/schemarshal/version"
)

// Generator of Go source code from JSON Schema
type Generator struct {
	name     string // package nage
	command  string
	imports  ImportSpec
	decls    []*typeSpec
	enumList EnumSpec
}

// NewGenerator create Generator struct
func NewGenerator(packageName string, command string) *Generator {
	return &Generator{
		name:     packageName,
		command:  command,
		imports:  ImportSpec{},
		decls:    []*typeSpec{},
		enumList: EnumSpec{},
	}
}

// addType add a type statement
func (g *Generator) addType(name string, jsonType *JSONType) {
	g.decls = append(g.decls, &typeSpec{
		name:     name,
		jsontype: jsonType,
	})
}

// AddSchema add JSONSchema to Generator
func (g *Generator) AddSchema(name string, js *JSONSchema) error {
	genType, err := js.parse(name, g)
	if err != nil {
		return err
	}
	g.addType(name, genType)
	return nil
}

// Generate gofmt-ed Go source code
func (g *Generator) Generate() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("// Code generated by %s `%s`\n", version.String(), g.command))
	buf.WriteString("// DO NOT RECOMMEND EDITING THIS FILE.\n\n")
	buf.WriteString(fmt.Sprintf("package %s\n\n", g.name))

	if len(g.imports) > 1 {
		buf.WriteString("import (\n")
		// sort map
		var keys []string
		for k := range g.imports {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, path := range keys {
			name := g.imports[path]
			buf.WriteString(fmt.Sprintf("%s %s\n", name, path))
		}
		buf.WriteString(")\n\n")
	} else if len(g.imports) == 1 {
		for path, name := range g.imports {
			buf.WriteString(fmt.Sprintf("import %s %s\n\n", name, path))
		}
	}

	if g.decls != nil {
		for i := range g.decls {
			buf.WriteString("type " + g.decls[i].name + " ")
			g.decls[i].jsontype.nullable = false
			buf.Write(g.decls[i].jsontype.generate())
			buf.WriteString("\n")
		}
	}

	if g.enumList != nil && len(g.enumList) > 0 {
		buf.WriteString("\n")
		// sort map
		var keys []string
		for k := range g.enumList {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, typeName := range keys {
			enum := g.enumList[typeName]
			buf.WriteString("type " + typeName + " int\n")
			buf.WriteString("const (\n")
			for i := range enum {
				if i == 0 {
					buf.WriteString(utils.UpperCamelCase(typeName+" "+fmt.Sprintf("%v", enum[0])) + " " + typeName + " = iota\n")
				} else {
					buf.WriteString(utils.UpperCamelCase(typeName+" "+fmt.Sprintf("%v", enum[i])) + "\n")
				}
			}
			buf.WriteString(")\n\n")

			var enumMapName = "_" + strings.ToLower(typeName[:1])
			if len(typeName) > 1 {
				enumMapName += typeName[1:]
			}
			buf.WriteString("var " + enumMapName + " = map[" + typeName + "]interface{}{\n")
			for i := range enum {
				buf.WriteString(utils.UpperCamelCase(typeName+" "+fmt.Sprintf("%v", enum[i])) + ": ")
				switch v := enum[i].(type) {
				case string:
					buf.WriteString(strconv.Quote(v))
				default:
					buf.WriteString(fmt.Sprintf("%v", v))
				}
				buf.WriteString(",\n")
			}
			buf.WriteString("}\n\n")

			buf.WriteString("func (enum " + typeName + ") MarshalJSON() ([]byte, error) {\n")
			buf.WriteString("switch v:= " + enumMapName + "[enum].(type) {\n")
			buf.WriteString("case string:\n")
			buf.WriteString("return []byte(strconv.Quote(v)), nil\n")
			buf.WriteString("default:\n")
			buf.WriteString("return []byte(fmt.Sprintf(\"%v\", v)), nil\n")
			buf.WriteString("}\n")
			buf.WriteString("}\n\n")

			buf.WriteString("func (enum *" + typeName + ") UnmarshalJSON(data []byte) error {\n")
			buf.WriteString("for i, v := range " + enumMapName + " {\n")
			buf.WriteString("switch vv := v.(type) {\n")
			buf.WriteString("case string:\n")
			buf.WriteString("if strconv.Quote(vv) == string(data) {\n")
			buf.WriteString("*enum = " + typeName + "(i)\n")
			buf.WriteString("return nil\n")
			buf.WriteString("}\n")
			buf.WriteString("default:\n")
			buf.WriteString("if fmt.Sprintf(\"%v\", v) == string(data) {\n")
			buf.WriteString("*enum = " + typeName + "(i)\n")
			buf.WriteString("return nil\n")
			buf.WriteString("}\n")
			buf.WriteString("}\n")
			buf.WriteString("}\n")
			buf.WriteString("return fmt.Errorf(\"Error: miss-matched " + typeName + " (%s)\", data)\n")
			buf.WriteString("}\n\n")

			buf.WriteString("func (enum " + typeName + ") String() string {\n")
			buf.WriteString("switch v:= " + enumMapName + "[enum].(type) {\n")
			buf.WriteString("case string:\n")
			buf.WriteString("return v\n")
			buf.WriteString("default:\n")
			buf.WriteString("return fmt.Sprintf(\"%v\", v)\n")
			buf.WriteString("}\n")
			buf.WriteString("}\n\n")

			buf.WriteString("func To" + typeName + "(val interface{}) (" + typeName + ", error) {\n")
			buf.WriteString("for i, v := range " + enumMapName + " {\n")
			buf.WriteString("if val == v {")
			buf.WriteString("return i, nil")
			buf.WriteString("}\n")
			buf.WriteString("}\n")
			buf.WriteString("return 0, fmt.Errorf(\"Error: Failed to " + typeName + ": %v\", val)")
			buf.WriteString("}\n\n")
		}
	}

	return format.Source(buf.Bytes())
}

// ImportSpec has `import` information
type ImportSpec map[string]string

type typeSpec struct {
	name     string // type name
	jsontype *JSONType
}

// EnumSpec has enum information
type EnumSpec map[string][]interface{}

type jsonFormat int

const (
	formatObject jsonFormat = iota
	formatArray
	formatString
	formatBoolean
	formatNumber
	formatInteger
	formatDatetime
)

// JSONType is type of json
type JSONType struct {
	format   jsonFormat
	nullable bool
	fields   []*field  // object has
	itemType *JSONType // array has
	typeName string    // object's array and object has
	enumType string    // enum has
}

func (t *JSONType) addField(f *field) {
	if t.fields == nil {
		t.fields = []*field{}
	}
	t.fields = append(t.fields, f)
}

func (t *JSONType) generate() []byte {
	var buf bytes.Buffer
	if t.nullable {
		buf.WriteString("*")
	}
	if t.enumType != "" {
		buf.WriteString(t.enumType)
	} else if t.format == formatObject {
		if t.fields == nil {
			buf.WriteString("map[string]interface{}")
		} else {
			if t.typeName != "" {
				buf.WriteString(t.typeName)
			} else {
				buf.WriteString("struct {\n")
				for i := range t.fields {
					buf.WriteString(t.fields[i].name)
					buf.WriteString(" ")
					buf.Write(t.fields[i].jsontype.generate())
					buf.WriteString(" ")
					buf.Write(t.fields[i].jsontag.generate())
					buf.WriteString("\n")
				}
				buf.WriteString("}")
			}
		}
	} else if t.format == formatArray {
		buf.WriteString("[]")
		if t.typeName != "" {
			buf.WriteString(t.typeName)
		} else {
			buf.Write(t.itemType.generate())
		}
	} else if t.format == formatString {
		buf.WriteString("string")
	} else if t.format == formatBoolean {
		buf.WriteString("bool")
	} else if t.format == formatNumber {
		buf.WriteString("float64")
	} else if t.format == formatInteger {
		buf.WriteString("int")
	} else if t.format == formatDatetime {
		buf.WriteString("time.Time")
	}
	return buf.Bytes()
}

type field struct {
	name     string
	jsontype *JSONType
	jsontag  *jsonTag
}

type jsonTag struct {
	name      string
	omitEmpty bool
}

// Generate JSON tag code
func (t *jsonTag) generate() []byte {
	var buf bytes.Buffer
	buf.WriteString("`json:\"")
	buf.WriteString(t.name)
	if t.omitEmpty {
		buf.WriteString(",omitempty")
	}
	buf.WriteString("\"`")
	return buf.Bytes()
}
