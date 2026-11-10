# colt

**Column selection that's never silently wrong.**

Pick and filter table columns by name. If colt can't tell where a column
starts or ends, it stops and tells you where. It never guesses.

```
$ kubectl get pods -A -o wide | colt -s namespace,name,restarts,node -w 'status!=Running'
NAMESPACE   NAME                  RESTARTS      NODE
default     api-7d9f8c6b5-zt8wn   5 (32s ago)   worker-2
payments    ledger-0              0             worker-2
payments    ledger-1              0             <none>
```

> **Status:** early. Built so far: header parsing and row splitting
> (`internals/`), the flag and subcommand skeleton, and `help` / `version`.
> `colt` doesn't select or filter anything yet: it currently only prints its
> input with line numbers. Everything else below is the target behaviour.
> Sample inputs live in [`testdata/`](testdata).

## Why

awk numbers fields by splitting on whitespace. Command-line tables aren't
whitespace-separated: cells like `5 (32s ago)`, headers like `NOMINATED NODE`,
and empty `MESSAGE` cells all shift the field numbers, on some rows only.

The obvious awk for the query above:

```
$ kubectl get pods -A -o wide | awk '$4 != "Running" {print $1, $2, $5, $8}'
NAMESPACE NAME RESTARTS NODE
default api-7d9f8c6b5-zt8wn 5 3d
payments ledger-0 0 worker-2
payments ledger-1 0 <none>
```

The crash-looping pod is reported on node `3d`. There's no error: the
`(32s ago)` pushed every later field one place right, on that row only.
Making it correct takes a regex to strip the restart detail, header handling,
and `column -t`, and dropping `-A` changes every field number again.

colt finds where each column starts and ends, and splits every row at those
boundaries. Columns are addressed by name, so adding, removing or reordering
columns doesn't break a command. When the boundaries are unclear, colt fails
with an error instead of printing a plausible wrong answer.

## Usage

```
colt [-s columns] [-w condition]... [--no-headers] [file...]
colt --explain [file...]
colt help | version | learn
```

Reads stdin when no file is given. The first line must be the header.

| Subcommand | Meaning |
|---|---|
| `help` | Print the usage and flags (same as `-h`). |
| `version` | Print the version. |
| `learn` | Not implemented yet; exits with an error. |

| Flag | Meaning |
|---|---|
| `-s`, `--select name,...` | Columns to print, in this order. Default: all. |
| `-w`, `--where condition` | Keep only rows matching the condition. Repeat to AND conditions. |
| `--no-headers` | Don't print the header row. |
| `--explain` | Show how colt split the table instead of printing it. |
| `-h`, `--help` | Flags, naming rules, `-w` syntax and examples. |

### Select columns

```
$ kubectl get pods | colt -s name,status
NAME      STATUS
busybox   Running
nginx     Running
```

### Filter rows

```
$ kubectl get pods -A | colt -w 'status!=Running'
$ kubectl get pods -A | colt -w 'namespace~^kube' -w 'restarts>0'
```

### Pipe the values elsewhere

```
$ kubectl get pods | colt -s name -w 'status=Completed' --no-headers | xargs kubectl delete pod
```

### Chain

Output keeps its header, so colt can read colt:

```
$ kubectl get pods -A | colt -w 'status!=Running' | colt -s name
```

## Never silently wrong

Finding column boundaries is the hard part, and the part colt refuses to
guess at.

- **Detection.** Boundaries come from the header's column positions, checked
  against the gaps that run down every row. Header positions alone break on
  right-aligned columns (the numbers in `df -h` and `ps aux` start left of
  their header), so the rows get a vote.
- **Failing loudly.** If the header and the rows disagree about a boundary,
  colt exits with an error naming the column and the line where it couldn't
  decide, and suggests `--explain`. It doesn't print anything it isn't sure of.
