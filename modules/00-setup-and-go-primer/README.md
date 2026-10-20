# Module 00: Setup & Go Primer

> **Depth:** environment setup `L2`; Go language `L2` (just enough to read
> and change the code in Modules 01–02; Module 03 takes Go to `L3`).
>
> **You will build:** a working environment that runs the Gatepass tests,
> and a small Go program of your own.
>
> **Time:** ~4 hours (less if you already know Go).

---

## 1. Why this exists

From Module 01 on, every idea gets built into **Gatepass**, our ticketing
platform, in Go. This module makes sure two things are true before you start:

1. Your machine can build and run the code and has the tools the labs use.
2. You can read and modify Go code: not master it (that's Module 03), just
   follow along without getting stuck on syntax.

If you already write Go, do section 2 and skim section 3.

---

## 2. Setting up your environment `L2`

### 2.1 Choose an operating system

Backends run on Linux, and many labs use Linux-only tools. Options:

| You have          | Do this                                                                 |
|-------------------|-------------------------------------------------------------------------|
| Linux             | Nothing extra. Commands below assume Ubuntu/Debian; Fedora equivalents are given. |
| Windows 10/11     | Install **WSL2** (Windows Subsystem for Linux). Open PowerShell as Administrator and run `wsl --install -d Ubuntu`, restart, then open "Ubuntu" from the Start menu. Do all course work inside it. |
| macOS             | Most labs work natively with Homebrew. A few Linux-only labs (`ss`, `tc`, `strace`) need a Linux VM or container. Each lab says when. To get a Linux shell: install Docker (§2.5) and run `docker run -it --rm --cap-add NET_ADMIN ubuntu:24.04 bash`. |

### 2.2 The terminal, minimum version

Module 02 teaches the shell properly. For now you need:

| Command             | Meaning                                          |
|---------------------|--------------------------------------------------|
| `pwd`               | print the current directory                      |
| `ls`, `ls -la`      | list files (`-la`: include hidden files, details) |
| `cd path`, `cd ..`  | change directory; `..` is the parent             |
| `mkdir -p a/b`      | create directories                               |
| `cat file`          | print a file                                     |
| `command &`         | run a command in the background                  |
| `Ctrl+C`            | stop the running command                         |
| `sudo command`      | run as administrator (asks for your password)    |

In this course, a line starting with `$` is a command you type (without the
`$`). Lines without `$` are output.

### 2.3 Install the toolchain

**Git** (version control; Module 04 teaches it):

```console
$ sudo apt update && sudo apt install -y git        # Ubuntu/Debian/WSL
$ sudo dnf install -y git                           # Fedora
$ xcode-select --install                            # macOS (includes git)
```

**Go 1.24 or newer.** Distribution packages are often older, so use the
official release:

```console
# Linux / WSL (x86-64). For ARM machines use linux-arm64 in the file name.
$ curl -LO https://go.dev/dl/go1.24.7.linux-amd64.tar.gz
$ sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.24.7.linux-amd64.tar.gz
$ echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
$ source ~/.bashrc

# macOS
$ brew install go
```

Check it:

```console
$ go version
go version go1.24.7 linux/amd64
```

`PATH` is the list of directories the shell searches for commands. The
`echo ... >> ~/.bashrc` line adds Go's directories to it for every new
terminal. (On macOS the default shell is zsh, so the file is `~/.zshrc`.
Homebrew sets the path for you.)

**Python 3**, only so we can start a throwaway web server with one command:

```console
$ sudo apt install -y python3        # macOS already has python3
```

**Network and system tools** used by the labs in Modules 01–02:

| Tool        | Used for                         | Ubuntu/Debian/WSL          | Fedora               | macOS                    |
|-------------|----------------------------------|----------------------------|----------------------|--------------------------|
| `curl`      | HTTP client                      | `apt install curl`         | `dnf install curl`   | built in                 |
| `nc`        | raw TCP connections              | `apt install netcat-openbsd` | `dnf install nmap-ncat` | built in             |
| `dig`       | DNS queries                      | `apt install dnsutils`     | `dnf install bind-utils` | built in             |
| `tcpdump`   | packet capture                   | `apt install tcpdump`      | `dnf install tcpdump`| built in                 |
| `ss`, `ip`, `tc` | sockets, interfaces, traffic control | `apt install iproute2` | `dnf install iproute iproute-tc` | Linux only |
| `strace`    | trace system calls               | `apt install strace`       | `dnf install strace` | Linux only (`dtruss` is similar) |
| `lsof`      | list open files and sockets      | `apt install lsof`         | `dnf install lsof`   | built in                 |

All at once on Ubuntu/Debian/WSL:

```console
$ sudo apt install -y curl netcat-openbsd dnsutils tcpdump iproute2 strace lsof
```

Optional: **Wireshark** (`sudo apt install wireshark` or download the macOS
app) shows the same packets as `tcpdump`, in a graphical window.

### 2.4 An editor

Any editor works. If you have no preference, use **VS Code** with the
official **Go** extension. On first opening a `.go` file it offers to install
`gopls` (the Go language server). Accept: it gives you autocompletion,
go-to-definition, and formatting on save. On WSL, install VS Code on Windows
and the "WSL" extension, then run `code .` from the Ubuntu terminal.

### 2.5 Docker (needed from Module 07)

Module 07 runs PostgreSQL in a container, and Module 10 uses containers for
tests. You don't need to understand containers yet (Module 23 explains them
in depth). You only need Docker installed:

