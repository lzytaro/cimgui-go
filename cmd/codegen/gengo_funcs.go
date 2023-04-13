package main

import "C"
import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// returnTypeType represents an arbitrary type of return value of the function.
// for example Known reffers to returnTypeWrappersMap (see below)
type returnTypeType byte

const (
	// default value - will cause the function to be skipped and an error will be printed to stdout
	returnTypeUnknown returnTypeType = iota
	// return type is void (in go - the function returns nothing)
	returnTypeVoid
	// METHOD returns nothing, but it has receiver (called self)
	returnTypeStructSetter
	// Known - reffers to getReturnTypeWrapperFunc
	returnTypeKnown
	// the return value is an enum type (autogenerated by gengo_enums.go)
	returnTypeEnum
	// return type is a pointer to ImGui struct
	returnTypeStructPtr
	// returns ImGui struct
	returnTypeStruct
	// the method is a constructor
	returnTypeConstructor
	// function with first arugment as pointer of return value
	returnTypeNonUDT
)

// generateGoFuncs generates given list of functions and writes them to file
func generateGoFuncs(prefix string, validFuncs []FuncDef, enumNames []string, structNames []string) error {
	generator := &goFuncsGenerator{
		prefix:      prefix,
		structNames: make(map[string]bool),
		enumNames:   make(map[string]bool),
	}

	for _, v := range structNames {
		generator.structNames[v] = true
	}

	for _, v := range enumNames {
		generator.enumNames[v] = true
	}

	generator.writeFuncsFileHeader()

	for _, f := range validFuncs {
		// check whether the function shouldn't be skipped
		if skippedFuncs[f.FuncName] {
			continue
		}

		args, argWrappers := generator.generateFuncArgs(f)

		if len(f.ArgsT) == 0 {
			generator.shouldGenerate = true
		}

		// stop, when the function should not be generated
		if !generator.shouldGenerate {
			fmt.Printf("not generated: %s%s\n", f.FuncName, f.Args)
			continue
		} else {
			fmt.Printf("generated: %s%s\n", f.FuncName, f.Args)
		}

		if noErrors := generator.GenerateFunction(f, args, argWrappers); !noErrors {
			continue
		}
	}

	fmt.Printf("Convert progress: %d/%d\n", generator.convertedFuncCount, len(validFuncs))

	goFile, err := os.Create(fmt.Sprintf("%s_funcs.go", prefix))
	if err != nil {
		panic(err.Error())
	}

	defer goFile.Close()

	_, err = goFile.WriteString(generator.sb.String())
	if err != nil {
		return fmt.Errorf("failed to write content of GO file: %w", err)
	}

	return nil
}

// goFuncsGenerator is an internal state of GO funcs' generator
type goFuncsGenerator struct {
	prefix                 string
	structNames, enumNames map[string]bool

	sb                 strings.Builder
	convertedFuncCount int

	shouldGenerate bool
}

// writeFuncsFileHeader writes a header of the generated file
func (g *goFuncsGenerator) writeFuncsFileHeader() {
	g.sb.WriteString(goPackageHeader)

	g.sb.WriteString(fmt.Sprintf(
		`// #include "extra_types.h"
// #include "%[1]s_structs_accessor.h"
// #include "%[1]s_wrapper.h"
import "C"
import (
	"unsafe"

	"github.com/AllenDang/cimgui-go/internal/wrapper"
)

`, g.prefix))
}

