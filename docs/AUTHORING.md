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
   with no files listed by `gofmt`. Code samples in the lesson that aren't
   in `gatepass/` (primer snippets, lab solutions) were compiled and run once.
2. **Self-containment pass:** read the module as a learner with no internet.
   Every lab and question can be completed and checked from the course
   alone (see the writing rules).
3. Update `LEDGER.md`:
   - module status and a short summary,
   - new terms, with the module and section that define them,
   - promises fulfilled (close them) and new promises made (open them),
   - code state: packages, commands and key types now in `gatepass/`,
   - decisions and conventions introduced,
   - any deviation from `SYLLABUS.md` or `PROJECT.md`.
4. If the module deviates from the plan, update `SYLLABUS.md`,
   `PROJECT.md` and `README.md` in the **same commit** so the plan always
   matches reality.
5. Commit as `Module NN: <title>`.

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
## 6. Labs                  "Lab setup" block, then step-by-step labs with "What you
                            should see" and answered questions; at least one per L3 sub-section
## 7. Project milestone     Delivered checklist, acceptance command, extension tasks
                            (with the knowledge they need included)
## 8. Checkpoint            ~10 questions, each with a <details> answer pointing to its section
## 9. Further reading (optional)   short and curated; never required
```

## Writing rules

- **Self-contained.** A reader must be able to finish the module, including
  every lab, checkpoint and extension task, using only the course. External
  sources may be listed in *Further reading (optional)*, never required.
  In particular:
  - every concept, protocol detail, format or algorithm a lab needs is
    explained in the module (or an earlier one listed in the ledger);
  - every tool is installed in Module 00 §2.3 (add new tools to that table)
    or in a "Lab setup" block, and every flag used is explained;
  - every lab says **what you should see**, using real output captured by
    running it (trim it, and say when values will differ). If a lab can't be
    run while authoring, give the expected result worked out from theory
    and say so;
  - every question in a lab or checkpoint has an answer in a collapsible
    `<details>` block;
  - language features beyond what's already taught (Module 00 primer, then
    Module 03) are explained where first used.
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
- Lab solutions are not committed as code files. Only milestone code is.
  Short solutions and answers go in `<details>` blocks in the lesson. "Extension tasks"
  in section 7 are for the learner and stay out of the reference code
  unless a later module needs them (then the ledger says so).
