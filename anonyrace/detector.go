package anonyrace

import (
	"fmt"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/cfg"
	"golang.org/x/tools/go/ssa"
	"os"
	"reflect"
	"sort"
)

// For PrintAll
type members []ssa.Member
func (m members) Len() int           { return len(m) }
func (m members) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m members) Less(i, j int) bool { return m[i].Pos() < m[j].Pos() }

// Print all contents in all public functions
func PrintAll(mainPkg *ssa.Package) {
	funcs := members([]ssa.Member{})
	for _, obj := range mainPkg.Members {
		if obj.Token() == token.FUNC {
			funcs = append(funcs, obj)
		}
	}
	// sort by Pos()
	sort.Sort(funcs)
	for _, f := range funcs {
		mainPkg.Func(f.Name()).WriteTo(os.Stdout)
		for _, anonFunc := range mainPkg.Func(f.Name()).AnonFuncs {
			anonFunc.WriteTo(os.Stdout)
		}
	}
}

// Deprecated: Print Instructions recursively, use printAll instead
func PrintRecursive(f *ssa.Function) {
	fmt.Println(f)
	for _, BB := range f.Blocks {
		fmt.Println("BB", BB)
		for _, II := range BB.Instrs {
			fmt.Println(II.Pos(), reflect.TypeOf(II), II)
		}
	}
	for _, anonFunc := range f.AnonFuncs {
		PrintRecursive(anonFunc)
	}
}

// Deprecated: For PrintCFG
// A trivial mayReturn predicate that looks only at syntax, not types.
func mayReturn(call *ast.CallExpr) bool {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		return fun.Name != "panic"
	case *ast.SelectorExpr:
		return fun.Sel.Name != "Fatal"
	}
	return true
}

// Deprecated: Print CFG for check
func PrintCFG(f *ssa.Function) {
	fmt.Println(f)
	node := f.Syntax()
	if decl, ok := node.(*ast.FuncDecl); ok {
		g := cfg.New(decl.Body, mayReturn)
		for _, BB := range g.Blocks {
			fmt.Println(BB)
		}
	}
}

// Check if GoInst dom StoreInst
// if in same block, then compare index
// if in different blocks, check block dom block
func InstrDominates(lhs *ssa.Go, rhs *ssa.Store) bool {
	lBlock := (*lhs).Block()
	rBlock := (*rhs).Block()

	if lBlock == rBlock {
		//fmt.Println("Go.Pos()", (*lhs).Pos(), "Store.Pos()", (*rhs).Pos())
		//return (*lhs).Pos() < (*rhs).Pos()
		lIndex := 0
		rIndex := 0
		var instGo *ssa.Go = nil
		var instStore *ssa.Store = nil

		for index, II := range lBlock.Instrs {

			if instGo == nil {
				if tmpInstGo := tryParseGo(&II); tmpInstGo != nil {
					if lhs == tmpInstGo {
						lIndex = index
						instGo = tmpInstGo
					}
				}
			}

			if instStore == nil {
				if tmpInstStore := tryParseStore(&II); tmpInstStore != nil {
					if rhs == tmpInstStore {
						rIndex = index
						instStore = tmpInstStore
					}
				}
			}


			if instGo != nil && instStore != nil {
				return lIndex < rIndex
			}
		}

		fmt.Println("Go Store not found in Block")
		return false
	} else {
		//fmt.Println("Blocks:", lBlock, rBlock)
		return lBlock.Dominates(rBlock)
	}
}


