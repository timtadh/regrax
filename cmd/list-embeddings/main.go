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
	"bufio"
	"fmt"
	"os"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/getopt"
)

import (
	"github.com/timtadh/sfp/cmd"
	"github.com/timtadh/sfp/config"
	"github.com/timtadh/sfp/types/digraph"
	"github.com/timtadh/sfp/types/digraph/subgraph"
)

func init() {
	cmd.UsageMessage = "list-embeddings --help"
	cmd.ExtendedMessage = `
list-embeddings -p <pattern> <graph>
`
}


func main() {
	os.Exit(run())
}

func loadNames(path string) (patterns []string, count int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	patterns = make([]string, 0, 10)
	seen := make(map[string]bool)
	count = 0
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		pattern := line
		count++
		if _, has := seen[pattern]; !has {
			seen[pattern] = true
			patterns = append(patterns, pattern)
		}
	}
	return patterns, count, nil
}

func run() int {
	args, optargs, err := getopt.GetOpt(
		os.Args[1:],
		"h:p:n:",
		[]string{
			"help",
			"pattern=",
			"cpu-profile=",
			"names=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	patterns := make([]string, 0, 10)
	namesPath := ""
	cpuProfile := ""
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		case "-p", "--pattern":
			patterns = append(patterns, oa.Arg())
		case "-n", "--names":
			namesPath = cmd.AssertFileExists(oa.Arg())
		case "--cpu-profile":
			cpuProfile = cmd.AssertFile(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}

	if namesPath != "" && len(patterns) > 0 {
		fmt.Fprintf(os.Stderr, "You cannot supply patterns with both (-p) and (-n)\n")
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	if len(patterns) == 0 && namesPath == "" {
		fmt.Fprintf(os.Stderr, "You must supply a pattern (-p, -n)\n")
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	if namesPath != "" {
		var err error
		patterns, _, err = loadNames(namesPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "There was error loading the probability file\n")
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
	}

	conf := &config.Config{}

	graphs := make([]*digraph.Digraph, 0, 10)
	for len(args) > 0 {
		loadDt, as := cmd.ParseType(args, conf)
		args = as
		dt, _ := loadDt(nil)
		graph := dt.(*digraph.Digraph)
		graphs = append(graphs, graph)
	}

	if cpuProfile != "" {
		defer cmd.CPUProfile(cpuProfile)()
	}

	errors.Logf("INFO", "looking for embeddings")
	for _, graph := range graphs {
		for _, pattern := range patterns {
			sg, err := subgraph.ParsePretty(pattern, graph.Labels)
			if err != nil {
				fmt.Fprintf(os.Stderr, "There was error during the parsing the pattern '%v'\n", pattern)
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return 1
			}
			// if sg.Pretty(graph.Labels) != pattern {
			// 	errors.Logf("ERROR", "bad load of pattern")
			// 	errors.Logf("ERROR", "expected %v", pattern)
			// 	errors.Logf("ERROR", "got      %v", sg.Pretty(graph.Labels))
			// 	return 1
			// }
			errors.Logf("INFO", "cur sg: %v", sg.Pretty(graph.Labels))
			ei, _ := sg.IterEmbeddings(subgraph.MostConnected, graph.Indices, nil, nil, nil)
			c := 0
			for _, next := ei(false); next != nil; _, next = next(false) {
				c++
			}
			errors.Logf("EMB", "total embs: %v", c)
			fmt.Println(sg.Dotty(graph.Labels, nil, nil))
		}
	}

	return 0
}
