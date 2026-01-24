package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

type InterpreterOptions struct {
	exit_on_error bool
	print_executed bool
}

type Interpreter struct {
	environment Environment
	programs map[string]Program
	options InterpreterOptions
	stdin io.Reader
	stdout io.Writer
	stderr io.Writer
}

func NewInterpreter(programs map[string]Program) *Interpreter {
	return &Interpreter{
		programs: programs,
		stdin: os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
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
	}

	return status, err
}

func (i *Interpreter) execute_call(cmd *syntax.CallExpr) (int, error) {
	args := make([]string, 0, len(cmd.Args))
	for _, w := range cmd.Args {
		// fmt.Fprintf(i.stdout, "ARG: %v", stringify_word(w))
		// fmt.Fprintf(i.stdout, " ")
		arg, err := i.run_word(w)
		if err != nil {
			return -1, err
		}
		args = append(args, arg)
	}
	// fmt.Fprintf(i.stdout, "\n")
	original_env := i.environment.make_copy()
	for _, assign := range cmd.Assigns {
		val, err := i.run_word(assign.Value)
		if err != nil {
			return -1, err
		}
		i.environment.set(assign.Name.Value, val)
	}
	if len(args) == 0 {
		return -1, fmt.Errorf("Invalid")
	}
	program, ok := i.programs[args[0]]
	if !ok {
		return -1, fmt.Errorf("bash: %s: command not found", args[0])
	}
	code, err := program.run(i.stdin, i.stdout, i.stderr, i, args[1:])
	i.environment = *original_env
	return code, err
}

func (i *Interpreter) run_word(word *syntax.Word) (string, error) {
	parts_str := make([]string, 0, len(word.Parts))
	for _, wp := range word.Parts {
		wp_str, err := i.run_word_part(wp)
		if err != nil {
			return "", err
		}
		parts_str = append(parts_str, wp_str)
	}
	// TODO: Unsure about this part. Revise it later
	return strings.Join(parts_str, ""), nil
}

func (i *Interpreter) run_word_part(wp syntax.WordPart) (string, error) {
	switch p := wp.(type) {
	case *syntax.ArithmExp:
	case *syntax.BraceExp:
	case *syntax.CmdSubst:
	case *syntax.DblQuoted:
		var sb strings.Builder
		for _, ip := range p.Parts {
			s, err := i.run_word_part(ip)
			if err != nil {
				return "", err
			}
			sb.WriteString(s)
		}
		return sb.String(), nil
	case *syntax.ExtGlob:
	case *syntax.Lit:
		return p.Value, nil
	case *syntax.ParamExp:
		var_name := p.Param.Value
		if p.Slice != nil {
			value := i.environment.get(var_name)
			offset_str, err_offset := i.run_arithm_expr(p.Slice.Offset)
			if err_offset != nil {
				return "", err_offset
			}
			length_str, err_len := i.run_arithm_expr(p.Slice.Length)
			if err_len != nil {
				return "", err_len
			}
			offset, err_off_atoi := strconv.Atoi(offset_str)
			if err_off_atoi != nil {
				return "", err_off_atoi
			}
			length, err_len_atoi := strconv.Atoi(length_str)
			if err_len_atoi != nil {
				return "", err_len_atoi
			}
			return value[offset:length], nil
		} else if p.Repl != nil {
			value := i.environment.get(var_name)
			orig, err := i.run_word(p.Repl.Orig)
			if err != nil {
				return "", err
			}
			with, err := i.run_word(p.Repl.With)
			if err != nil {
				return "", err
			}
			if p.Repl.All {
				value = strings.ReplaceAll(value, orig, with)
			} else {
				value = strings.Replace(value, orig, with, 1)
			}
			return value, nil
		} else if p.Length {
			value := i.environment.get(var_name)
			return strconv.Itoa(len(value)), nil
		} else if p.Excl {
			value, ok := i.environment.get_raw(var_name)
			if !ok {
				return "", fmt.Errorf("bash: %s: invalid indirect expansion", var_name)
			}
			actual_value := i.environment.get(value)
			return actual_value, nil
		} else if p.Width {

		} else if p.Exp != nil {

		} else if p.Index != nil {
			
		}
	case *syntax.ProcSubst:
	case *syntax.SglQuoted:
		return p.Value, nil
	}
	return "", nil
}

func (i *Interpreter) run_arithm_expr(ae syntax.ArithmExpr) (string, error) {
	switch expr := ae.(type) {
	case *syntax.BinaryArithm:
	case *syntax.ParenArithm:
		return i.run_arithm_expr(expr.X)
	case *syntax.UnaryArithm:
		val, err := i.run_arithm_expr(expr.X)
		if err != nil {
			return "", err
		}
		val_int, err2 := strconv.Atoi(val)
		if err2 != nil {
			return "", err2
		}
		switch expr.Op {
		case syntax.BitNegation:
			return strconv.Itoa(^val_int), nil
		case syntax.Dec:
			if expr.Post {
				i.environment.set(val, strconv.Itoa(val_int - 1))
				return strconv.Itoa(val_int), nil
			} else {
				i.environment.set(val, strconv.Itoa(val_int - 1))
				return strconv.Itoa(val_int - 1), nil
			}
		case syntax.Inc:
			if expr.Post {
				i.environment.set(val, strconv.Itoa(val_int + 1))
				return strconv.Itoa(val_int), nil
			} else {
				i.environment.set(val, strconv.Itoa(val_int + 1))
				return strconv.Itoa(val_int + 1), nil
			}
		case syntax.Minus:
			return strconv.Itoa(-val_int), nil
		case syntax.Not:
			if val_int == 0 {
				return strconv.Itoa(1), nil
			} else {
				return strconv.Itoa(0), nil
			}
		case syntax.Plus:
			return strconv.Itoa(+val_int), nil
	}
	case *syntax.Word:
		return i.run_word(expr)
	}

	panic("Unreachable")
}