//func CollectSharedVariable_Old(f *ssa.Function) map[*ssa.Value] *ssa.Instruction {
//
//	var m map[*ssa.Value] *ssa.Instruction
//
//	//fmt.Println("AnonFunc: ", f)
//
//	for _, BB := range f.Blocks {
//		//fmt.Println("BB: ", BB)
//		for _, II := range BB.Instrs {
//		//	switch inst := II.(type) {
//		//	case *ssa.Store:
//		//		//fmt.Println("Store: ", reflect.TypeOf(inst.Addr))
//		//		switch addr := inst.Addr.(type) {
//		//		case *ssa.IndexAddr:
//		//			fmt.Println("IndexAddr: ", reflect.TypeOf(addr))
//		//
//		//		case *ssa.FreeVar:
//		//			//fmt.Printf("parent %d\n", addr.Parent().Name())
//		//			//if addr.Parent() != f {
//		//			//	m[]
//		//			//}
//		//
//		//			for _, useInst := range *(addr.Referrers()) {
//		//				fmt.Printf("free use parent: %s\n", useInst.Parent())
//		//
//		//			}
//		//		}
//		//
//		//	case *ssa.UnOp:
//		//		unOp := II.(*ssa.UnOp)
//		//		if unOp.Op == token.MUL {
//		//			fmt.Println(unOp.Op)
//		//			fmt.Println(unOp.X.Referrers())
//		//			for _, useInst := range *(unOp.X.Referrers()) {
//		//				fmt.Printf("free use parent: %s\n", useInst.Parent())
//		//			}
//		//		}
//		//	}
//			fmt.Println(reflect.TypeOf(II), II)
//		}
//	}
//
//	return m
//}
//
//func CollectSharedVariable(f *ssa.Function) map[*ssa.Value] *ssa.Instruction {
//
//	var m map[*ssa.Value] *ssa.Instruction
//
//	for _, BB := range f.Blocks {
//		for _, II := range BB.Instrs {
//			switch II.(type) {
//			case *ssa.UnOp:
//				unOp := II.(*ssa.UnOp)
//				if unOp.Op == token.MUL { // t0 = *x
//					switch unOp.X.(type) {
//					case *ssa.FreeVar:
//						freeVar := unOp.X.(*ssa.FreeVar)
//						fmt.Println("load")
//						for _, referrer := range *freeVar.Referrers() {
//							fmt.Println(referrer, ":", referrer.Parent())
//						}
//					}
//				}
//			case *ssa.Store: // *x = y
//				store := II.(*ssa.Store)
//				switch store.Addr.(type) {
//				case *ssa.FreeVar:
//					freeVar := store.Addr.(*ssa.FreeVar)
//					fmt.Println("store")
//					for _, referrer := range *freeVar.Referrers() {
//						fmt.Println(referrer, ":", referrer.Parent())
//					}
//				}
//			}
//		}
//	}
//
//	return m
//}
//
//func CheckFunction(f *ssa.Function) {
//	//fmt.Println("func name:", f.Name())
//	for _, BB := range f.Blocks {
//
//		//fmt.Println("func BB: ", BB)
//
//		for _, II := range BB.Instrs {
//			switch II.(type) {
//			case *ssa.Go:
//				var rands[] *ssa.Value
//				myRands := II.Operands(rands)
//				//fmt.Println("Go: ", II)
//				for _, r := range myRands {
//					switch (*r).(type) {
//					case *ssa.MakeClosure:
//						//fmt.Println("make closure: ", *r)
//						a := (*r).(*ssa.MakeClosure)
//
//						// fmt.Println("$", a.Bindings)
//
//						for _, binding := range a.Bindings {
//							for _, referrer := range *binding.Referrers() {
//								fmt.Println("$", binding, referrer, referrer.Parent())
//
//								fmt.Println("T", reflect.TypeOf(referrer))
//							}
//						}
//
//						var rands2[] *ssa.Value
//						myRands2 := a.Operands(rands2)
//
//						for _, r2 := range myRands2 {
//							// fmt.Println(reflect.TypeOf(*r2), ":", *r2)
//
//							switch (*r2).(type) {
//							case *ssa.Function:
//								var myFunc = (*r2).(*ssa.Function)
//								for _, anonyf := range f.AnonFuncs {
//									if myFunc == anonyf {
//										CollectSharedVariable(anonyf)
//									}
//								}
//
//							}
//
//						}
//
//
//					}
//
//
//				}
//			}
//			//fmt.Println(reflect.TypeOf(II), ":", II)
//		}
//	}
//
//	//for _, anonyf := range f.AnonFuncs {
//	//	// TODO: go func() // check go
//	//	CollectSharedVariable(anonyf)
//	//}
//}

// Record LOAD/STORE instructions refer to shared var
type SharedVarReferrer struct {
	LoadInsts []*ssa.UnOp
	StoreInsts []*ssa.Store
}

