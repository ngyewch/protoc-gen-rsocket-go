package generator

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/compiler/protogen"
	"strconv"
)

const (
	protoPackage   = "google.golang.org/protobuf/proto"
	runtimePackage = "github.com/ngyewch/protoc-gen-rsocket-go/runtime"
	moPackage      = "github.com/samber/mo"
	rsocketPackage = "github.com/rsocket/rsocket-go"
)

type Generator struct {
	options            Options
	commonGeneratedMap map[string]bool
}

type Options struct {
	GenerateClient  bool
	GenerateServer  bool
	GenerateSync    bool
	GenerateAsync   bool
	GenerateRSocket bool
}

type TemplateData struct {
	Options   Options
	ProtoFile *protogen.File
}

func New(options Options) *Generator {
	return &Generator{
		options:            options,
		commonGeneratedMap: make(map[string]bool),
	}
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
		syncClientFile := jen.NewFilePath(importPath)
		syncServerFile := jen.NewFilePath(importPath)
		asyncClientFile := jen.NewFilePath(importPath)
		asyncServerFile := jen.NewFilePath(importPath)

		for _, service := range f.Services {
			syncInterfaceType := syncInterfaceFile.Type().Id(service.GoName)
			asyncInterfaceType := asyncInterfaceFile.Type().Id(service.GoName + "Async")

			syncClientStructName := service.GoName + "Client"
			syncClientFile.Type().Id(syncClientStructName).Struct(
				jen.Id("selector").Id("uint64"),
				jen.Id("handler").Qual(runtimePackage, "ClientRequestResponseHandler"),
			)
			syncClientFile.Func().Id("New"+syncClientStructName).
				Params(
					jen.Id("selector").Id("uint64"),
					jen.Id("handler").Qual(runtimePackage, "ClientRequestResponseHandler"),
				).
				Op("*").Id(syncClientStructName).
				Block(
					jen.Return(jen.Op("&").Id(syncClientStructName)).Values(jen.Dict{
						jen.Id("selector"): jen.Id("selector"),
						jen.Id("handler"):  jen.Id("handler"),
					}),
				)
			if g.options.GenerateRSocket {
				syncClientFile.Func().Id("NewRSocket"+syncClientStructName).
					Params(
						jen.Id("selector").Id("uint64"),
						jen.Id("rs").Qual(rsocketPackage, "RSocket"),
					).
					Op("*").Id(syncClientStructName).
					Block(
						jen.Return(jen.Op("&").Id(syncClientStructName)).Values(jen.Dict{
							jen.Id("selector"): jen.Id("selector"),
							jen.Id("handler"):  jen.Qual(runtimePackage, "RSocketClientRequestResponseHandler").Call(jen.Id("rs")),
						}),
					)
			}

			syncServerStructName := service.GoName + "Server"
			syncServerFile.Type().Id(syncServerStructName).Struct(
				jen.Id("selector").Id("uint64"),
				jen.Id("service").Id(service.GoName),
			)
			syncServerFile.Func().Id("New"+syncServerStructName).
				Params(
					jen.Id("selector").Id("uint64"),
					jen.Id("service").Id(service.GoName),
				).
				Op("*").Id(syncServerStructName).
				Block(
					jen.Return(jen.Op("&").Id(syncServerStructName)).Values(jen.Dict{
						jen.Id("selector"): jen.Id("selector"),
						jen.Id("service"):  jen.Id("service"),
					}),
				)
			syncServerFile.Func().
				Params(
					jen.Id("s").Op("*").Id(syncServerStructName),
				).
				Id("Selector").
				Params().
				Id("uint64").
				Block(
					jen.Return(jen.Id("s").Op(".").Id("selector")),
				)

			asyncClientStructName := service.GoName + "ClientAsync"
			asyncClientFile.Type().Id(asyncClientStructName).Struct(
				jen.Id("selector").Id("uint64"),
				jen.Id("handler").Qual(runtimePackage, "ClientRequestResponseHandlerAsync"),
			)
			asyncClientFile.Func().Id("New"+asyncClientStructName).
				Params(
					jen.Id("selector").Id("uint64"),
					jen.Id("handler").Qual(runtimePackage, "ClientRequestResponseHandlerAsync"),
				).
				Op("*").Id(asyncClientStructName).
				Block(
					jen.Return(jen.Op("&").Id(asyncClientStructName)).Values(jen.Dict{
						jen.Id("selector"): jen.Id("selector"),
						jen.Id("handler"):  jen.Id("handler"),
					}),
				)
			if g.options.GenerateRSocket {
				asyncClientFile.Func().Id("NewRSocket"+asyncClientStructName).
					Params(
						jen.Id("selector").Id("uint64"),
						jen.Id("rs").Qual(rsocketPackage, "RSocket"),
					).
					Op("*").Id(asyncClientStructName).
					Block(
						jen.Return(jen.Op("&").Id(asyncClientStructName)).Values(jen.Dict{
							jen.Id("selector"): jen.Id("selector"),
							jen.Id("handler"):  jen.Qual(runtimePackage, "RSocketClientRequestResponseHandlerAsync").Call(jen.Id("rs")),
						}),
					)
			}

			var syncInterfaceStatements []jen.Code
			var asyncInterfaceStatements []jen.Code
			var syncServerCases []jen.Code
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
						Op("*").Qual(moPackage, "Future").Types(jen.Op("*").Add(toQual(method.Output.GoIdent))),
				)

				syncClientFile.Func().
					Params(
						jen.Id("c").Op("*").Id(syncClientStructName),
					).
					Id(method.GoName).
					Params(
						jen.Id("ctx").Qual("context", "Context"),
						jen.Id("req").Op("*").Add(toQual(method.Input.GoIdent)),
					).
					Params(
						jen.Op("*").Add(toQual(method.Output.GoIdent)),
						jen.Id("error"),
					).
					Block(
						jen.List(
							jen.Id("rspBytes"),
							jen.Id("err"),
						).Op(":=").Qual(runtimePackage, "HandleClientRequestResponse").Call(
							jen.Id("ctx"),
							jen.Id("c").Op(".").Id("selector"),
							jen.Lit(string(method.Desc.Name())),
							jen.Id("req"),
							jen.Id("c").Op(".").Id("handler"),
						),
						jen.If(jen.Id("err").Op("!=").Nil()).Block(
							jen.Return(
								jen.Nil(),
								jen.Id("err"),
							),
						),
						jen.Var().Id("rsp").Add(toQual(method.Output.GoIdent)),
						jen.Id("err").Op("=").Qual(protoPackage, "Unmarshal").Call(
							jen.Id("rspBytes"),
							jen.Op("&").Id("rsp"),
						),
						jen.If(jen.Id("err").Op("!=").Nil()).Block(
							jen.Return(
								jen.Nil(),
								jen.Id("err"),
							),
						),
						jen.Return(
							jen.Op("&").Id("rsp"),
							jen.Nil(),
						),
					)

				asyncClientFile.Func().
					Params(
						jen.Id("c").Op("*").Id(asyncClientStructName),
					).
					Id(method.GoName).
					Params(
						jen.Id("ctx").Qual("context", "Context"),
						jen.Id("req").Op("*").Add(toQual(method.Input.GoIdent)),
					).
					Params(
						jen.Op("*").Qual(moPackage, "Future").Types(jen.Op("*").Add(toQual(method.Output.GoIdent))),
					).
					Block(
						jen.Return().Qual(moPackage, "NewFuture").Call(
							jen.Func().
								Params(
									jen.Id("resolve").Func().
										Params(
											jen.Op("*").Add(toQual(method.Output.GoIdent)),
										),
									jen.Id("reject").Func().
										Params(
											jen.Id("error"),
										),
								).
								Block(
									jen.Qual(runtimePackage, "HandleClientRequestResponseAsync").
										Call(
											jen.Id("ctx"),
											jen.Id("c").Op(".").Id("selector"),
											jen.Lit(string(method.Desc.Name())),
											jen.Id("req"),
											jen.Id("c").Op(".").Id("handler"),
										).
										Op(".").Line().
										Id("Catch").
										Call(
											jen.Func().
												Params(
													jen.Id("err").Id("error"),
												).
												Params(
													jen.Id("[]byte"),
													jen.Id("error"),
												).
												Block(
													jen.Id("reject").Call(jen.Id("err")),
													jen.Return(
														jen.Nil(),
														jen.Id("err"),
													),
												),
										).
										Op(".").Line().
										Id("Then").
										Call(
											jen.Func().
												Params(
													jen.Id("rspBytes").Id("[]byte"),
												).
												Params(
													jen.Id("[]byte"),
													jen.Id("error"),
												).
												Block(
													jen.Var().Id("rsp").Add(toQual(method.Output.GoIdent)),
													jen.Id("err").Op(":=").Qual(protoPackage, "Unmarshal").Call(
														jen.Id("rspBytes"),
														jen.Op("&").Id("rsp"),
													),
													jen.If(jen.Id("err").Op("!=").Nil()).Block(
														jen.Return(
															jen.Nil(),
															jen.Id("err"),
														),
													),
													jen.Id("resolve").Call(jen.Op("&").Id("rsp")),
													jen.Return(
														jen.Id("rspBytes"),
														jen.Nil(),
													),
												),
										),
								),
						),
					)

				syncServerCases = append(syncServerCases, jen.Case(jen.Lit(string(method.Desc.Name()))).
					Block(
						jen.Var().Id("req").Add(toQual(method.Input.GoIdent)),
						jen.Id("err").Op(":=").Qual(protoPackage, "Unmarshal").Call(
							jen.Id("reqWrapper").Op(".").Id("Payload"),
							jen.Op("&").Id("req"),
						),
						jen.If(jen.Id("err").Op("!=").Nil()).Block(
							jen.Return(
								jen.Nil(),
								jen.Id("err"),
							),
						),
						jen.Return(
							jen.Id("s").Op(".").Id("service").Op(".").Id(method.GoName).Call(
								jen.Id("ctx"),
								jen.Op("&").Id("req"),
							),
						),
					),
				)
			}

			syncInterfaceType.Interface(syncInterfaceStatements...)
			asyncInterfaceType.Interface(asyncInterfaceStatements...)

			syncServerCases = append(syncServerCases, jen.Default().
				Block(
					jen.Return(
						jen.Nil(),
						jen.Qual("fmt", "Errorf").Call(
							jen.Lit("unknown method: %s"),
							jen.Id("reqWrapper").Op(".").Id("MethodName"),
						),
					),
				),
			)
			syncServerFile.Func().
				Params(
					jen.Id("s").Op("*").Id(syncServerStructName),
				).
				Id("HandleRequestResponse").
				Params(
					jen.Id("ctx").Qual("context", "Context"),
					jen.Id("reqWrapper").Op("*").Qual(runtimePackage, "RequestWrapper"),
				).
				Params(
					jen.Qual(protoPackage, "Message"),
					jen.Id("error"),
				).
				Block(
					jen.If(jen.Id("reqWrapper").Op(".").Id("Selector").Op("!=").Id("s").Op(".").Id("selector").
						Block(
							jen.Return(
								jen.Nil(),
								jen.Qual(runtimePackage, "ErrorSelectorMismatch"),
							),
						)),
					jen.Switch(jen.Id("reqWrapper").Op(".").Id("MethodName").
						Block(
							syncServerCases...,
						)),
				)
		}

		if g.options.GenerateSync {
			_, err := gen.NewGeneratedFile(fmt.Sprintf("%s.sync.interface.go", f.GeneratedFilenamePrefix), f.GoImportPath).
				Write([]byte(fmt.Sprintf("%#v", syncInterfaceFile)))
			if err != nil {
				return err
			}

			if g.options.GenerateClient {
				_, err = gen.NewGeneratedFile(fmt.Sprintf("%s.sync.client.go", f.GeneratedFilenamePrefix), f.GoImportPath).
					Write([]byte(fmt.Sprintf("%#v", syncClientFile)))
				if err != nil {
					return err
				}
			}

			if g.options.GenerateServer {
				_, err = gen.NewGeneratedFile(fmt.Sprintf("%s.sync.server.go", f.GeneratedFilenamePrefix), f.GoImportPath).
					Write([]byte(fmt.Sprintf("%#v", syncServerFile)))
				if err != nil {
					return err
				}
			}
		}

		if g.options.GenerateAsync {
			_, err := gen.NewGeneratedFile(fmt.Sprintf("%s.async.interface.go", f.GeneratedFilenamePrefix), f.GoImportPath).
				Write([]byte(fmt.Sprintf("%#v", asyncInterfaceFile)))
			if err != nil {
				return err
			}

			if g.options.GenerateClient {
				_, err = gen.NewGeneratedFile(fmt.Sprintf("%s.async.client.go", f.GeneratedFilenamePrefix), f.GoImportPath).
					Write([]byte(fmt.Sprintf("%#v", asyncClientFile)))
				if err != nil {
					return err
				}
			}

			if g.options.GenerateServer {
				_, err = gen.NewGeneratedFile(fmt.Sprintf("%s.async.server.go", f.GeneratedFilenamePrefix), f.GoImportPath).
					Write([]byte(fmt.Sprintf("%#v", asyncServerFile)))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
