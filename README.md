# REGRAX - REcurring GRAph eXtractor

by Tim A. D. Henderson (tadh@case.edu)

Licensed under the GPLv3

This is a in progress research project. Parts of this work are discribed in

> **Tim A. D. Henderson**. [Frequent Subgraph Analysis and its Software
> Engineering
> Applications](http://hackthology.com/frequent-subgraph-analysis-and-its-software-engineering-applications.html).
> [Case Western Reserve University](http://case.edu/).
> Doctoral Dissertation. 2017.

However, this work is a combined platform which contains experiments which are
not fully explained in the above dissertation. If you use REGRAX in any way
please

1. Cite the paper and this repository
2. Contact me via (tadh@case.edu) to understand what bits you are using and if
   they require special consideration during publication.

# What this is

REGRAX is a platform for high performance frequent subgraph analysis of large
connected graphs. It supports both frequent subgraph mining and probablistic
markov based sampling methods for extracting frequent subgraphs. It does not
(and never will) support discriminative or importance mining.

## What is a frequent subgraph

A frequent subgraph is a graph fragment which recurs at least `n` times in
either a large graph or a database of small graphs. When considering computing
frequency or recurrence in large graphs there are subtle considerations as a
fragment overlap itself through automorphism. See my dissertation for a detailed
discussion.

# Install

1. Install go <https://golang.org>

2. Install dep <https://github.com/golang/dep>

3. Set up the directory structure

        $ mkdir -p regrax/src/github.com/timtadh/
        $ cd regrax

3. Create a python virtualenv

        $ virtualenv --no-site-packages env

3. Clone the repo

        $ src/github.com/timtadh
        $ git clone https://github.com/timtadh/regrax

4. Activate the virtualenv and set the go path (you have to do this everytime
   you want to use regrax):

        $ cd src/github.com/timtadh
        $ source .activate

4. Run dep (in the repo) to install the go dependencies:

        $ dep ensure

4. Install the python deps:

        $ pip install -r requirements.txt

5. Compile the regrax

        $ go install github.com/timtadh/regrax

6. Check to make sure it works

        $ regrax
        Support <= 0, must be > 0
        regrax --help
        Try -h or --help for help