func (svr SharedVarReferrer) String() string {
	var formatStr string
	formatStr += fmt.Sprintf("Load: %v\t", svr.LoadInsts)
	formatStr += fmt.Sprintf("Store: %v", svr.StoreInsts)
	return formatStr
}

// Record AnonFunc and CallerFunc referrers of the shared var
type ResultInfo struct {
	AnonFunc *ssa.Function
	AnonReferrer SharedVarReferrer
	CallerVar ssa.Value
	CallerFunc *ssa.Function
	CallerReferrer SharedVarReferrer
}

type Results map[*ssa.FreeVar] ResultInfo

func (results Results) String() string {

	var formatStr string

	for freeVar, resultInfo := range results {
		formatStr += fmt.Sprintf("AnonVar: %v\n\tAnonFunc: %v\n\tAnonReferrer: %v\nCallerVar: %v\n\tCallerFunc: %v\n\tCallerReferrer: %v\n",
			freeVar, resultInfo.AnonFunc, resultInfo.AnonReferrer, resultInfo.CallerVar, resultInfo.CallerFunc, resultInfo.CallerReferrer)
	}

	return formatStr
}

// If AnonFunc is called by Go, return the MakeClosure and Go instructions
// TODO: AnonFunc can be called by other functions, but cannot exhaust all of them now
func GetMakeClosureCalledByGo(anonFunc *ssa.Function) (*ssa.MakeClosure, *ssa.Go, bool) {

	for _, referrer := range *anonFunc.Referrers() {

		if makeClosure := tryParseMakeClosure(&referrer); makeClosure != nil {

			for _, mReferrer := range *makeClosure.Referrers() {

				if instGo := tryParseGo(&mReferrer); instGo != nil {
					return makeClosure, instGo, true
				}
			}
		}
	}

	return nil, nil, false
}


// Record LOAD/STORE instructions that refer to MakeClosure.bindings in CallerFunc
func RecordVarsInCallerFunc(bindings []ssa.Value) map[ssa.Value] SharedVarReferrer {

	m := make(map[ssa.Value] SharedVarReferrer)

	for _, binding := range bindings {

		var svr SharedVarReferrer

		for _, referrer := range *binding.Referrers() {

			if load := tryParseLoad(&referrer); load != nil {
				svr.LoadInsts = append(svr.LoadInsts, load)
			}

			if store := tryParseStore(&referrer); store != nil {
				svr.StoreInsts = append(svr.StoreInsts, store)
			}
		}

		m[binding] = svr
	}

	return m
}

// Record LOAD/STORE instructions that refer to anonFunc.FreeVar in AnonFunc
func RecordVarsInAnonFunc(freeVars []*ssa.FreeVar) map[*ssa.FreeVar] SharedVarReferrer {

	m := make(map[*ssa.FreeVar] SharedVarReferrer)

	for _, freeVar := range freeVars {

		var svr SharedVarReferrer

		for _, referrer := range *freeVar.Referrers() {

			if load := tryParseLoad(&referrer); load != nil {
				svr.LoadInsts = append(svr.LoadInsts, load)
			}

			if store := tryParseStore(&referrer); store != nil {
				svr.StoreInsts = append(svr.StoreInsts, store)
			}
		}

		m[freeVar] = svr
	}

	return m
}

// Map FreeVar in AnonFunc to bindings in CallerFunc
func GetMapAnon2Caller(freeVars []*ssa.FreeVar, bindings []ssa.Value) map[*ssa.FreeVar] ssa.Value {
	anon2caller := make(map[*ssa.FreeVar] ssa.Value)

	if len(freeVars) != len(bindings) {
		fmt.Println("len(anonFunc.FreeVars) != len(makeClosure.Bindings)")
		return nil
	}

	for i := 0; i < len(freeVars); i++ {
		anon2caller[freeVars[i]] = bindings[i]
	}

	return anon2caller
}