- **Linux:** `sudo apt install -y docker.io docker-compose-v2`, then
  `sudo usermod -aG docker $USER`, log out and back in.
- **Windows/WSL and macOS:** install **Docker Desktop**. On Windows, enable
  "Use the WSL 2 based engine" in its settings.

Check with `docker run --rm hello-world`. It should print "Hello from Docker!".

### 2.6 Get the course code and run it

```console
$ git clone https://github.com/nmutovkin/backend-course.git
$ cd backend-course/gatepass
$ go test ./...
ok      github.com/nmutovkin/backend-course/gatepass/cmd/gpcurl                 0.03s
ok      github.com/nmutovkin/backend-course/gatepass/internal/platform/httpwire 0.01s
```

The repository has three parts:

```
backend-course/
├── modules/      the lessons (you are reading one)
├── gatepass/     the project's code, which grows module by module
└── docs/         course-maintenance documents
```

If `go test` printed `ok` lines, your setup works.

---

## 3. Go primer `L2`

This section teaches enough Go to read every line of Modules 01–02. Type the
examples in and run them: reading isn't enough.

### 3.1 A program

Create a directory with a file `main.go`:

```console
$ mkdir -p ~/goplay/hello && cd ~/goplay/hello
$ go mod init hello
```

`go mod init` creates `go.mod`, which says "this directory is a Go **module**
named `hello`". A module is a unit of code you build and version together.
Gatepass is one module, named `github.com/nmutovkin/backend-course/gatepass`.

```go
// main.go
package main

import (
	"fmt"
	"os"
)

func main() {
	name := "Lisbon"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	fmt.Println("Hello,", name)
}
```

```console
$ go run .            # compile and run the package in this directory
Hello, Lisbon
$ go run . Porto
Hello, Porto
$ go build -o hello . && ./hello   # build a binary, then run it
Hello, Lisbon
```

- Every file starts with `package <name>`. The package named `main` with a
  function `main` is a runnable program.
- `import` lists the packages you use. `fmt` (formatting and printing) and
  `os` (operating system access) are in the **standard library**, which ships
  with Go.
- `os.Args` is the list of command-line arguments; `os.Args[0]` is the
  program itself.

### 3.2 Variables and basic types

```go
var count int = 3     // explicit declaration
total := 4500         // short form: type inferred (int). Only inside functions.
const maxSeats = 100  // compile-time constant

var price float64     // declared without a value: gets the "zero value" 0
var label string      // zero value ""
var ok bool           // zero value false
```

