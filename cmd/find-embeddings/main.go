package main

/* Tim Henderson (tadh@case.edu)
*
* Copyright (c) 2015, Tim Henderson, Case Western Reserve University
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
	"math"
	"io"
	"os"
	"os/signal"
	"syscall"
	"runtime/pprof"
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
	cmd.UsageMessage = "find-embeddings --help"
	cmd.ExtendedMessage = `
find-embeddings -p <pattern> <graph>
`
}

func main() {
	os.Exit(run())
}

func run() int {
	args, optargs, err := getopt.GetOpt(
		os.Args[1:],
		"h:p:",
		[]string{
			"help",
			"pattern=",
			"cpu-profile=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	pattern := ""
	cpuProfile := ""
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		case "-p", "--pattern":
			pattern = oa.Arg()
		case "--cpu-profile":
			cpuProfile = cmd.AssertFile(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}

	if pattern == "" {
		fmt.Fprintf(os.Stderr, "You must supply a pattern (-p)\n")
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	conf := &config.Config{}

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "You must supply exactly an input path\n")
		fmt.Fprintf(os.Stderr, "You gave: %v\n", args)
		cmd.Usage(cmd.ErrorCodes["opts"])
	}
	inputPath := cmd.AssertFileOrDirExists(args[0])
	getInput := func() (io.Reader, func()) {
		return cmd.Input(inputPath)
	}

	loader, err := digraph.NewVegLoader(conf, digraph.MinImgSupported, 0, int(math.MaxInt32), 0, int(math.MaxInt32))
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was error constructing the loader\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	errors.Logf("INFO", "Got configuration about to load dataset")
	dt, err := loader.Load(getInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was error during the loading process\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	graph := dt.(*digraph.Digraph)

	sg, err := subgraph.ParsePretty(pattern, graph.G.Labels)
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was error during the parsing the pattern '%v'\n", pattern)
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	errors.Logf("INFO", "loaded pattern %v", sg.Pretty(graph.G.Colors))

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

	
	seen := make(map[int]bool)
	ei, err := subgraph.FilterAutomorphs(sg.IterEmbeddings(
		graph.Indices,
		func(lcv int, chain []*subgraph.Edge) func(b *subgraph.FillableEmbeddingBuilder) bool {
			return func(b *subgraph.FillableEmbeddingBuilder) bool {
				for _, id := range b.Ids {
					if id < 0 {
						continue
					}
					if _, has := seen[id]; !has {
						return false
					}
				}
				return true
			}
	}))
	if err != nil {
		fmt.Fprintf(os.Stderr, "There was error constructing the embedding iterator\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	total := 0
	for emb, ei := ei(); ei != nil; emb, ei = ei() {
		errors.Logf("INFO", "embedding %v", emb.Pretty(graph.G.Colors))
		for _, id := range emb.Ids {
			seen[id] = true
		}
		total++
	}
	errors.Logf("INFO", "total embeddings %v", total)

	errors.Logf("INFO", "done")
	return 0
}