func (g *goFuncsGenerator) GenerateFunction(f FuncDef, args []string, argWrappers []ArgumentWrapperData) (noErrors bool) {
	var returnType, returnStmt, receiver string
	funcName := f.FuncName
	shouldDefer := false

	// determine kind of function:
	returnTypeType := returnTypeUnknown
	rf, err := getReturnTypeWrapperFunc(f.Ret)
	if err == nil {
		returnTypeType = returnTypeKnown
	}

	// attention! order is _probably_ important here so consider that
	// before changing anything here
	goEnumName := f.Ret
	if f.NonUDT == 1 {
		returnTypeType = returnTypeNonUDT
	} else if f.Ret == "void" {
		if f.StructSetter {
			returnTypeType = returnTypeStructSetter
		} else {
			returnTypeType = returnTypeVoid
		}
	} else if g.enumNames[goEnumName] {
		returnTypeType = returnTypeEnum
	} else if strings.HasSuffix(f.Ret, "*") && (g.structNames[strings.TrimSuffix(f.Ret, "*")] || g.structNames[strings.TrimSuffix(strings.TrimPrefix(f.Ret, "const "), "*")]) {
		returnTypeType = returnTypeStructPtr
	} else if f.StructGetter && g.structNames[f.Ret] {
		returnTypeType = returnTypeStruct
	} else if f.Constructor {
		returnTypeType = returnTypeConstructor
	}

	// determine function name, return type (and return statement)
	switch returnTypeType {
	case returnTypeVoid:
		// noop
	case returnTypeNonUDT:
		outArg := argWrappers[0]
		returnType = strings.TrimPrefix(outArg.ArgType, "*")

		returnStmt = fmt.Sprintf("return *%s", f.ArgsT[0].Name)

		argWrappers[0].ArgDef = fmt.Sprintf(`%s := new(%s)
%s
		`, f.ArgsT[0].Name, returnType, outArg.ArgDef)
		args = args[1:]
	case returnTypeStructSetter:
		funcParts := strings.Split(f.FuncName, "_")
		funcName = strings.TrimPrefix(f.FuncName, funcParts[0]+"_")
		if len(funcName) == 0 || !strings.HasPrefix(funcName, "Set") || skippedStructs[funcParts[0]] {
			return false
		}

		receiver = funcParts[0]
	case returnTypeKnown:
		shouldDefer = true
		returnType = rf.returnType
		returnStmt = rf.returnStmt
	case returnTypeEnum:
		shouldDefer = true
		returnType = goEnumName
	case returnTypeStructPtr:
		// return Im struct ptr
		shouldDefer = true
		returnType = strings.TrimPrefix(f.Ret, "const ")
		returnType = strings.TrimSuffix(returnType, "*")
	case returnTypeStruct:
		shouldDefer = true
		returnType = f.Ret
	case returnTypeConstructor:
		shouldDefer = true
		parts := strings.Split(f.FuncName, "_")

		returnType = parts[0]

		if !g.structNames[returnType] {
			return false
		}

		suffix := ""
		if len(parts) > 2 {
			suffix = strings.Join(parts[2:], "")
		}

		funcName = "New" + returnType + suffix
	default:
		fmt.Printf("Unknown return type \"%s\" in function %s\n", f.Ret, f.FuncName)
		return false
	}

	g.sb.WriteString(g.generateFuncDeclarationStmt(receiver, funcName, args, returnType, f))
	argInvokeStmt, declarations, finishers := g.generateFuncBody(argWrappers)
	g.sb.WriteString(strings.Join(declarations, "\n"))
	if len(declarations) > 0 {
		g.sb.WriteString("\n")
	}

	if shouldDefer {
		g.writeFinishers(shouldDefer, finishers)
	}

	// write non-return function calls (finalizers called normally)
	switch returnTypeType {
	case returnTypeVoid, returnTypeNonUDT:
		g.sb.WriteString(fmt.Sprintf("C.%s(%s)\n", f.CWrapperFuncName, argInvokeStmt))
	case returnTypeStructSetter:
		g.sb.WriteString(fmt.Sprintf("C.%s(self.handle(), %s)\n", f.CWrapperFuncName, argInvokeStmt))
	}

	if !shouldDefer {
		g.writeFinishers(shouldDefer, finishers)
	}

	switch returnTypeType {
	case returnTypeNonUDT:
		g.sb.WriteString(fmt.Sprintf("%s", returnStmt))
	case returnTypeKnown:
		g.sb.WriteString(fmt.Sprintf(returnStmt, fmt.Sprintf("C.%s(%s)", f.CWrapperFuncName, argInvokeStmt)))
	case returnTypeEnum:
		g.sb.WriteString(fmt.Sprintf("return %s(C.%s(%s))", renameGoIdentifier(returnType), f.CWrapperFuncName, argInvokeStmt))
	case returnTypeStructPtr:
		g.sb.WriteString(fmt.Sprintf("return (%s)(unsafe.Pointer(C.%s(%s)))", renameGoIdentifier(returnType), f.CWrapperFuncName, argInvokeStmt))
	case returnTypeStruct:
		g.sb.WriteString(fmt.Sprintf("return new%sFromC(C.%s(%s))", renameGoIdentifier(f.Ret), f.CWrapperFuncName, argInvokeStmt))
	case returnTypeConstructor:
		g.sb.WriteString(fmt.Sprintf("return (%s)(unsafe.Pointer(C.%s(%s)))", renameGoIdentifier(returnType), f.CWrapperFuncName, argInvokeStmt))
	}

	g.sb.WriteString("}\n\n")
	g.convertedFuncCount += 1

	return true
}

