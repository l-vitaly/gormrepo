// Copyright 2017 The Vitaly Lobchuk. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// For example, given this struct:
//
// package model
//
// type User struct {
//    gorm.Model
//    FirstName string
//    LastName string
//    Birthday time.Time
// }
//
// running this command:
//
// gormrepogen -t=User
//
// Typically this process would be run using go generate, like this:
//
//      //go:generate gormrepogen -t=User
//
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	typeNames = flag.String("t", "", "comma-separated list of type names; must be set")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\tgormrepogen [flags] -t T [directory]\n")
	fmt.Fprintf(os.Stderr, "\tgormrepogen [flags] -t T files... # Must be a single package\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func init() {
	log.SetFlags(0)
	log.SetPrefix("gormrepogen: ")
	flag.Usage = Usage
	flag.Parse()
}

func main() {
	if len(*typeNames) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	types := strings.Split(*typeNames, ",")

	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	var g Generator

	if len(args) == 1 && isDirectory(args[0]) {
		g.parsePackageDir(args[0])
	} else {
		g.parsePackageFiles(args)
	}

	// Run generate for each type.
	for _, typeName := range types {
		g.generate(typeName)
	}
}

func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

// prefixDirectory places the directory name on the beginning of each name in the list.
func prefixDirectory(directory string, names []string) []string {
	if directory == "." {
		return names
	}
	ret := make([]string, len(names))
	for i, name := range names {
		ret[i] = filepath.Join(directory, name)
	}
	return ret
}

type File struct {
	name string    // Name of the constant type.
	file *ast.File // Parsed AST.
}

type Generator struct {
	buf   bytes.Buffer // Accumulated output.
	files []*File
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

func (g *Generator) parsePackageDir(directory string) {
	pkg, err := build.Default.ImportDir(directory, 0)
	if err != nil {
		log.Fatalf("cannot process directory %s: %s", directory, err)
	}
	var names []string
	names = append(names, pkg.GoFiles...)
	names = append(names, pkg.CgoFiles...)
	names = append(names, pkg.SFiles...)
	names = prefixDirectory(directory, names)

	g.parsePackage(directory, names, nil)
}

func (g *Generator) parsePackageFiles(names []string) {
	g.parsePackage(".", names, nil)
}

func (g *Generator) parsePackage(directory string, names []string, text interface{}) {
	var files []*File
	fs := token.NewFileSet()
	for _, name := range names {
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		parsedFile, err := parser.ParseFile(fs, name, text, 0)
		if err != nil {
			log.Fatalf("parsing package: %s: %s", name, err)
		}

		files = append(files, &File{
			file: parsedFile,
			name: name,
		})
	}
	if len(files) == 0 {
		log.Fatalf("%s: no buildable Go files", directory)
	}

	g.files = files
}

func (g *Generator) format() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.buf.Bytes()
	}
	return src
}