// Note: I cannot find a way to get ALL the functions by traversing pkg.Members
// because member functions are missed. Only public functions are listed.
// I have to find all the Call instructions to get the inner functions, maybe some
// functions are still missed.
// TODO: bring up a better method to get All the functions
func GetMemFuncs(callerFunc *ssa.Function) []*ssa.Function {
	memFuncs := make([]*ssa.Function, 0)

	for _, BB := range callerFunc.Blocks {
		for _, II := range BB.Instrs {
			switch II.(type) {
			case *ssa.Call:
				if callInst, ok := II.(*ssa.Call); ok {
					callValue := callInst.Common().Value
					if funcValue, ok := callValue.(*ssa.Function); ok {
						memFuncs = append(memFuncs, funcValue)
					}
				}
			}
		}
	}

	return memFuncs
}

// apply only to ```go func(){}```, TODO: other func call not considered
// FreeVars in AnonFunc ==> Bindings in CallerFunc
// Find LOAD/STORE in AnonFunc referring to FreeVars
// Find LOAD/STORE in CallerFunc referring to Bindings
// GO instruction must dominate STORE in CallerFunc (TODO: Apply only to Figure 5 in the paper, the other situation not considered)
// For a shared var, if LOAD or STORE in AnonFunc, and both LOAD and STORE(after GO) in AnonFunc, then output (TODO: too strict conditions)
func CheckCallerFunc(callerFunc *ssa.Function) []Results {

	var results []Results

	for _, anonFunc := range callerFunc.AnonFuncs {
		// find go func(){}
		if makeClosure, instGo, isCalledByGo := GetMakeClosureCalledByGo(anonFunc); isCalledByGo == true {
			// FreeVars in AnonFunc ==> Bindings in CallerFunc
			if anon2caller := GetMapAnon2Caller(anonFunc.FreeVars, makeClosure.Bindings); anon2caller != nil {
				// Find LOAD/STORE in AnonFunc referring to FreeVars
				// Find LOAD/STORE in CallerFunc referring to Bindings
				anonRecords := RecordVarsInAnonFunc(anonFunc.FreeVars)
				callerRecords := RecordVarsInCallerFunc(makeClosure.Bindings)

				result := make(map[*ssa.FreeVar] ResultInfo)

				for freeVar, value := range anon2caller {

					var resultInfo ResultInfo

					resultInfo.AnonFunc = anonFunc
					resultInfo.AnonReferrer = anonRecords[freeVar]
					resultInfo.CallerVar = value
					resultInfo.CallerFunc = callerFunc
					resultInfo.CallerReferrer = callerRecords[value]

					freeVarValue := fmt.Sprintf("%v", freeVar)

					// Note: Testing should not be considered, chan type should not be considered either.
					// Too many situations to be exhausted.
					// TODO: expand to regex, including testing.B, chan type ...
					if freeVarValue == "freevar t : **testing.T" {
						continue
					}

					// GO instruction must dominate STORE in CallerFunc
					storeInsts := make([]*ssa.Store, 0)
					for _, storeInst := range resultInfo.CallerReferrer.StoreInsts {
						if InstrDominates(instGo, storeInst) {
							storeInsts = append(storeInsts, storeInst)
						}
					}
					resultInfo.CallerReferrer.StoreInsts = storeInsts

					// if LOAD or STORE in AnonFunc, and both LOAD and STORE(after GO) in AnonFunc
					if len(resultInfo.AnonReferrer.StoreInsts) + len(resultInfo.AnonReferrer.LoadInsts) > 0 &&
						len(resultInfo.CallerReferrer.StoreInsts) * len(resultInfo.CallerReferrer.LoadInsts) > 0 {
						result[freeVar] = resultInfo
					}
				}

				results = append(results, result)
			}

		}

	}

	return results
}




