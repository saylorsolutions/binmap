package main

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	//go:embed file.got
	genTemplText string
	genTempl     = template.Must(template.New("bingen").Funcs(map[string]any{
		"id": identifier,
		"hasStruct": func(field fieldMapping) bool {
			return len(field.Struct.TypeName) > 0
		},
		"excludedField": func(field fieldMapping) bool {
			switch MapType(field.BinType) {
			case StructPad:
				return true
			case Magic:
				return true
			default:
				return false
			}
		},
	}).Parse(genTemplText))
)

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		log.Fatalln("Not enough args")
	}
	procFilePath, outDir := args[0], args[1]
	if err := run(procFilePath, outDir); err != nil {
		log.Fatalln(err)
	}
	log.Println("Generated ")
}

func run(procFilePath, outDir string) error {
	in, err := os.Open(procFilePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Close()
	}()
	s, err := parseMapping(in)
	if err != nil {
		return err
	}

	generatedFile := strings.ToLower(identifier(filepath.Base(procFilePath))) + ".go"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	out, err := os.Create(filepath.Join(outDir, generatedFile))
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()
	s.Package = filepath.Base(outDir)
	if err := genTempl.Execute(out, s); err != nil {
		return err
	}
	return nil
}
