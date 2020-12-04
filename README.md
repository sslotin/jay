# Jay

Jay is an experimental command-line tool for preparing programming contest problems and stress-testing solutions.

The end goal of the project is to create a lightweight modern replacement for [polygon](https://polygon.codeforces.com/)
and core parts of [ejudge](https://github.com/blackav/ejudge) that could be used by both contestants and problemsetters.

It is named after [John Jay](https://en.wikipedia.org/wiki/John_Jay) (1745–1829), the first US Chief Justice.

## Install

```bash
go get github.com/sslotin/jay
# you may want to prefix it with "sudo GOBIN=/usr/local/bin/" on Linux to put it in path right away
```

Jay currently relies on a few Unix utilities, so some parts may not work on Windows yet, but it will be made cross-platform soon.

TODO: make it installable with a single `snap`, `brew` and whatever it is that Windows users feel comfortable with.

## Basic Usage

Stress-test a solution: `jay test [solution] [flags...]`

Primary options:

* `-r/--reference`: reference solution used to generate correct answer
* `-g/--generator`: program that writes a single test case to stdout; could leverage SEED environment variable to make test generation reproducible
* `-c/--checker`: program that checks correctness of output; it should take input, output and answer files as argv and return any exit code to denote WA
* `-i/--interactor`: essentially generator+checker in case of interactive problems; if set, everything else will be ignored
* `-t/--tests`: path to directory with tests; could be used together with a generator and/or reference

Jay is language-agnostic; "source string" for each executable can be one of the following:

* Path to source file. Only supports C++ and Python for now: `a.cpp`, `generator.py`
* Path to executable. Any file with an abscent or unknown extension: `run`, `a.out`
* Run command. String that specifies how to compile & run it: `go run d.go`, `node solution.js`
* Compile command. It should contain `@` symbol that will be replaced with tempfile for output binary: `g++ f.cpp --std=c++17 -o @`

If tests directory is specified, every file inside it will be treated as a test case with files with an `.a` extension considered reference answers (you do not always have to have it). You can also parse sample cases from online judges: `jay parse [url] [path]`. The parser is quite stupid: it just looks for `<pre>` tags and copies odd ones as inputs and even ones as oututs. This works for CodeForces, Yandex.Contest and few other judges.

TODO: figure out how to parse NEERC-style PDFs

## Examples

Let's walk through the contents of the `examples` directory.

```
examples/
├── apb
│   ├── generator.py
│   ├── reference.hs
│   └── solution.cpp
├── binsearch
│   ├── interactor.py
│   └── solution.py
└── knapsack
    ├── checker.py
    ├── generator.py
    └── solution.cpp
```

Take the "A + B Problem" as first example. It has:

* Generator in Python that gradually increases input size (by default Jay passes sequential numbers as SEED environment variable so one can use it to implement logic like that)
* Buggy solutin in C++ (with `int` instead of `long long`)
* Correct solution in Haskell (we need to showcase "compile command" somehow)

Here is what testing typically looks like:

```
> jay test solution.cpp -r "ghc reference.hs -o @" -g generator.py
Directory for this run: .jay/484831718
Compiling: ghc reference.hs -o @
Compiling: g++ solution.cpp --std=c++17 -o @

Results:
  1  OK     3ms  solution.cpp
  2  OK     4ms  solution.cpp
  3  OK     2ms  solution.cpp
  4  OK     5ms  solution.cpp
  5  OK     4ms  solution.cpp
  6  OK     4ms  solution.cpp
  7  OK     3ms  solution.cpp
  8  OK     3ms  solution.cpp
  9  OK     2ms  solution.cpp
 10  OK     2ms  solution.cpp
 11  WA     4ms  solution.cpp

--- Test 11 --------------------------------------------------

Input:
1015243460 8559580387

Output:                 Reference:
-1132240189        |    9574823847
```

Hint: during debugging like that it is handy to add `--reload` flag, which will "hot reload" the whole run on any source changes.

### Checkers

You do not always have to specify a checker: by default, the output will be compared against reference byte-by-byte (except for trailing empty lines). You only need to provide one if the answer is not unique.

Checker is basically any program that takes input, output and optionally answer files and exist with a non-zero code (either explicitly of by asserts) if it detects wrong answer. Here is an example for a knapsack-type problem (which doesn't need a reference solution):

```python
import sys

def parse_ints(s: str):
    return list(map(int, s.split()))

with open(sys.argv[1], "r") as input_file:
    lines = input_file.readlines()
    n = int(lines[0])
    a = parse_ints(lines[1])

with open(sys.argv[2], "r") as output_file:
    k = parse_ints(output_file.readlines()[0])
    b = [a[i - 1] for i in k]

assert sum(b) % n == 0
assert len(k) > 0
assert len(set(k)) == len(k)
assert all(1 <= x <= n for x in k)

```

### Interactors

To test intactive problems, you need to add `--interactor` and leave the rest of primary options unspecified. Interactor basically needs to do the job of both generator and checker. Here is an example for binary search:

```python
import os
import sys
import random

random.seed(os.environ['SEED'])

x = random.randint(1, 1000)
print(x, flush=True)

request_count = 0

while True:
    c, y = input().split()
    y = int(y)

    if c == '!':
        assert x == y, 'Wrong answer'
        break
    elif x > y:
        print('>', flush=True)
    elif x < y:
        print('<', flush=True)
    else:
        print("=", flush=True)

    request_count += 1
    assert request_count <= 20, "Too many requests made"
```
