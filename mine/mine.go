package mine

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
	"io"
	"os"
	"runtime"
	"strings"
)

import (
	"github.com/timtadh/data-structures/errors"
	"github.com/timtadh/getopt"
)

import (
	"github.com/timtadh/regrax/cmd"
	"github.com/timtadh/regrax/config"
	"github.com/timtadh/regrax/mine/miners/dfs"
	"github.com/timtadh/regrax/mine/miners/index_speed"
	"github.com/timtadh/regrax/mine/miners/qsplor"
	"github.com/timtadh/regrax/mine/miners/vsigram"
	"github.com/timtadh/regrax/sample/miners"
)

func vsigramMode(argv []string, conf *config.Config) (miners.Miner, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hc",
		[]string{
			"help",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		cmd.Usage(cmd.ErrorCodes["opts"])
	}
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}
	return vsigram.NewMiner(conf), args
}

func dfsMode(argv []string, conf *config.Config) (miners.Miner, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"h",
		[]string{
			"help",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		cmd.Usage(cmd.ErrorCodes["opts"])
	}
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}
	return dfs.NewMiner(conf), args
}

func indexSpeedMode(argv []string, conf *config.Config) (miners.Miner, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"h",
		[]string{
			"help",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		cmd.Usage(cmd.ErrorCodes["opts"])
	}
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}
	return index_speed.NewMiner(conf), args
}

func qsplorMode(argv []string, conf *config.Config) (miners.Miner, []string) {
	args, optargs, err := getopt.GetOpt(
		argv,
		"hs:m:",
		[]string{
			"help",
			"score-function=",
			"max-queue-size=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		cmd.Usage(cmd.ErrorCodes["opts"])
	}
	var scorer qsplor.Scorer = qsplor.Scorers["random"]
	var maxQueueSize int = 10
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		case "-m", "--max-queue-size":
			maxQueueSize = cmd.ParseInt(oa.Arg())
		case "-s", "--score-function":
			if _, has := qsplor.Scorers[oa.Arg()]; !has {
				fmt.Fprintf(os.Stderr, "Unknown score function: %v\n", oa.Arg())
				fmt.Fprintf(os.Stderr, "Valid score functions:\n")
				for name, _ := range qsplor.Scorers {
					fmt.Fprintf(os.Stderr, "%v\n", name)
				}
				cmd.Usage(cmd.ErrorCodes["opts"])
			}
			scorer = qsplor.Scorers[oa.Arg()]
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}
	return qsplor.NewMiner(conf, scorer, maxQueueSize), args
}

func Run(argv []string) int {
	modes := map[string]cmd.Mode{
		"dfs":         dfsMode,
		"index-speed": indexSpeedMode,
		"vsigram":     vsigramMode,
		"qsplor":      qsplorMode,
	}

	args, optargs, err := getopt.GetOpt(
		argv,
		"ho:c:p:",
		[]string{
			"help",
			"output=", "cache=",
			"support=",
			"modes", "types", "reporters",
			"skip-log=",
			"cpu-profile=",
			"parallelism=",
		},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "could not process your arguments (perhaps you forgot a mode?) try:")
		fmt.Fprintf(os.Stderr, "$ %v breadth %v\n", os.Args[0], strings.Join(os.Args[1:], " "))
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	output := ""
	cache := ""
	support := 0
	cpuProfile := ""
	parallelism := -1
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-h", "--help":
			cmd.Usage(0)
		case "-o", "--output":
			output = cmd.EmptyDir(oa.Arg())
		case "-c", "--cache":
			cache = cmd.EmptyDir(oa.Arg())
		case "-p", "--parallelism":
			parallelism = cmd.ParseInt(oa.Arg())
		case "--support":
			support = cmd.ParseInt(oa.Arg())
		case "--types":
			fmt.Fprintln(os.Stderr, "Types:")
			for k := range cmd.Types {
				fmt.Fprintln(os.Stderr, "  ", k)
			}
			os.Exit(0)
		case "--modes":
			fmt.Fprintln(os.Stderr, "Modes:")
			for k := range modes {
				fmt.Fprintln(os.Stderr, "  ", k)
			}
			os.Exit(0)
		case "--reporters":
			fmt.Fprintln(os.Stderr, "Reporters:")
			for k := range cmd.Reporters {
				fmt.Fprintln(os.Stderr, "  ", k)
			}
			os.Exit(0)
		case "--skip-log":
			level := oa.Arg()
			errors.Logf("INFO", "not logging level %v", level)
			errors.SkipLogging[level] = true
		case "--cpu-profile":
			cpuProfile = cmd.AssertFile(oa.Arg())
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag '%v'\n", oa.Opt())
			cmd.Usage(cmd.ErrorCodes["opts"])
		}
	}

	if support <= 0 {
		fmt.Fprintf(os.Stderr, "Support <= 0, must be > 0\n")
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	if output == "" {
		fmt.Fprintf(os.Stderr, "You must supply an output dir (-o)\n")
		cmd.Usage(cmd.ErrorCodes["opts"])
	}

	if cpuProfile != "" {
		defer cmd.CPUProfile(cpuProfile)()
	}

	conf := &config.Config{
		Cache:       cache,
		Output:      output,
		Support:     support,
		Parallelism: parallelism,
	}

	return cmd.Main(args, conf, modes)
}

var profileDone chan bool

func profileWriter(w io.Writer) {
	for {
		data := runtime.CPUProfile()
		if data == nil {
			break
		}
		errors.Logf("DEBUG", "profileWriter got data %v", len(data))
		w.Write(data)
	}
	profileDone <- true
}

func profileUnsafe(w io.Writer, hz int) {
	errors.Logf("DEBUG", "profileUnsafe at %v hz", hz)
	profileDone = make(chan bool)
	runtime.SetCPUProfileRate(hz)
	go profileWriter(w)
}

func stopProfile() {
	runtime.SetCPUProfileRate(0)
	<-profileDone
}
