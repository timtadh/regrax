package main

/* Tim Henderson (tadh@case.edu)
*
* Copyright (c) 2016, Tim Henderson, Case Western Reserve University
* Cleveland, Ohio 44106. All Rights Reserved.
*
* This library is free software; you can redistribute it and/or modify
* it under the terms of the GNU General Public License as published by
* the Free Software Foundation; either version 3 of the License, or (at
* your option) any later version.
*
* This library is distributed in the hope that it will be useful, but
* WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
* General Public License for more details.
*
* You should have received a copy of the GNU General Public License
* along with this library; if not, write to the Free Software
* Foundation, Inc.,
*   51 Franklin Street, Fifth Floor,
*   Boston, MA  02110-1301
*   USA
 */

import (
	"fmt"
	"log"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"runtime/pprof"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/dot"
	"github.com/timtadh/combos"
	"github.com/timtadh/getopt"
)

import (
	"github.com/timtadh/sfp/cmd"
)

func init() {
	cmd.UsageMessage = "clean-go-pprof --help"
	cmd.ExtendedMessage = `
go tool pprof -dot -output <profile.dot> <program> <profile.pprof>
clean-go-pprof -i profile.dot -o profile-clean.dot
`
}

func main() {
	os.Exit(run())
}

func run() int {
	args, optargs, err := getopt.GetOpt(
		os.Args[1:],
		"h:i:o:",
		[]string{
			"input=",
			"output=",
			"cpu-profile=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		cmd.Usage(cmd.ErrorCodes["opts"])
	}
	if len(args) != 0 {
		fmt.Fprintln(os.Stderr, "trailing args: %v", args)
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	cpuProfile := ""
	inputPath := ""
	outputPath := ""
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		case "--cpu-profile":
			cpuProfile = cmd.AssertFile(oa.Arg())
		case "-i", "--input":
			inputPath = cmd.AssertFile(oa.Arg())
		case "-o", "--output":
			outputPath = cmd.AssertFile(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}

	if cpuProfile != "" {
		errors.Logf("DEBUG", "starting cpu profile: %v", cpuProfile)
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Fatal(err)
		}
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig:=<-sigs
			errors.Logf("DEBUG", "closing cpu profile")
			pprof.StopCPUProfile()
			err := f.Close()
			errors.Logf("DEBUG", "closed cpu profile, err: %v", err)
			panic(errors.Errorf("caught signal: %v", sig))
		}()
		defer func() {
			errors.Logf("DEBUG", "closing cpu profile")
			pprof.StopCPUProfile()
			err := f.Close()
			errors.Logf("DEBUG", "closed cpu profile, err: %v", err)
		}()
	}

	var inputf io.ReadCloser
	if inputPath == "" {
		inputPath = "stdin"
		inputf = os.Stdin
	} else {
		var err error
		inputf, err = os.Open(inputPath)
		if err != nil {
			errors.Logf("ERROR", "could not open %v : %v", inputPath, err)
			return 1
		}
	}
	input, err := ioutil.ReadAll(inputf)
	inputf.Close()
	if err != nil {
		errors.Logf("ERROR", "could not read input %v : %v", inputPath, err)
		return 1
	}

	var output io.WriteCloser
	if outputPath == "" {
		inputPath = "stdout"
		output = os.Stdout
	} else {
		var err error
		output, err = os.Create(outputPath)
		if err != nil {
			errors.Logf("ERROR", "could not create output %v : %v", outputPath, err)
			return 1
		}
	}
	defer output.Close()

	errors.Logf("INFO", "cleaning %v writing to %v", inputPath, outputPath)
	err = clean(input, output)
	if err != nil {
		errors.Logf("ERROR", "error cleaning profile %v", err)
		return 1
	}
	return 0
}


func clean(input []byte, output io.Writer) (err error) {
	p := &dotParse{output: output}
	return dot.StreamParse(input, p)
}

type dotParse struct {
	output io.Writer
	subgraph int
}

func (p *dotParse) Enter(name string, n *combos.Node) error {
	if name == "SubGraph" {
		p.subgraph += 1
		return nil
	}
	graphName := n.Get(1).Value.(string)
	fmt.Fprintf(p.output, `digraph "%v" {%v`, p.escapeString(graphName), "\n")
	return nil
}

func (p *dotParse) Stmt(n *combos.Node) error {
	if false {
		errors.Logf("DEBUG", "stmt %v", n)
	}
	if p.subgraph > 0 {
		return nil
	}
	switch n.Label {
	case "Node":
		return p.cleanNode(n)
	case "Edge":
		return p.cleanEdge(n)
	}
	return nil
}

func (p *dotParse) Exit(name string) error {
	if name == "SubGraph" {
		p.subgraph--
		return nil
	}
	_, err := fmt.Fprintln(p.output, "}")
	return err
}

func (p *dotParse) cleanNode(n *combos.Node) (err error) {
	sid := n.Get(0).Value.(string)
	attrs := make(map[string]interface{})
	for _, attr := range n.Get(1).Children {
		name := attr.Get(0).Value.(string)
		value := attr.Get(1).Value.(string)
		attrs[name] = value
	}
	label := sid
	if l, has := attrs["label"]; has {
		label = l.(string)
	}
	attrs["original_label"] = label
	attrs["label"] = p.cleanLabel(label)

	sattrs := make([]string, 0, len(attrs))
	for name, value := range attrs {
		sattrs = append(sattrs, p.attrString(name, value.(string)))
	}

	_, err = fmt.Fprintf(p.output, "	%v [%v];\n", sid, strings.Join(sattrs, ", "))
	return err
}

func (p *dotParse) cleanEdge(n *combos.Node) (err error) {
	srcSid := n.Get(0).Value.(string)
	targSid := n.Get(1).Value.(string)
	_, err = fmt.Fprintf(p.output, "	%v -> %v;\n", srcSid, targSid)
	return err
}

func (p *dotParse) cleanLabel(label string) (string) {
	return strings.SplitN(label, `\n`, 2)[0]
}


func (p *dotParse) attrString(name, value string) (string) {
	name = p.escapeString(name)
	value = p.escapeString(value)
	return fmt.Sprintf(`"%v"="%v"`, name, value)
}

func (p *dotParse) escapeString(s string) string {
	bytes := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i + 1 < len(s) {
			bytes = append(bytes, s[i])
			i++
			bytes = append(bytes, s[i])
		} else if s[i] == '"' {
			bytes = append(bytes, '\\')
			bytes = append(bytes, s[i])
		} else {
			bytes = append(bytes, s[i])
		}
	}
	return string(bytes)
}