func (g *Generator) ucFirst(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func (g *Generator) lcFirst(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

func (g *Generator) getFileByTypeName(typeName string) *File {
	for _, f := range g.files {
		for _, d := range f.file.Decls {
			if gd, ok := d.(*ast.GenDecl); ok {
				for _, s := range gd.Specs {
					if tp, ok := s.(*ast.TypeSpec); ok {
						if tp.Name.Name == typeName {
							return f
						}
					}
				}
			}
		}
	}
	return nil
}

// generate repository for the named type.
func (g *Generator) generate(typeName string) {
	f := g.getFileByTypeName(typeName)
	if f != nil {
		repoName := g.lcFirst(typeName) + "BaseRepo"
		repoNameRecv := "*" + repoName
		typeNameWithPointer := "*" + typeName

		// Print the header and package clause.
		g.Printf("// Code generated by \"gormrepogen %s\"; DO NOT EDIT\n", strings.Join(os.Args[1:], " "))
		g.Printf("\n")
		g.Printf("package %s", f.file.Name.Name)
		g.Printf("\n")
		g.Printf("import (\n")
		g.Printf("  \"github.com/l-vitaly/gormrepo\"\n")
		g.Printf(")\n")

		g.Printf(baseRepo, repoName)
		g.Printf(repoApplyCriteria, repoNameRecv)
        g.Printf(repoRelated, repoNameRecv, typeNameWithPointer)
		g.Printf(repoGet, repoNameRecv, typeNameWithPointer, typeName)
		g.Printf(repoGetAll, repoNameRecv, typeNameWithPointer)
		g.Printf(repoGetBy, repoNameRecv, typeNameWithPointer)
		g.Printf(repoGetByFirst, repoNameRecv, typeNameWithPointer, typeName)
		g.Printf(repoGetByLast, repoNameRecv, typeNameWithPointer, typeName)
		g.Printf(repoCreate, repoNameRecv, typeName, typeNameWithPointer)
		g.Printf(repoUpdate, repoNameRecv, typeNameWithPointer)
        g.Printf(repoDelete, repoNameRecv, typeNameWithPointer)
		g.Printf(repoAutomigrate, repoNameRecv, typeName)
        g.Printf(repoAddUniqueIndex, repoNameRecv, typeName)
        g.Printf(repoAddForeignKey, repoNameRecv, typeName)
        g.Printf(repoAddIndex, repoNameRecv, typeName)

		//Format the output.
		src := g.format()

		absPath, _ := filepath.Abs(f.name)
		dir := filepath.Dir(absPath)

		baseName := fmt.Sprintf("%s_base_repo.go", typeName)
		outputName := filepath.Join(dir, strings.ToLower(baseName))

		err := ioutil.WriteFile(outputName, src, 0644)
		if err != nil {
			log.Fatalf("writing output: %s", err)
		}

		fmt.Printf("Type %s repository is generated: %s\n", typeName, outputName)

		g.buf.Reset()
	} else {
		fmt.Printf("Type %s is not found\n", typeName)
	}
}

const baseRepo = `
type %[1]s struct {
    *gorm.DB
}`

const repoApplyCriteria = `
func (r %[1]s) applyCriteria(criteria []gormrepo.CriteriaOption) *gorm.DB {
	search := r.DB
	for _, co := range criteria {
		search = co(search)
	}
	return search
}
`

const repoRelated = `
func (r %[1]s) Related(claim %[2]s, related interface{}, criteria ...gormrepo.CriteriaOption) error {
	return r.applyCriteria(criteria).Model(claim).Related(related).Error
}
`

const repoGet = `
func (r %[1]s) Get(id uint) (%[2]s, error) {
    var entity %[3]s
	err := r.DB.Where(map[string]interface{}{"id": id}).Find(&entity).Error
	return &entity, err
}`

const repoGetAll = `
func (r %[1]s) GetAll() ([]%[2]s, error) {
    return r.GetBy()
}
`

const repoGetBy = `
func (r %[1]s) GetBy(criteria ...gormrepo.CriteriaOption) ([]%[2]s, error) {
    var entities []%[2]s
	err := r.applyCriteria(criteria).Find(&entities).Error
	return entities, err
}
`

const repoGetByFirst = `
func (r %[1]s) GetByFirst(criteria ...gormrepo.CriteriaOption) (%[2]s, error) {
    var entity %[3]s
	err := r.applyCriteria(criteria).First(&entity).Error
	return &entity, err
}
`

const repoGetByLast = `
func (r %[1]s) GetByLast(criteria ...gormrepo.CriteriaOption) (%[2]s, error) {
    var entity %[3]s
	err := r.applyCriteria(criteria).Last(&entity).Error
	return &entity, err
}
`

const repoCreate = `
func (r %[1]s) Create(entity %[2]s) (%[3]s, error) {
    if !r.DB.NewRecord(entity) {
		return nil, gormrepo.ErrPrimaryNotBlank
	}
	err := r.DB.Create(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}
`

const repoUpdate = `
func (r %[1]s) Update(entity %[2]s, fields gormrepo.Fields, criteria ...gormrepo.CriteriaOption) error {
    return r.applyCriteria(criteria).Model(entity).Updates(fields).Error
}
`

const repoDelete = `
func (r %[1]s) Delete(entity %[2]s, criteria ...gormrepo.CriteriaOption) error {
	return r.applyCriteria(criteria).Delete(entity).Error
}
`

const repoAutomigrate = `
func (r %[1]s) AutoMigrate() error {
    return r.DB.AutoMigrate(&%[2]s{}).Error
}
`

const repoAddUniqueIndex = `
func (r %[1]s) AddUniqueIndex(name string, columns ...string) error {
    return r.DB.Model(&%[2]s{}).AddUniqueIndex(name, columns...).Error
}
`
const repoAddForeignKey = `
func (r %[1]s) AddForeignKey(field string, dest string, onDelete string, onUpdate string) error {
    return r.DB.Model(&%[2]s{}).AddForeignKey(field, dest, onDelete, onUpdate).Error
}
`
const repoAddIndex = `
func (r %[1]s) AddIndex(name string, columns ...string) error {
    return r.DB.Model(&%[2]s{}).AddIndex(name, columns...).Error
}
`
