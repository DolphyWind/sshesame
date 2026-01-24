package main

import (
	"strings"
	"maps"

	"mvdan.cc/sh/v3/syntax"
)

type Environment struct {
	env map[string]string
}

func NewEnv(parent *Environment) *Environment {
	if parent != nil {
		return parent.make_copy()
	}
	return &Environment{env: make(map[string]string)}
}

func (e *Environment) make_copy() *Environment {
	var env map[string]string = make(map[string]string)
	maps.Copy(env, e.env)
	return &Environment{
		env: env,
	}
}

func (e *Environment) get(key string) string {
	val, ok := e.env[key]
	if ok {
		return val
	}
	return ""
}

func (e *Environment) get_raw(key string) (string, bool) {
	val, ok := e.env[key]
	return val, ok
}

func (e *Environment) set(key string, value string) {
	e.env[key] = value
}

func stringify_stmt(stmt *syntax.Stmt) string {
	var sb strings.Builder
	if stmt.Negated {
		sb.WriteString("! ")
	}

	sb.WriteString(stringify_command(stmt.Cmd))

	// Redirection
	for _, r := range stmt.Redirs {
		if r.N != nil {
			sb.WriteString(r.N.Value)
		}
		switch r.Op {
		// Real redirects
		case syntax.RdrIn:
			sb.WriteString("<")
		case syntax.RdrOut:
			sb.WriteString(">")

		// Mental illnesses
		case syntax.AppAll:
			sb.WriteString("&>>")
		case syntax.AppOut:
			sb.WriteString(">>")
		case syntax.ClbOut:
			sb.WriteString(">|")
		case syntax.DashHdoc:
			sb.WriteString("<<-")
		case syntax.DplIn:
			sb.WriteString("<&")
		case syntax.DplOut:
			sb.WriteString(">&")
		case syntax.Hdoc:
			sb.WriteString("<<")
		case syntax.RdrAll:
			sb.WriteString("&>")
		case syntax.RdrInOut:
			sb.WriteString("<>")
		case syntax.WordHdoc:
			sb.WriteString("<<<")
		}
		sb.WriteString(stringify_word(r.Word))
	}

	return sb.String()
}

func stringify_command(c syntax.Command) string {
	var sb strings.Builder

	sb.WriteString("UNIMPLEMENTED")
	switch c.(type) {
	case *syntax.ArithmCmd:
	case *syntax.BinaryCmd:
	case *syntax.Block:
	case *syntax.CallExpr:
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

	return sb.String()
}

func stringify_word(w *syntax.Word) string {
	var content strings.Builder
	for _, part := range w.Parts {
		content.WriteString(stringify_wordpart(part))
	}
	return content.String()
}

func stringify_wordpart(wp syntax.WordPart) string {
	// I don't plan to implement the rest of these as this was meant to be a debug function
	switch w := wp.(type) {
	case *syntax.ArithmExp:
		return "UNIMPLEMENTED"
	case *syntax.BraceExp:
		return "UNIMPLEMENTED"
	case *syntax.CmdSubst:
		return "UNIMPLEMENTED"
	case *syntax.DblQuoted:
		var content strings.Builder
		for _, p := range w.Parts {
			content.WriteString(stringify_wordpart(p))
		}
		return content.String()
	case *syntax.ExtGlob:
		return "UNIMPLEMENTED"
	case *syntax.Lit:
		return w.Value
	case *syntax.ParamExp:
		
	case *syntax.ProcSubst:
		var content strings.Builder
		switch w.Op {
		case syntax.CmdIn:
			content.WriteString("<(")
		case syntax.CmdOut:
			content.WriteString(">(")
		}
		for _, s := range w.Stmts {
			content.WriteString(stringify_stmt(s))
		}
		content.WriteString(")")
		return content.String()
	case *syntax.SglQuoted:
		return w.Value
	}
	return "UNREACHABLE"
}