| Type               | Examples                         | Notes |
|--------------------|----------------------------------|-------|
| `int`, `int64`, `uint64` | `42`, `-7`                 | `int` is 64 bits on modern machines |
| `float64`          | `3.14`                           | Never use for money (Module 06 explains why) |
| `string`           | `"A12"`, `` `raw\n string` ``    | Immutable bytes, usually UTF-8. Backticks: raw string, no escapes |
| `byte`             | `'A'`                            | Alias for `uint8` |
| `bool`             | `true`, `false`                  | |

Go never converts types implicitly. `int64(x)`, `float64(n)` and `string(b)`
convert explicitly.

### 3.3 Slices, maps and structs

**Slices** are growable lists:

```go
seats := []string{"A12", "A13"}
seats = append(seats, "A14")        // append returns the (possibly new) slice
fmt.Println(len(seats), seats[0])   // 3 A12
fmt.Println(seats[1:3])             // [A13 A14]  (from index 1 up to, not including, 3)
for i, s := range seats {           // range gives index and value
	fmt.Println(i, s)
}
buf := make([]byte, 1024)           // a slice of 1024 zero bytes
```

**Maps** are hash tables:

```go
prices := map[string]int{"A12": 4500, "A13": 4500}
prices["B1"] = 3000
p, found := prices["Z9"]   // p = 0, found = false: missing keys give the zero value
delete(prices, "B1")
```

**Structs** group fields. A name starting with an **upper-case letter** is
exported (visible to other packages); lower-case is private to its package.
That's Go's only visibility rule.

```go
type Seat struct {
	Row    string
	Number int
	held   bool // unexported: only code in this package can see it
}

s := Seat{Row: "A", Number: 12}
s.Number = 13
```

### 3.4 Functions, multiple results and errors

Functions can return several values. Go has no exceptions for ordinary
failures. A function that can fail returns an `error` as its last result,
and the caller checks it immediately:

```go
import (
	"errors"
	"fmt"
	"strconv"
)

var ErrSoldOut = errors.New("sold out") // a "sentinel" error value others can compare against

func parseQuantity(s string) (int, error) {
	n, err := strconv.Atoi(s) // string to int
	if err != nil {
		return 0, fmt.Errorf("parsing quantity %q: %w", s, err)
	}
	if n > 10 {
		return 0, ErrSoldOut
	}
	return n, nil // nil means "no error"
}

func main() {
	n, err := parseQuantity("abc")
	if err != nil {
		fmt.Println("error:", err) // error: parsing quantity "abc": strconv.Atoi: parsing "abc": invalid syntax
		return
	}
	fmt.Println(n)
}
```

- `fmt.Errorf` builds an error message. `%q` prints a quoted string, `%v` any
  value, `%d` an integer, `%s` a string.
- **`%w` wraps** the original error inside the new one. That adds context
  ("which step failed") while keeping the original error inspectable.
- `errors.Is(err, ErrSoldOut)` is true if `err` is `ErrSoldOut` *or wraps
  it*, at any depth. The course's tests use this a lot.
- `_` discards a value you don't need: `n, _ := strconv.Atoi("5")`.

### 3.5 Pointers and methods

A **pointer** holds the address of a value. `&x` takes the address; `*p`
follows it. Pass a pointer when a function must modify the caller's value,
or to avoid copying a large struct:

```go
func hold(s *Seat) { s.held = true } // s.held is shorthand for (*s).held

seat := Seat{Row: "A", Number: 12}
hold(&seat)       // seat.held is now true
```

A **method** is a function attached to a type. The **receiver** goes before
the name:

```go
func (s Seat) Label() string { return fmt.Sprintf("%s%d", s.Row, s.Number) } // value receiver: gets a copy
func (s *Seat) Release()     { s.held = false }                               // pointer receiver: can modify

fmt.Println(seat.Label()) // A12
seat.Release()            // Go takes &seat automatically
```

Rule of thumb: if any method needs a pointer receiver, give them all pointer
receivers.

### 3.6 Interfaces, and the two most important ones

An **interface** is a set of method signatures. Any type with those methods
satisfies it automatically. There is no `implements` keyword.

