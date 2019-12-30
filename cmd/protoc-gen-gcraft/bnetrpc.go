package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	"github.com/superp00t/etc"

	"github.com/superp00t/gophercraft/bnet/bgs/protocol"
	"github.com/superp00t/gophercraft/cmd/protoc-gen-gcraft/generator"
)

func trace(args ...interface{}) {
	dat := fmt.Sprintln(args...)
	dat = strings.Replace(dat, "\n", "\r\n", -1)
	f, _ := os.OpenFile(etc.TmpDirectory().Concat("protoc.txt").Render(), os.O_APPEND|os.O_RDWR, 0700)
	f.Write([]byte(dat))
	f.Close()
}

type bnetRPC struct {
	// gen     *generator.Generator
	// file    *generator.FileDescriptor
	imports [][]string
	out     *etc.Buffer
}

func (b *bnetRPC) require(i string) string {
	for _, imp := range b.imports {
		if imp[1] == i {
			return imp[0]
		}
	}

	path := strings.Split(i, "/")
	lhImport := path[len(path)-1]

start:

	for _, imp := range b.imports {
		if imp[0] == lhImport {
			lhImport = path[len(path)-2] + lhImport
			goto start
		}
	}

	b.imports = append(b.imports, []string{lhImport, i})

	return lhImport
}

func (b *bnetRPC) Name() string {
	return "bnet_rpc"
}

func (b *bnetRPC) Init(gen *generator.Generator) {
	// for k, v := range gen.ImportMap {
	// 	b.imports = append(b.imports, []string{k, v})
	// }
	// b.gen = gen
	b.imports = [][]string{
		{"proto", "github.com/golang/protobuf/proto"},
		{"fmt", "fmt"},
		{"math", "math"},
	}
}

func (b *bnetRPC) GenerateImports(fd *generator.FileDescriptor) {
}

func (b *bnetRPC) inputType(s string) string {
	sd := strings.Split(s, ".")[1:]
	endType := sd[len(sd)-1:][0]
	sd = sd[:len(sd)-1]
	dat := strings.Join(sd, "/")

	packageType := b.require("github.com/superp00t/gophercraft/bnet/" + dat)
	return "*" + packageType + "." + endType
}

func (b *bnetRPC) P(args ...interface{}) {
	for _, v := range args {
		b.out.Write([]byte(fmt.Sprint(v)))
	}
	b.out.Write([]byte{'\n'})
}

func purify(ot string) string {
	if ot == "*protocol.NoData" || ot == "*protocol.NO_RESPONSE" {
		ot = ""
	}

	return ot
}

