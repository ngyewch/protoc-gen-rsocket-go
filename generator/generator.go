package generator

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/compiler/protogen"
	"strconv"
	"text/template"
)

type Generator struct {
	options            Options
	templates          *template.Template
	commonGeneratedMap map[string]bool
}

type Options struct {
	GenerateClient bool
	GenerateServer bool
}

type TemplateData struct {
	Options   Options
	ProtoFile *protogen.File
}

func New(options Options) (*Generator, error) {
	templates, err := template.ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		return nil, err
	}

	return &Generator{
		options:            options,
		templates:          templates,
		commonGeneratedMap: make(map[string]bool),
	}, nil
}

func toImportPath(goImportPath protogen.GoImportPath) string {
	importPath, err := strconv.Unquote(goImportPath.String())
	if err != nil {
		panic(err)
	}
	return importPath
}

func toQual(goIdent protogen.GoIdent) *jen.Statement {
	return jen.Qual(toImportPath(goIdent.GoImportPath), goIdent.GoName)
}

func (g *Generator) Generate(gen *protogen.Plugin) error {
	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}

		if len(f.Services) == 0 {
			continue
		}

		importPath := toImportPath(f.GoImportPath)
		syncInterfaceFile := jen.NewFilePath(importPath)
		asyncInterfaceFile := jen.NewFilePath(importPath)

		for _, service := range f.Services {
			syncInterfaceType := syncInterfaceFile.Type().Id(service.GoName)
			asyncInterfaceType := asyncInterfaceFile.Type().Id(service.GoName + "Async")
			var syncInterfaceStatements []jen.Code
			var asyncInterfaceStatements []jen.Code
			for _, method := range service.Methods {
				if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
					// TODO streaming currently not supported
					continue
				}
				syncInterfaceStatements = append(syncInterfaceStatements,
					jen.Id(method.GoName).
						Params(
							jen.Qual("context", "Context"),
							jen.Op("*").Add(toQual(method.Input.GoIdent)),
						).
						Params(
							jen.Op("*").Add(toQual(method.Output.GoIdent)),
							jen.Id("error"),
						),
				)
				asyncInterfaceStatements = append(asyncInterfaceStatements,
					jen.Id(method.GoName).
						Params(
							jen.Qual("context", "Context"),
							jen.Op("*").Add(toQual(method.Input.GoIdent)),
						).
						Op("*").Qual("github.com/samber/mo", "Future").Types(jen.Op("*").Add(toQual(method.Output.GoIdent))),
					//.Add(toQual(method.Output.GoIdent)),
				)
			}
			syncInterfaceType.Interface(syncInterfaceStatements...)
			asyncInterfaceType.Interface(asyncInterfaceStatements...)
		}

		_, err := gen.NewGeneratedFile(fmt.Sprintf("%s.sync.interface.go", f.GeneratedFilenamePrefix), f.GoImportPath).
			Write([]byte(fmt.Sprintf("%#v", syncInterfaceFile)))
		if err != nil {
			return err
		}

		_, err = gen.NewGeneratedFile(fmt.Sprintf("%s.async.interface.go", f.GeneratedFilenamePrefix), f.GoImportPath).
			Write([]byte(fmt.Sprintf("%#v", asyncInterfaceFile)))
		if err != nil {
			return err
		}
	}
	return nil
}
