package code

import (
	"go/ast"
)

const (
	literalErr = "err"
	literalNil = "nil"
)

type errType int

const (
	errPropagate errType = iota
	errPanic
	errSuppress
	errOther
)

func classify(ifStmt *ast.IfStmt) errType {
	if len(ifStmt.Body.List) != 1 {
		return errOther
	}

	if len(ifStmt.Body.List) != 1 {
		return errOther
	}

	ret, ok := ifStmt.Body.List[0].(*ast.ReturnStmt)
	if !ok {
		if isPanic(ifStmt.Body.List[0]) {
			return errPanic
		}

		return errOther
	}

	if len(ret.Results) == 0 {
		return errSuppress
	}

	last := ret.Results[len(ret.Results)-1]

	ident, ok := last.(*ast.Ident)
	if ok {
		if ident.Name == literalErr {
			return errPropagate
		}

		return errSuppress
	}

	call, ok := last.(*ast.CallExpr)
	if ok {
		for _, arg := range call.Args {
			ident, ok = arg.(*ast.Ident)
			if ok && ident.Name == literalErr {
				return errPropagate
			}
		}
	}

	return errSuppress
}

func isPanic(stmt ast.Stmt) bool {
	exp, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}

	funcExp, ok := exp.X.(*ast.CallExpr)
	if !ok {
		return false
	}

	ident, ok := funcExp.Fun.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "panic"
}