func (b *bnetRPC) Generate(fd *generator.FileDescriptor) {
	// trace(spew.Sdump(b.Pkg))

	for _, svc := range fd.Service {
		path := etc.ParseSystemPath(os.Getenv("GOPATH")).Concat("src", "github.com", "superp00t", "gophercraft", "bnet", "x_svc_"+svc.GetName()+".go")

		path.Remove()

		b.out = etc.NewBuffer()

		hash := HashServiceName(svc.GetName())

		if proto.HasExtension(svc.GetOptions(), protocol.E_ServiceOptions) {
			svco, _ := proto.GetExtension(svc.GetOptions(), protocol.E_ServiceOptions)
			hash = HashServiceName(svco.(*protocol.BGSServiceOptions).GetDescriptorName())
		}

		b.P("const ", svc.GetName(), "Hash = ", hash)
		b.P()

		b.P("type ", svc.GetName(), " interface {")

		for _, m := range svc.GetMethod() {
			b.P("\t", m.GetName()+"(*Conn, uint32, "+b.inputType(m.GetInputType())+")")
		}

		b.P("}")
		b.P()

		// error: should fix
		b.P("func Dispatch" + svc.GetName() + "(conn *Conn, svc " + svc.GetName() + ", token uint32, method uint32, data []byte) error {")
		b.P("switch method {")

		for _, v := range svc.GetMethod() {
			if !proto.HasExtension(v.GetOptions(), protocol.E_MethodOptions) {
				continue
			}

			id, err := proto.GetExtension(v.GetOptions(), protocol.E_MethodOptions)
			if err != nil {
				trace(err)
				panic(err)
			}

			trace(spew.Sdump(id))
			b.P(fmt.Sprintf("case %d:", id.(*protocol.BGSMethodOptions).GetId()))
			b.P("var args ", strings.TrimLeft(b.inputType(v.GetInputType()), "*"))
			b.P("err := proto.Unmarshal(data, &args)")
			b.P("if err != nil { return err }")
			b.P("svc." + v.GetName() + "(conn, token, &args)")
		}
		b.P("}")
		b.P("return nil")
		b.P("}")
		b.P("")
		b.P("type Empty", svc.GetName(), " struct {}")
		b.P("")

		for _, m := range svc.GetMethod() {
			b.P("func (e Empty", svc.GetName(), ") ", m.GetName(), "(conn *Conn, token uint32, args "+b.inputType(m.GetInputType())+") {")
			b.P("\tconn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)")
			b.P("}")
		}

		b.P()

		for _, m := range svc.GetMethod() {
			it := purify(b.inputType(m.GetInputType()))
			ot := purify(b.inputType(m.GetOutputType()))

			outSig := "error"
			if ot != "" {
				outSig = "(" + ot + ", error)"
			}

			inSig := ""
			if it != "" {
				inSig = "args " + it
			}

			mext, _ := proto.GetExtension(m.GetOptions(), protocol.E_MethodOptions)
			mid := mext.(*protocol.BGSMethodOptions).GetId()

			bt := "_"

			if ot != "" {
				bt = "bytes"
			}

			b.P("func (c *Conn) ", svc.GetName(), "_Request_", m.GetName(), "(", inSig, ") ", outSig, " {")
			if it != "" {
				b.P("header, ", bt, ", err := c.Request(", svc.GetName(), "Hash, ", mid, ", args)")
			} else {
				b.P("header, ", bt, ", err := c.Request(", svc.GetName(), "Hash, ", mid, ", nil)")
			}

			b.P("if err != nil {")

			if ot == "" {
				b.P("return err")
			} else {
				b.P("return nil, err")
			}
			b.P("}")
			b.P("if Status(header.GetStatus()) != ERROR_OK {")
			t := `fmt.Errorf("bnet: non-ok status 0x%08X", header.GetStatus())`
			if ot == "" {
				b.P("return " + t)
			} else {
				b.P("return nil, " + t)
			}
			b.P("}")

			if ot == "" {
				b.P("return nil")
			} else {
				b.P("var out ", ot[1:])
				b.P("err = proto.Unmarshal(bytes, &out)")
				b.P("if err != nil { return nil, err }")
				b.P("return &out, nil")
			}
			b.P("}")
			b.P()
		}

		dat := b.out.ToString()
		b.out = etc.NewBuffer()
		b.P("// generated by protoc-gen-gcraft : DO NOT EDIT")
		b.P("package bnet")
		b.P()
		b.P("import (")
		for _, v := range b.imports {
			b.P("\t", v[0], " \"", v[1], "\"")
		}
		b.P(")")
		b.P()

		b.P("// shut the compiler up")
		// shut the compiler up
		for _, v := range []string{
			"proto.Marshal",
			"fmt.Errorf",
			"math.Inf",
			"protocol.E_MethodOptions",
		} {
			b.P("var _ = ", v)
		}
		b.P()
		b.out.Write([]byte(dat))
		ioutil.WriteFile(path.Render(), b.out.Bytes(), 0700)
		exec.Command("gofmt", "-w", path.Render()).Run()
		b.out = etc.NewBuffer()
	}
}

func HashServiceName(name string) string {
	var hash uint32 = 0x811C9DC5
	for i := 0; i < len(name); i++ {
		hash ^= uint32(name[i])
		hash *= 0x1000193
	}

	return fmt.Sprintf("0x%08X", hash)
}