The standard library's most important interfaces are in package `io`:

```go
type Reader interface {
	Read(p []byte) (n int, err error) // fill p with up to len(p) bytes; return how many
}
type Writer interface {
	Write(p []byte) (n int, err error) // write the bytes in p
}
```

Files, network connections, HTTP bodies, compressors and in-memory buffers
are all `Reader`s and `Writer`s, so code written against the interface works
with all of them. `Read` returns `io.EOF` when there is no more data. That
isn't a failure, just the end.

```go
import ("io"; "os"; "strings")

r := strings.NewReader("hello from Riverside Hall\n") // a Reader over a string
io.Copy(os.Stdout, r)                                 // copy until EOF; os.Stdout is a Writer
```

You can write your own Reader by wrapping another one. This is how Module 01
measures when the first byte arrives:

```go
type countingReader struct {
	r     io.Reader
	total int
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p) // delegate to the wrapped reader
	c.total += n          // ...and observe what passed through
	return n, err
}
```

`bufio.Reader` wraps a Reader with a buffer, adding methods like
`ReadString('\n')` that read up to a delimiter. Reading byte-by-byte from a
network connection without a buffer would cost a system call per byte.

### 3.7 `defer`

`defer` schedules a call to run when the surrounding function returns, however
it returns. Use it for cleanup right after acquiring something:

```go
f, err := os.Open("seats.csv")
if err != nil {
	return err
}
defer f.Close() // runs when the function returns, even on an early return
```

### 3.8 Packages and the project layout

A **package** is a directory of `.go` files that share a `package` name. You
import a package by its module path plus directory:

```go
import "github.com/nmutovkin/backend-course/gatepass/internal/platform/httpwire"

httpwire.ReadResponse(...) // use exported names with the package name as prefix
```

A directory named `internal` is special: packages under it can only be
imported by code inside the parent of `internal`. Gatepass keeps almost
everything under `gatepass/internal/` so nothing outside the project can
depend on it. Runnable programs live in `gatepass/cmd/<name>/`, one directory
per program.

### 3.9 `context.Context`, briefly

Backend code constantly needs to say "give up after 5 seconds" or "the client
went away, stop working". Go passes that as a `context.Context`, by
convention the first parameter of any function that does I/O:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()                     // release resources when done
addrs, err := resolver.LookupNetIP(ctx, "ip", "example.com") // gives up when ctx expires
```

Module 03 explains contexts properly, along with goroutines (`go f()`, Go's
lightweight threads) and channels. Modules 01–02 only use contexts as above.

### 3.10 Tests

Tests live next to the code in files ending in `_test.go`. A test is a
function `TestXxx(t *testing.T)`:

```go
// label_test.go
package main

import "testing"

func TestLabel(t *testing.T) {
	tests := []struct { // a "table" of cases: an anonymous struct type
		name string
		seat Seat
		want string
	}{
		{"simple", Seat{Row: "A", Number: 12}, "A12"},
		{"two-letter row", Seat{Row: "AA", Number: 1}, "AA1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { // each case runs as a named sub-test
			if got := tt.seat.Label(); got != tt.want {
				t.Errorf("Label() = %q, want %q", got, tt.want) // marks failure, keeps going
			}
		})
	}
}
```

`t.Errorf` reports a failure and continues; `t.Fatalf` reports and stops the
test. This **table-driven** style is used throughout the course.

### 3.11 The tools you'll run every day

| Command                     | What it does                                                        |
|-----------------------------|---------------------------------------------------------------------|
| `go run ./cmd/gpcurl`       | compile and run a program                                           |
| `go build ./...`            | compile everything (`./...` means "this directory and all below")   |
| `go test ./...`             | run all tests                                                       |
| `go test -race ./...`       | run tests with the **race detector**, which catches unsafe concurrent memory access (Module 03 explains races) |
| `go test -run TestLabel -v` | run only matching tests, with verbose output                        |
| `go vet ./...`              | static checks for common mistakes                                   |
| `gofmt -w .`                | format code the one standard way (your editor does this on save)     |
| `go doc io.Reader`          | show documentation from the terminal                                |

Every milestone in this course is "done" when this passes:

```console
$ cd gatepass && gofmt -l . && go vet ./... && go test -race ./...
```

(`gofmt -l` lists files that are *not* formatted; it should print nothing.)

---

## 4. Build it

There is no Gatepass milestone in this module. Instead, write a small
program that uses everything from section 3, so the code in Module 01 feels
familiar.

**Task: `seatcount`.** Read lines like `A12 held` or `A13 free` from standard
input and print how many seats are in each state, then any lines it couldn't
parse.

```console
$ printf 'A12 held\nA13 free\nA14 held\nbroken\n' | go run .
free 1
held 2
error: line 4: expected "<seat> <state>", got "broken"
```

Try it yourself before looking at the solution.

<details>
<summary>Solution</summary>

```go
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