- **`--explain`.** Prints the header, the boundaries colt found, and the
  column names it derived, and marks any row that doesn't fit:

  ```
  $ kubectl get events | colt --explain
  LAST SEEN   TYPE     REASON                    OBJECT                      MESSAGE
  ^           ^        ^                         ^                           ^
  last_seen   type     reason                    object                      message
  ```

  The exact format isn't final.
- **Typos fail too.** An unknown column name in `-s` or `-w` is an error that
  lists the available columns. It never quietly matches nothing.

## Column names

- Taken from the header row, lowercased, with runs of spaces turned into `_`:
  `NOMINATED NODE` becomes `nominated_node`.
- Matched case-insensitively: `-s Name,STATUS` works.

## `-w` conditions

`column` `operator` `value`, with optional spaces around the operator.
Everything after the operator is the value, spaces included.

| Operator | Meaning |
|---|---|
| `=` / `!=` | Exact, case-sensitive string equality / inequality |
| `~` | Matches regular expression (Go RE2 syntax, unanchored) |
| `>` `<` `>=` `<=` | Numeric comparison |

Numeric comparisons use the cell's leading number, so `5 (32s ago)` is `5`.
Units are ignored: `5m` and `5m3s` both compare as `5`. A cell with no leading
number never matches a numeric comparison.

Multiple `-w` flags must all match (AND). There is no OR yet.

## Output

- Header and aligned columns by default, like the input.
- Column widths are copied from the input, so each row is printed as soon as
  it's read, with no buffering.
- `--no-headers` drops the header row, for `xargs` and similar.

## Input

Any aligned table with a header row: `kubectl get` (including `-o wide`, `-A`,
`--show-labels`, `-o custom-columns`), `kubectl top`, `docker ps`, `df -h`,
`ps aux`, and similar.

Out of scope: `-o json` / `-o yaml` (use `jq` / `yq`), `describe`, and input
without a header row.

## Not yet

- **Unit-aware comparisons** (`age<5m`, `memory>100Mi`). `m` means minutes in
  AGE and millicores in CPU, and there's no reliable way to tell which.
- **Multiple tables in one input** (`kubectl get pods,svc`).
- **Streaming re-detection.** `kubectl get --watch` may format later rows with
  different widths than the first header.
- **awk programs**, likely by passing tab-separated output to the system's
  `awk` with column names translated to `$N`.
- **Other output formats** (`-o tsv`, `-o json`).
- **Man page.** `colt --help` is the reference for now.

## Known limits

Column boundaries currently come from the header line alone: every run of
non-space characters is a column, and each row is cut at those start offsets.
That is right for left-aligned tables with single-word headers (`kubectl get
pods`) and silently wrong for two other kinds. There is no overlap check yet, so
nothing reports it.

- **Multi-word headers.** `NOMINATED NODE` becomes two columns, `NOMINATED` and
  `NODE`, so the wide pods table ends up with two columns named `NODE`.
  Seen in `get_pods_all_namespaces_wide.txt`, `docker_ps.txt` and `df_h.txt`.
- **Right-aligned columns.** A value can start left of its header. On
  `ps_eo_high_pids.txt`, the 6-digit PID `382356` is cut into `382` (in the
  `USER` cell) and `356  38` (in the `PID` cell).
- **Headers with no gap between them** (`ss -tulpn`: `Peer Address:PortProcess`).

## Build

```
go build -o colt ./cmd/colt
```

## Tests

Test inputs and expected outputs come before parsing code.

`testdata/` holds real captures: `kubectl` output from a kind cluster
(`kubectl get pods` in its various output formats, plus the restart-cell and
nodes captures), and local tools with differently shaped tables (`df -h`,
`ps`, `ss`, `lsblk`, `free -h`, `docker ps`, `docker images`).

So far the tests are unit tests next to the code (`internals/`, `cmd/colt/`)
using short inline inputs. The plan is for each end-to-end test to run colt on a
capture and compare its output with a stored expected-output (golden) file. AGE,
IP and event timestamps are point-in-time values: fine to compare inside a fixed
capture, but don't assert on them against a live cluster.