// this method is responsible for createing a function declaration statement.
// it takes function name, list of arguments and return type and returns go statement.
// e.g.: func (self *ImGuiType) FuncName(arg1 type1, arg2 type2) returnType {
func (g *goFuncsGenerator) generateFuncDeclarationStmt(receiver string, funcName string, args []string, returnType string, f FuncDef) (functionDeclaration string) {
	funcParts := strings.Split(funcName, "_")
	typeName := funcParts[0]

	// Generate default param value hint
	var commentSb strings.Builder
	if len(f.Defaults) > 0 {
		commentSb.WriteString("// %s parameter default value hint:\n")

		// sort lexicographically for determenistic generation
		type defaultParam struct {
			name  string
			value string
		}
		defaults := make([]defaultParam, 0, len(f.Defaults))
		for n, v := range f.Defaults {
			defaults = append(defaults, defaultParam{name: n, value: v})
		}
		sort.Slice(defaults, func(i, j int) bool {
			return defaults[i].name < defaults[j].name
		})

		for _, p := range defaults {
			commentSb.WriteString(fmt.Sprintf("// %s: %s\n", p.name, p.value))
		}
	}

	// convert func(self *receiverType) into a method
	if len(funcParts) > 1 &&
		len(args) > 0 &&
		strings.Contains(args[0], "self ") {

		funcName = strings.TrimPrefix(funcName, typeName+"_")
		receiver = strings.TrimPrefix(args[0], "self ")
		args = args[1:]
	}

	if len(receiver) > 0 {
		receiver = fmt.Sprintf("(self %s)", renameGoIdentifier(receiver))
	}

	funcName = renameGoIdentifier(funcName)

	// if file comes from imgui_internal.h,prefix Internal is added.
	// ref: https://github.com/AllenDang/cimgui-go/pull/118
	if strings.Contains(f.Location, "imgui_internal") {
		funcName = "Internal" + funcName
	}

	return fmt.Sprintf("%sfunc %s %s(%s) %s {\n",
		strings.Replace(commentSb.String(), "%s", renameGoIdentifier(funcName), 1),
		renameGoIdentifier(receiver),
		funcName,
		strings.Join(args, ","),
		renameGoIdentifier(returnType))
}

func (g *goFuncsGenerator) generateFuncArgs(f FuncDef) (args []string, argWrappers []ArgumentWrapperData) {
	for i, a := range f.ArgsT {
		g.shouldGenerate = false

		if a.Name == "type" {
			a.Name = "typeArg"
		}

		if i == 0 && f.StructSetter {
			g.shouldGenerate = true
		}

		if f.StructGetter && g.structNames[a.Type] {
			args = append(args, fmt.Sprintf("%s %s", a.Name, renameGoIdentifier(a.Type)))
			argWrappers = append(argWrappers, ArgumentWrapperData{
				VarName: fmt.Sprintf("%s.handle()", a.Name),
			})

			g.shouldGenerate = true

			continue
		}

		if v, err := argWrapper(a.Type); err == nil {
			arg := v(a)
			argWrappers = append(argWrappers, arg)

			args = append(args, fmt.Sprintf("%s %s", a.Name, renameGoIdentifier(arg.ArgType)))

			g.shouldGenerate = true
			continue
		}

		if goEnumName := a.Type; g.isEnum(goEnumName) {
			args = append(args, fmt.Sprintf("%s %s", a.Name, renameGoIdentifier(goEnumName)))
			argWrappers = append(argWrappers, ArgumentWrapperData{
				VarName: fmt.Sprintf("C.%s(%s)", a.Type, a.Name),
			})

			g.shouldGenerate = true
			continue
		}

		if strings.HasSuffix(a.Type, "*") {
			pureType := strings.TrimPrefix(a.Type, "const ")
			pureType = strings.TrimSuffix(pureType, "*")

			if g.structNames[pureType] {
				args = append(args, fmt.Sprintf("%s %s", a.Name, renameGoIdentifier(pureType)))
				argWrappers = append(argWrappers, ArgumentWrapperData{
					VarName: fmt.Sprintf("%s.handle()", a.Name),
					ArgType: renameGoIdentifier(pureType),
				})

				g.shouldGenerate = true
				continue
			}
		}

		if !g.shouldGenerate {
			fmt.Printf("Unknown argument type \"%s\" in function %s\n", a.Type, f.FuncName)
			break
		}
	}

	return args, argWrappers
}

// Generate function body
// and returns function call arguments
// e.g.:
// it will write the following into the buffer:
func (g *goFuncsGenerator) generateFuncBody(argWrappers []ArgumentWrapperData) (invokeStatement string, declarations, finishers []string) {
	var invokeStmt []string
	declarations, finishers = make([]string, 0, len(argWrappers)), make([]string, 0, len(argWrappers))

	for _, aw := range argWrappers {
		invokeStmt = append(invokeStmt, aw.VarName)
		if len(aw.ArgDef) > 0 {
			declarations = append(declarations, aw.ArgDef)
			if aw.Finalizer != "" {
				finishers = append(finishers, aw.Finalizer)
			}
		}
	}

	return strings.Join(invokeStmt, ","), declarations, finishers
}

func (g *goFuncsGenerator) writeFinishers(shouldDefer bool, finishers []string) {
	if len(finishers) == 0 {
		return
	}

	g.sb.WriteString("\n")

	if shouldDefer {
		g.sb.WriteString("defer func() {\n")
		defer g.sb.WriteString("\n}()\n")
	}

	g.sb.WriteString(strings.Join(finishers, "\n"))
	g.sb.WriteString("\n\n")
}

// isEnum returns true when given string is a valid enum type.
func (g *goFuncsGenerator) isEnum(argType string) bool {
	for en := range g.enumNames {
		if argType == en {
			return true
		}
	}

	return false
}