//func CheckCallerFunc(anonFunc *ssa.Function) {
//
//	for _, referrer := range *anonFunc.Referrers() {
//
//		isCalledByGo := false
//
//		if makeClosure := tryParseMakeClosure(&referrer); makeClosure != nil {
//
//			for _, mReferrer := range *makeClosure.Referrers() {
//				if instGo := tryParseGo(&mReferrer); instGo != nil {
//					isCalledByGo = true
//					break
//				}
//			}
//
//			if isCalledByGo {
//
//				if len(anonFunc.FreeVars) != len(makeClosure.Bindings) {
//					fmt.Println(anonFunc, "len(anonFunc.FreeVars) != len(makeClosure.Bindings)")
//					continue
//				}
//
//				for _, binding := range makeClosure.Bindings {
//					for _, referrer := range *binding.Referrers() {
//
//						if !sharedVarInfo.IsCalledByLoadInCallerFuncs {
//							if load := tryParseLoad(&referrer); load != nil {
//								sharedVarInfo.IsCalledByLoadInCallerFuncs = true
//							}
//						}
//
//						if !sharedVarInfo.IsCalledByStoreInCallerFuncs {
//							if store := tryParseStore(&referrer); store != nil {
//								sharedVarInfo.IsCalledByLoadInCallerFuncs = true
//							}
//						}
//					}
//				}
//			}
//		}
//	}
//}
//
//func CheckFunction2(f *ssa.Function) {
//
//	for _, anonFunc := range f.AnonFuncs {
//
//		var sharedVarInfos []SharedVarInfo
//
//		isCalledByGo := false
//
//		if isCalledByGo {
//
//			for _, freeVar := range anonFunc.FreeVars {
//				for _, referrer := range *freeVar.Referrers() {
//					if !sharedVarInfo.IsCalledByLoadInAnonFuncs {
//						if load := tryParseLoad(&referrer); load != nil {
//							sharedVarInfo.IsCalledByLoadInAnonFuncs = true
//						}
//					}
//
//					if !sharedVarInfo.IsCalledByStoreInAnonFuncs {
//						if store := tryParseStore(&referrer); store != nil {
//							sharedVarInfo.IsCalledByStoreInAnonFuncs = true
//						}
//					}
//				}
//
//			}
//		}
//
//		if (sharedVarInfo.IsCalledByLoadInCallerFuncs || sharedVarInfo.IsCalledByStoreInCallerFuncs) &&
//			(sharedVarInfo.IsCalledByLoadInAnonFuncs || sharedVarInfo.IsCalledByStoreInAnonFuncs) {
//			sharedVarInfos = append(sharedVarInfos, sharedVarInfo)
//		}
//	}
//
//	//return sharedVarInfos
//}


// First get the function lists (unique), then check them.
func Check(mainPkg *ssa.Package) {

	var funcList []*ssa.Function

	for name, member := range mainPkg.Members {
		switch f := member.(type) {
		case *ssa.Function:
			if name == "init" {
				break
			}

			funcList = append(funcList, f)
			for _, funcVar := range GetMemFuncs(f) {
				funcList = append(funcList, funcVar)
			}
		}
	}

	var uniqFunclist []*ssa.Function

	encountered := make(map[*ssa.Function] bool)
	for _, funcVar := range funcList {
		if !encountered[funcVar] {
			encountered[funcVar] = true
			uniqFunclist = append(uniqFunclist, funcVar)
		}
	}

	for _, funcVar := range uniqFunclist {

		//PrintCFG(f)
		//PrintRecursive(f)

		results := CheckCallerFunc(funcVar)
		for _, result := range results {
			fmt.Println(result)
		}
	}

}

// Conversion functions
func tryParseLoad(inst *ssa.Instruction) *ssa.UnOp {
	switch (*inst).(type) {
	case *ssa.UnOp:
		unOp := (*inst).(*ssa.UnOp)
		if unOp.Op == token.MUL { // t0 = *x
			return unOp
		}
	}
	return nil
}

func tryParseStore(inst *ssa.Instruction) *ssa.Store {
	switch (*inst).(type) {
	case *ssa.Store:
		store := (*inst).(*ssa.Store)
		return store
	}
	return nil
}

func tryParseMakeClosure(inst *ssa.Instruction) *ssa.MakeClosure {
	switch (*inst).(type) {
	case *ssa.MakeClosure:
		makeClosure := (*inst).(*ssa.MakeClosure)
		return makeClosure
	}
	return nil
}

func tryParseGo(inst *ssa.Instruction) *ssa.Go {
	switch (*inst).(type) {
	case *ssa.Go:
		instGo := (*inst).(*ssa.Go)
		return instGo
	}
	return nil
}