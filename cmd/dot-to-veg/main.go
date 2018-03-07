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
	"io"
	"io/ioutil"
	"encoding/json"
	"compress/gzip"
	"os"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/dot"
	"github.com/timtadh/combos"
	"github.com/timtadh/getopt"
)

import (
	"github.com/timtadh/regrax/cmd"
)

func init() {
	cmd.UsageMessage = "dot-to-veg --help"
	cmd.ExtendedMessage = `
dot-to-veg -i graph.dot -o graph.veg
cat graph.dot | dot-to-veg > out.veg
dot-to-veg -i graph.dot > out.veg
cat graph.dot | dot-to-veg -o graph.veg
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

	inputPath := ""
	outputPath := ""
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		case "-i", "--input":
			inputPath = cmd.AssertFile(oa.Arg())
		case "-o", "--output":
			outputPath = cmd.AssertFile(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}

	var input io.Reader
	if inputPath != "" {
		inputf, err := os.Open(inputPath)
		if err != nil {
			errors.Logf("ERROR", "could not open %v : %v", inputPath, err)
			return 1
		}
		defer inputf.Close()
		input = inputf
	} else {
		inputPath = "<stdin>"
		input = os.Stdin
	}

	var output io.Writer
	if outputPath != "" {
		outputf, err := os.Create(outputPath)
		if err != nil {
			errors.Logf("ERROR", "could not open %v : %v", outputPath, err)
			return 1
		}
		defer outputf.Close()
		if strings.HasSuffix(outputPath, ".gz") {
			z := gzip.NewWriter(outputf)
			defer z.Close()
			output = z
		} else {
			output = outputf
		}
	} else {
		outputPath = "<stdout>"
		output = os.Stdout
	}

	errors.Logf("INFO", "converting %v writing to %v", inputPath, outputPath)
	err = convert(input, output)
	if err != nil {
		errors.Logf("ERROR", "error converting dot to veg %v", err)
		return 1
	}
	return 0
}


func convert(input io.Reader, output io.Writer) (err error) {
	bytes, err := ioutil.ReadAll(input)
	if err != nil {
		return err
	}
	p := &dotParse{
		output: output,
		vertices: make(map[graphVertex]int),
	}
	return dot.StreamParse(bytes, p)
}

type graphVertex struct {
	graph int
	sid string
}

type dotParse struct {
	output io.Writer
	graph, subgraph int
	vertices map[graphVertex]int
	nextVertex int
}

func (p *dotParse) Enter(name string, n *combos.Node) error {
	if name == "Graph" {
		p.graph += 1
	} else if name == "SubGraph" {
		p.subgraph += 1
	}
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
		return p.vertex(n)
	case "Edge":
		return p.edge(n)
	}
	return nil
}

func (p *dotParse) Exit(name string) error {
	if name == "SubGraph" {
		p.subgraph--
		return nil
	}
	return nil
}

func (p *dotParse) vid(sid string) int {
	x := graphVertex{p.graph, sid}
	if vid, has := p.vertices[x]; has {
		return vid
	} else {
		vid := p.nextVertex
		p.nextVertex++
		p.vertices[x] = vid
		return vid
	}
}

func (p *dotParse) vertex(n *combos.Node) (err error) {
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
	attrs["label"] = label
	attrs["id"] = p.vid(sid)

	j, err := json.Marshal(attrs)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(p.output, "vertex	%v\n", string(j))
	return err
}

func (p *dotParse) edge(n *combos.Node) (err error) {
	srcSid := n.Get(0).Value.(string)
	targSid := n.Get(1).Value.(string)
	attrs := make(map[string]interface{})
	for _, attr := range n.Get(2).Children {
		name := attr.Get(0).Value.(string)
		value := attr.Get(1).Value.(string)
		attrs[name] = value
	}
	label := ""
	if l, has := attrs["label"]; has {
		label = l.(string)
	}
	attrs["label"] = label
	attrs["src"] = p.vid(srcSid)
	attrs["targ"] = p.vid(targSid)

	j, err := json.Marshal(attrs)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(p.output, "edge	%v\n", string(j))
	return err
}

