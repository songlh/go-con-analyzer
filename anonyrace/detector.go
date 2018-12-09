package anonyrace

import (
	"fmt"
	"reflect"
	"golang.org/x/tools/go/ssa"
	//"github.com/songlh/GCatch/ssabuilder"
)

func CollectSharedVariable(f *ssa.Function) map[*ssa.Value] *ssa.Instruction {

	var m map[*ssa.Value] *ssa.Instruction

	for _, BB := range f.Blocks {
		for _, II := range BB.Instrs {
			switch inst := II.(type) {
			case *ssa.Store:
				fmt.Println(reflect.TypeOf(inst.Addr))
				switch addr := inst.Addr.(type) {
				case *ssa.IndexAddr:
					fmt.Println(reflect.TypeOf(addr))
				case *ssa.FreeVar:
					//fmt.Printf("parent %d\n", addr.Parent().Name())
					//if addr.Parent() != f {
					//	m[]
					//}

					for _, useInst := range *(addr.Referrers()) {
						fmt.Printf("free use parent: %s\n", useInst.Parent())
					}
				}
			}

		}
	}

	return m
}



func CheckFunction(f *ssa.Function) {
	fmt.Println(f.Name())
	for _, anonyf := range f.AnonFuncs {
		CollectSharedVariable(anonyf)
	}
}



func Check(mainPkg *ssa.Package) {
	for name, member := range mainPkg.Members {
		//fmt.Printf("%s %s\n", name, reflect.TypeOf(member))
		switch f := member.(type) {
		case *ssa.Function:
			if name == "init" {
				break
			}
			CheckFunction(f)
		}
	}	
}