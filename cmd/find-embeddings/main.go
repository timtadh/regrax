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
	"math"
	"io"
	"os"
	"strings"
	"strconv"
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


func loadProbabilities(path string) (prs []float64, patterns []string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	prs = make([]float64, 0, 10)
	patterns = make([]string, 0, 10)
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, ",", 2)
		pr, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return nil, nil, err
		}
		pattern := strings.TrimSpace(fields[1])
		prs = append(prs, pr)
		patterns = append(patterns, pattern)
	}
	return prs, patterns, nil
}


func main() {
	os.Exit(run())
}

func run() int {
	args, optargs, err := getopt.GetOpt(
		os.Args[1:],
		"h:p:v:",
		[]string{
			"help",
			"pattern=",
			"cpu-profile=",
			"visualize=",
			"probabilities=",
			"samples=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	visual := ""
	patterns := make([]string, 0, 10)
	prPath := ""
	cpuProfile := ""
	samples := -1
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		case "-p", "--pattern":
			patterns = append(patterns, oa.Arg())
		case "--probabilities":
			prPath = cmd.AssertFileExists(oa.Arg())
		case "--samples":
			samples = cmd.ParseInt(oa.Arg())
		case "--cpu-profile":
			cpuProfile = cmd.AssertFile(oa.Arg())
		case "-v", "--visual":
			visual = cmd.AssertFile(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}

	if prPath != "" && len(patterns) > 0 {
		fmt.Fprintf(os.Stderr, "You cannot supply patterns with both (-p) and (--probabilities)\n")
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	var prs []float64 = nil
	if prPath != "" {
		var err error
		prs, patterns, err = loadProbabilities(prPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "There was error loading the probability file\n")
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
	}

	if len(prs) != 0 && samples == -1 {
		fmt.Fprintf(os.Stderr, "You must supply the total number samples (with replacement) (--samples) when using (--probabilities)\n")
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	if len(patterns) == 0 {
		fmt.Fprintf(os.Stderr, "You must supply a pattern (-p)\n")
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	if len(prs) == 0 {
		prs = make([]float64, 0, len(patterns))
		p := 1.0/float64(len(patterns))
		for _ = range patterns {
			prs = append(prs, p)
		}
	}

	conf := &config.Config{}

	loadDt, args := cmd.ParseType(args, conf)
	dt, _ := loadDt(nil)
	graph := dt.(*digraph.Digraph)

	if cpuProfile != "" {
		defer cmd.CPUProfile(cpuProfile)()
	}

	var visualize io.Writer = nil
	if visual != "" {
		f, err := os.Create(visual)
		if err != nil {
			fmt.Fprintf(os.Stderr, "There was error opening the visualization output file\n")
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		defer f.Close()
		visualize = f
	}

	matches := make([]float64, 0, len(patterns))
	matched := make([]*subgraph.SubGraph, 0, len(patterns))
	total := 0.0
	for i, pattern := range patterns {
		sg, err := subgraph.ParsePretty(pattern, &graph.G.Colors, graph.G.Labels)
		if err != nil {
			fmt.Fprintf(os.Stderr, "There was error during the parsing the pattern '%v'\n", pattern)
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		match, csg, err := sg.EstimateMatch(graph.Indices)
		if err != nil {
			errors.Logf("ERROR", "%v", err)
			return 1
		}
		matches = append(matches, match)
		matched = append(matched, csg)
		fmt.Printf("%v, %v, %v\n", i+1, match, pattern)
		total += match
		if visualize != nil {
			dotty, err := csg.VisualizeEmbedding(sg.AsIndices(graph.Indices))
			if err != nil {
				fmt.Fprintf(os.Stderr, "There was error visualizing the embedding '%v'\n", csg)
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return 1
			}
			fmt.Fprintln(visualize, dotty)
		}
	}
	fmt.Printf(", %v, sample total of frequent pattern coverage\n", total)
	fmt.Printf(", %v, sample mean of frequent pattern coverage\n", total/float64(len(patterns)))
	estMean, estVarMean := estPopMean(samples, prs, matches)
	fmt.Printf(", %v, estimated population mean\n", estMean)
	fmt.Printf(", %v, estimated variance population mean\n", estVarMean)
	fmt.Printf(", %v, estimated standard dev population mean\n", math.Sqrt(estVarMean))
	return 0
}

func estPopMean(n int, prs, ys []float64) (estMu, estVarMu float64) {
	if len(prs) != len(ys) {
		panic("len(prs) != len(ys)")
	}
	pis := samplingPrs(n, prs)
	estN := 0.0
	for _, pi := range pis {
		estN += 1.0/pi
	}
	estT := 0.0
	for i := range pis {
		estT += ys[i]/pis[i]
	}
	estMu = estT/estN
	estVarMu = estVarMean(n, estN, estMu, prs, pis, ys)
	return estMu, estVarMu
}

func samplingPrs(n int, prs []float64) []float64 {
	pis := make([]float64, 0, len(prs))
	for _, pr := range prs {
		pis = append(pis, 1 - math.Pow(1 - pr, float64(n)))
	}
	return pis
}

func jointSamplingPrs(n int, prs, pis []float64) [][]float64 {
	cJpi := func(i, j int) float64 {
		return pis[i] + pis[j] - (1 - math.Pow(1 - prs[i] - prs[j], float64(n)))
	}
	jpis := make([][]float64, 0, len(pis))
	for i := range pis {
		jpi := make([]float64, 0, len(pis))
		for j := range pis {
			x := cJpi(i, j)
			if x == 0.0 {
				x = 1e-17 // fixup in case of extremely small joint pr
			}
			jpi = append(jpi, x)
		}
		jpis = append(jpis, jpi)
	}
	return jpis
}

func estVarMean(n int, estN, estMean float64, prs, pis, ys []float64) float64 {
	jpis := jointSamplingPrs(n, prs, pis)
	a := 0.0
	for _, pi := range pis {
		a += (1 - pi)/(math.Pow(pi, 2))
	}
	b := 0.0
	for i := range pis {
		for j := range pis {
			if i == j {
				continue
			}
			lhs := ((jpis[i][j] - pis[i]*pis[j])/(pis[i]*pis[j]))
			rhs := (((ys[i] - estMean)*(ys[j] - estMean))/(jpis[i][j]))
			b += lhs * rhs
		}
	}
	scale := 1/math.Pow(estN, 2)
	return scale*(a + b)
}
