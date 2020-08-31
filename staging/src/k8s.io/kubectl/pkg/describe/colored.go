package describe

import (
	"fmt"
	"io"
	"strings"

	"github.com/juju/ansiterm"
)

type coloredWriter struct {
	out    *ansiterm.Writer
	scheme []kvStyle
}

type kvStyle struct {
	key   *ansiterm.Context
	value *ansiterm.Context
	title *ansiterm.Context
}

var defaultContext = ansiterm.Foreground(ansiterm.Default)

func NewColoredWriter(out io.Writer, scheme []kvStyle) PrefixWriter {
	writer := ansiterm.NewWriter(out)
	writer.SetColorCapable(true)
	return &coloredWriter{out: writer, scheme: scheme}
}

func (cw *coloredWriter) Write(level int, format string, a ...interface{}) {
	levelSpace := "  "
	prefix := ""
	for i := 0; i < level; i++ {
		prefix += levelSpace
	}
	str := fmt.Sprintf(format, a...)
	st := cw.style(level)
	defaultContext.Fprint(cw.out, prefix)
	if strings.HasSuffix(str, ":\n") { // if ends with ":" consider as title
		// some titles have space padding - we're removing it to avoid incorrect styling
		spacePadding := 0
		str = strings.TrimLeftFunc(str, func(r rune) bool {
			if r == ' ' {
				spacePadding++
				return true
			}
			return false
		})
		defaultContext.Fprint(cw.out, strings.Repeat(" ", spacePadding))
		st.title.Fprintf(cw.out, str)
	} else {
		delimiters := []string{":\t", ": ", "\t", "="}
		found := false
		for _, delim := range delimiters {
			if index := strings.Index(str, delim); index != -1 {
				found = true
				index = index + len(delim)
				st.key.Fprintf(cw.out, str[:index])
				st.value.Fprintf(cw.out, str[index:])
				break
			}
		}
		if !found {
			st.value.Fprintf(cw.out, str)
		}
	}
}

func (cw *coloredWriter) WriteLine(a ...interface{}) {
	st := cw.style(0)
	st.value.Fprint(cw.out, fmt.Sprintln(a...))
}

func (cw *coloredWriter) style(level int) kvStyle {
	if len(cw.scheme) == 0 {
		return kvStyle{ansiterm.Foreground(ansiterm.Default), ansiterm.Foreground(ansiterm.Default), ansiterm.Foreground(ansiterm.Default).SetStyle(ansiterm.Underline)}
	}
	return cw.scheme[level%len(cw.scheme)]
}

func (cw *coloredWriter) Flush() {
	if f, ok := cw.out.Writer.(flusher); ok {
		f.Flush()
	}
}
