package main

import (
	"flag"
	"github.com/ngyewch/protoc-gen-rsocket-go/generator"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	var flags flag.FlagSet
	genClient := flags.Bool("gen-client", true, "Generate client")
	genServer := flags.Bool("gen-server", true, "Generate server")
	genSync := flags.Bool("gen-sync", true, "Generate sync")
	genAsync := flags.Bool("gen-async", true, "Generate async")
	genRSocket := flags.Bool("gen-rsocket", true, "Generate rsocket")
	opts := &protogen.Options{
		ParamFunc: flags.Set,
	}
	opts.Run(func(plugin *protogen.Plugin) error {
		g := generator.New(generator.Options{
			GenerateClient:  *genClient,
			GenerateServer:  *genServer,
			GenerateSync:    *genSync,
			GenerateAsync:   *genAsync,
			GenerateRSocket: *genRSocket,
		})
		return g.Generate(plugin)
	})
}
