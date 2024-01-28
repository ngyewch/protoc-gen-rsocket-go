package main

import (
	"flag"
	"github.com/ngyewch/protoc-gen-rsocket-go/generator"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	var flags flag.FlagSet
	genClient := flags.Bool("gen-client", false, "Generate client")
	genServer := flags.Bool("gen-server", false, "Generate server")
	opts := &protogen.Options{
		ParamFunc: flags.Set,
	}
	opts.Run(func(plugin *protogen.Plugin) error {
		g := generator.New(generator.Options{
			GenerateClient: *genClient,
			GenerateServer: *genServer,
		})
		return g.Generate(plugin)
	})
}
