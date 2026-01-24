package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Program interface {
	run(stdin io.Reader, stdout io.Writer, stderr io.Writer, interpreter *Interpreter, args []string) (int, error)
}

type EchoProgram struct {
	Program
}

func (cmd *EchoProgram) run(stdin io.Reader, stdout io.Writer, stderr io.Writer, interpreter *Interpreter, args []string) (int, error) {
	no_newline := false
	escapes := false
	var args_to_print []string = make([]string, 0, len(args))

	for i, s := range args {
		if s == "-n" {
			no_newline = true
		} else if s == "-e" {
			escapes = true
		} else {
			args_to_print = args[i:]
			break
		}
	}
	str_to_print := strings.Join(args_to_print, " ")
	if escapes {
		var err error
		str_to_print, err = strconv.Unquote(str_to_print)
		if err != nil {
			return -1, err
		}
	}
	fmt.Fprintf(stdout, "%s", str_to_print)

	if !no_newline {
		fmt.Fprintf(stdout, "\n")
	}
	return 0, nil
}
