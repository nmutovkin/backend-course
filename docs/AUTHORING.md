# Authoring Guide

Modules are written **one at a time, in order**. Each one must agree with
what came before it and set up what comes after. Two documents make that
possible:

- [`LEDGER.md`](./LEDGER.md) is the course's memory: what has been built,
  which terms are defined, which decisions are fixed, and which promises
  ("covered in Module N") are still open.
- This guide is the procedure for writing a module.

## Procedure for module N

### Before writing

1. Read `LEDGER.md` in full.
2. Read module N's entry in `SYLLABUS.md` (topics, depth tags, labs) and
   milestone M*N* in `PROJECT.md` (acceptance criteria).
3. List every **open promise** in the ledger that targets module N. Each one
   must be fulfilled in module N, or moved to a later module in the ledger,
   with a reason.
4. Read the "Delivered" section and the code of the previous milestone. New
   code builds on that code; it doesn't replace it silently.
5. Skim the syllabus entries for modules N+1 to N+3. Anything they will need
   (a package, a type, a term) should be introduced in a form they can build on.

### While writing

Follow the module template below and the rules after it.

### After writing

1. `cd gatepass && gofmt -l . && go vet ./... && go test -race ./...` passes,
   with no files listed by `gofmt`.
2. Update `LEDGER.md`:
   - module status and a short summary,
   - new terms, with the module and section that define them,
   - promises fulfilled (close them) and new promises made (open them),
   - code state: packages, commands and key types now in `gatepass/`,
   - decisions and conventions introduced,
   - any deviation from `SYLLABUS.md` or `PROJECT.md`.
3. If the module deviates from the plan, update `SYLLABUS.md`,
   `PROJECT.md` and `README.md` in the **same commit** so the plan always
   matches reality.
4. Commit as `Module NN: <title>`.

## Module template

File: `modules/NN-slug/README.md`.

```
# Module NN: Title

> **Depth:** which sub-sections are at L3, L2 and L1
> **You will build:** one sentence
> **Time:** ~N hours

## 1. Why this exists       the problem, told through Gatepass
## 2. Concepts              language-agnostic; depth tag on each sub-section
## 3. Under the hood        protocol bytes, data structures, algorithms (L3 material)
## 4. Build it              walk through the reference code in gatepass/
## 5. Across languages      the same idea in Python, TS/Node, Java/Kotlin, C#, Rust
## 6. Labs                  exercises; at least one per L3 sub-section
## 7. Project milestone     Delivered checklist, acceptance command, extension tasks
## 8. Checkpoint            ~10 questions answerable without notes
## 9. Further reading       short and curated
```

## Writing rules

- **Respect the depth tags.** `L3` means the learner could rebuild it: show
  the bytes, the algorithm, the failure modes. `L1` means a paragraph or a
  table of trade-offs, no more. Don't let `L1` topics grow.
- **Define before use.** A term is used only after it is defined in this
  module or in an earlier one listed in the ledger. When reusing a term, link
  back to where it was defined if it's been a while.
- **Forward references are promises.** Writing "Module 12 covers this" opens
  a promise in the ledger. Only make promises the syllabus supports.
- **One running story.** Use the canonical examples in the ledger (event,
  venue, cities, domain, numbers) instead of inventing new ones.
- **Plain, direct prose.** Second person, short paragraphs, concrete
  examples, tables for comparisons. Explain *why* before *how*.
- **Documentation-safe values.** Domains under `.test` or `example.com`; IPs
  from the documentation ranges (`192.0.2.0/24`, `198.51.100.0/24`,
  `203.0.113.0/24`, `2001:db8::/32`).

## Code rules

- All project code lives in `gatepass/` (one Go module). Lessons link to it
  with relative paths; they don't paste whole files, only the lines that matter.
- **Stdlib first.** A third-party dependency needs a reason, recorded in the
  ledger's dependency table.
- Code must be correct and tested. The reference implementation is what
  learners copy.
- Comments explain *why*, at the same density as the existing code.
- Errors are wrapped with `%w` and say which step failed.
- Tests are table-driven where there are several cases; run with `-race`.
- Lab solutions are not committed. Only milestone code is. "Extension tasks"
  in section 7 are for the learner and stay out of the reference code
  unless a later module needs them (then the ledger says so).