var errFormat = errors.New(`expected "<seat> <state>"`)

func parseLine(line string) (seat, state string, err error) {
	fields := strings.Fields(line) // split on whitespace
	if len(fields) != 2 {
		return "", "", fmt.Errorf("%w, got %q", errFormat, line)
	}
	return fields[0], fields[1], nil
}

func count(r io.Reader) (map[string]int, []error) {
	counts := map[string]int{}
	var errs []error
	sc := bufio.NewScanner(r) // reads line by line
	for n := 1; sc.Scan(); n++ {
		_, state, err := parseLine(sc.Text())
		if err != nil {
			errs = append(errs, fmt.Errorf("line %d: %w", n, err))
			continue
		}
		counts[state]++
	}
	return counts, errs
}

func main() {
	counts, errs := count(os.Stdin)
	states := make([]string, 0, len(counts))
	for s := range counts {
		states = append(states, s)
	}
	sort.Strings(states) // map iteration order is random; sort for stable output
	for _, s := range states {
		fmt.Println(s, counts[s])
	}
	for _, err := range errs {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
}
```

Notice: `count` takes an `io.Reader`, not `os.Stdin`, so a test can pass
`strings.NewReader("...")`. Errors are wrapped with the line number and stay
checkable with `errors.Is(err, errFormat)`.

</details>

Then write `count_test.go` with a table-driven test for `count`, and run it
with `go test -v`.

---

## 5. Across languages

If you know another language, this map makes the Go code readable:

| Go                         | Python                    | TypeScript                | Java/Kotlin                 | C#                         | Rust                        |
|----------------------------|---------------------------|---------------------------|-----------------------------|----------------------------|-----------------------------|
| `x, err := f()`            | `try/except`              | `try/catch`, rejected promise | `try/catch`             | `try/catch`                | `let x = f()?;` (`Result`)  |
| `[]T` slice                | `list`                    | `Array<T>`                | `ArrayList<T>` / `List<T>`  | `List<T>`                  | `Vec<T>`                    |
| `map[K]V`                  | `dict`                    | `Map<K,V>`                | `HashMap<K,V>`              | `Dictionary<K,V>`          | `HashMap<K,V>`              |
| `struct` + methods         | `class` / `@dataclass`    | `class` / `interface`     | `class` / `data class`      | `class` / `record`         | `struct` + `impl`           |
| implicit interfaces        | duck typing / `Protocol`  | structural types          | explicit `implements`       | explicit `: IFoo`          | explicit `impl Trait for`   |
| `defer`                    | `with` / `finally`        | `finally`                 | try-with-resources / `use`  | `using`                    | `Drop`                      |
| `io.Reader`                | file-like objects         | `ReadableStream`          | `InputStream`               | `Stream`                   | `std::io::Read`             |
| `context.Context`          | `asyncio` timeouts        | `AbortSignal`             | (varies) / coroutine scope  | `CancellationToken`        | (varies) / `tokio` timeouts |
| `go test`                  | `pytest`                  | `vitest` / `jest`         | JUnit                       | xUnit                      | `cargo test`                |

---

## 6. Labs

1. **Break the build on purpose.** In `~/goplay/hello`, add an unused import
   and run `go build`. Then declare an unused variable. Go refuses to compile
   both. Remove them again.
   *What you should see:* `"strings" imported and not used` and
   `declared and not used: x`. Go treats these as errors to keep code clean.

2. **Wrap and inspect errors.** Write a function that returns
   `fmt.Errorf("reserving A12: %w", ErrSoldOut)`. In `main`, print
   `errors.Is(err, ErrSoldOut)`. Then change `%w` to `%v` and print again.
   *What you should see:* `true`, then `false`. `%v` only copies the text; `%w`
   keeps the original error inside.

3. **Your own `io.Writer`.** Write a type `upperWriter` whose `Write` method
   converts bytes to upper case (`bytes.ToUpper`) and writes them to
   `os.Stdout`. Use it with `io.Copy(upperWriter{}, strings.NewReader("riverside hall"))`.
   *What you should see:* `RIVERSIDE HALL`.

4. **Explore the course code.** In `gatepass/`, run `go doc ./internal/platform/httpwire`
   and `go test -v ./internal/platform/httpwire`. Find the test named
   `TestReadResponse/chunked_with_extension_and_trailer` in the output.
   You'll understand every line of that package by the end of Module 01.

---

## 7. Milestone

No Gatepass milestone. You're done when:

- [ ] `go version` shows 1.24 or newer.
- [ ] `go test ./...` in `gatepass/` prints `ok` for every package.
- [ ] `tcpdump --version`, `dig -v` and `nc -h` all run (macOS: `ss`, `tc`
      and `strace` can wait for a Linux VM or container).
- [ ] `seatcount` works and has a passing table-driven test.
- [ ] (Before Module 07) `docker run --rm hello-world` works.

---

## 8. Checkpoint

<details>
<summary>1. What does <code>go.mod</code> define, and what is the Gatepass module's name?</summary>

It marks a directory as the root of a Go module and gives the module's name
(its import path prefix), plus the Go version and dependencies. Gatepass's is
`github.com/nmutovkin/backend-course/gatepass`.
</details>

<details>
<summary>2. How does Go decide whether a name is visible outside its package?</summary>

By its first letter: upper-case names (`ReadResponse`, `Header`) are
exported; lower-case names (`readLine`) are private to the package.
</details>

<details>
<summary>3. What is the difference between <code>%w</code> and <code>%v</code> in <code>fmt.Errorf</code>?</summary>

Both put the error's text into the message. `%w` also wraps the original
error so `errors.Is` and `errors.As` can still find it.
</details>

<details>
<summary>4. When does a method need a pointer receiver?</summary>

When it must modify the receiver, or when the type is large enough that
copying it on every call is wasteful. If one method needs a pointer receiver,
give them all one, for consistency.
</details>

<details>
<summary>5. What does <code>Read</code> return at the end of a stream, and is it an error?</summary>

It returns `io.EOF`. That's a normal end-of-data signal, not a failure.
Callers like `io.Copy` and `io.ReadAll` treat it as success.
</details>

<details>
<summary>6. Why accept an <code>io.Reader</code> instead of a file or connection?</summary>

Any source of bytes satisfies it (file, socket, string, HTTP body), so the
same function works everywhere and tests can pass `strings.NewReader(...)`.
</details>

<details>
<summary>7. When does a deferred call run?</summary>

When the surrounding function returns, whether normally, through an early
`return`, or by a panic. Deferred calls run in reverse order.
</details>

<details>
<summary>8. What does a package under <code>internal/</code> protect against?</summary>

It can only be imported by code inside the directory that contains
`internal/`, so outside code can't come to depend on it.
</details>

<details>
<summary>9. What does <code>go test -race</code> add?</summary>

It instruments the program to detect data races (unsynchronized concurrent
access to the same memory) and fails the test if one happens.
</details>

---

## 9. Further reading (optional)

Everything you need is above. If you want more Go before Module 03:

- *A Tour of Go* (go.dev/tour): interactive exercises.
- *Effective Go* (go.dev/doc/effective_go): idioms.
- Alan Donovan & Brian Kernighan, *The Go Programming Language*.
