package main 

import (
	"log"
	"fmt"

	//"golang.org/x/tools/go/ssa"

	"github.com/songlh/GCatch/ssabuilder"
	"github.com/songlh/GCatch/anonyrace"
)



func main() {
	//fmt.Println("Hello world")
	files := []string{"tests/anonyrace1/anonyrace1.go"}

	fmt.Println(files)

	conf, err := ssabuilder.NewConfig(files)
	if err != nil {
		log.Fatal(err)
	}

	ssainfo, err := conf.Build()
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println(reflect.TypeOf(ssainfo.Prog))

	//for _, pkg := range ssainfo.Prog.AllPackages() {
	//	fmt.Println(pkg.Pkg.Name())	
	//}

	mainPkg := ssabuilder.GetMainPkg(ssainfo.Prog)

	if mainPkg == nil {
		fmt.Println("Fail to find main package")
	}

	//for _, m := range mainPkg.Members {
	//	fmt.Println(reflect.TypeOf(m))
	//}

	//fmain := mainPkg.Func("main")

	anonyrace.Check(mainPkg)


	//anonyList := fmain.AnonFuncs


	//ssabuilder.PrintCFG(fmain)


	//fmt.Println(len(anonyList))
	//printCFG(anonyList[0])

	//printCFG(fmain)
	//printBasicBlock(fmain.Blocks[0])

}