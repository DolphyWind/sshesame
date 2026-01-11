package main

import (
	"fmt"
	"io"

	"mvdan.cc/sh/v3/syntax"
)

type InterpreterOptions struct {
	exit_on_error bool
	print_executed bool
}

type Interpreter struct {
	environment Environment
	options InterpreterOptions
	stdin io.Reader
	stdout io.Writer
}

func NewInterpreter() *Interpreter {
	return &Interpreter{}
}

func (i *Interpreter) execute_file(file *syntax.File) (int, error) {
	exit_code := 0
	for _, stmt := range file.Stmts {
		if i.options.print_executed {
			fmt.Fprintf(i.stdout, "+ ")
			syntax.NewPrinter().Print(i.stdout, stmt)
			fmt.Fprintf(i.stdout, "\n")
		}
		exit_code, _ = i.execute_stmt(stmt)
		if i.options.exit_on_error && exit_code != 0 {
			break
		}
	}

	return exit_code, nil
}

func (i *Interpreter) execute_stmt(stmt *syntax.Stmt) (int, error) {
	if stmt.Cmd == nil {
		return 0, nil
	}
	
	if stmt.Background {
		// TODO: Maybe spawn a goroutine but it feels like overkill to emulate background porcesses
		return 0, nil
	}
	if stmt.Coprocess {
		// Same reasoning as above
		return 0, nil
	}
	// TODO: Implement redirecting output
	//
	// for _, redir := range stmt.Redirs {
	// 	_ = redir
	// }
	status, err := i.execute_command(stmt.Cmd)

	if err != nil {
		return -1, err
	}
	
	if stmt.Negated {
		if status == 0 {
			status = 1
		} else {
			status = 0
		}
	}
	return status, nil
}

func (i *Interpreter) execute_command(c syntax.Command) (int, error) {
	status := 0
	var err error = nil
	switch cmd := c.(type) {
	case *syntax.ArithmCmd:
	case *syntax.BinaryCmd:
	case *syntax.Block:
	case *syntax.CallExpr:
		status, err = i.execute_call(cmd)
	case *syntax.CaseClause:
	case *syntax.CoprocClause:
	case *syntax.DeclClause:
	case *syntax.ForClause:
	case *syntax.FuncDecl:
	case *syntax.IfClause:
	case *syntax.LetClause:
	case *syntax.Subshell:
	case *syntax.TestClause:
	case *syntax.TestDecl:
	case *syntax.TimeClause:
	case *syntax.WhileClause:
	default:
		panic(fmt.Sprintf("unexpected syntax.Command: %#v", c))
	}

	return status, err
}

func (i *Interpreter) execute_call(cmd *syntax.CallExpr) (int, error) {
	// Temporary calls
	for _, w := range cmd.Args {
		fmt.Fprintf(i.stdout, "ARG: %v", stringify_word(w))
		fmt.Fprintf(i.stdout, " ")
	}
	fmt.Fprintf(i.stdout, "\n")

	return 0, nil
}
