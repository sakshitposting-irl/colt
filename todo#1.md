# todo#1: select and filter kubectl columns by name

## Goal

```
kubectl get po -n ns | colt -s NAME,FAME -w 'Status=Running'
```

- Pick columns by header name (`-s`), filter rows (`-w`).
- `FAME` isn't a column, so this exact command must **fail loudly**:
  ```
  colt: no column "FAME" (have: NAME, READY, STATUS, RESTARTS, AGE)
  ```
- Column names are case-insensitive (`Status` matches `STATUS`).

## Decision

Going table-first. The aligned-table reader and `-s`/`-w` move up from v0.3
to now. `colt learn` comes after.

## Part A (one sprint)

- [ ] **1. Header boundaries.** Each column starts where its header word
      starts. Slice every row at those offsets.
      Test input: `testdata/get_pods.txt`
- [ ] **2. `-s NAME,READY`.** Print only those columns, in the order given.
      Match case-insensitively. Error on unknown names (the `FAME` case).
- [ ] **3. `-w 'STATUS=Running'`.** Support only `=` and `!=` for now,
      as string comparison. Error on unknown column names here too.

Done when this works on `get_pods.txt`:
```
$ colt -s NAME,STATUS -w 'status=Running' < testdata/get_pods.txt
NAME      STATUS
busybox   Running
nginx     Running
```

## Part B (next sprint)

- [ ] **4. Multi-word headers.** `NOMINATED NODE` and `READINESS GATES`
      (in `testdata/get_pods_all_namespaces_wide.txt`). A single space is
      inside a header; two or more spaces separate headers.
- [ ] **5. Fail loudly on overlap.** If a row has non-space text crossing a
      column boundary, error with the line number instead of slicing wrong.
      Catches cells like `5 (32s ago)` in RESTARTS.

## Tests (write these first)

- Golden tests over `testdata/get_pods*.txt`.
- Error cases: unknown `-s` column, unknown `-w` column, bad `-w` syntax,
  empty input.
- Not an error: a header with no data rows prints the header and exits 0.
- Don't assert on AGE or IP values.

## Out of scope for todo#1

`<`, `>`, numbers, units, `$name` awk syntax, JSON output, `--explain`,
`colt learn`.

## Reading

- https://pkg.go.dev/bufio#Scanner
- https://pkg.go.dev/strings
- https://pkg.go.dev/flag (for `-s` and `-w`)
- https://go.dev/wiki/TableDrivenTests

## Which tools hit which problem

Captured from the running `kind-cyclops` cluster and this machine. Files are
in `testdata/`. Nothing was created in the cluster; every capture is read-only.

### Type 1. Delimited (a character marks the boundary)

Not a colt problem; `cut`, `awk -F` or a CSV tool handles it. Control case only.
CSV, TSV, `/etc/passwd`, `docker ps --format '{{.ID}}\t{{.Names}}'`.

### Type 2. Aligned, single-word headers, but cells contain spaces

Splitting on whitespace shifts every later field, on those rows only.

| Tool | Cell with spaces | Capture |
|---|---|---|
| `kubectl get pods -A -o wide` | RESTARTS `44 (26m ago)` | `get_pods_all_namespaces_wide_restarts.txt` |
| `kubectl get nodes -o wide` | OS-IMAGE `Debian GNU/Linux 12 (bookworm)` | `get_nodes_wide.txt` |
| `systemctl list-units` | DESCRIPTION free text; every line starts with 2 spaces | not saved |

Reproduced. The README's awk query on the real capture (Running pods, want
NAMESPACE NAME RESTARTS NODE):

```
$ awk '$4=="Running" && $5!=0 {print $1, $2, $5, $8}' get_pods_all_namespaces_wide_restarts.txt
cert-manager cert-manager-6cb788c8fd-9jbw7 44 8d     <- NODE shows AGE (8d); real node is cyclops-control-plane
```

No error. Header-start slicing handles this type (steps 1-3).

### Type 3. Multi-word headers

One space is inside a name; two or more separate columns. A split on every space
invents extra columns.

| Tool | Multi-word headers | Capture |
|---|---|---|
| `kubectl get pods -o wide` | `NOMINATED NODE`, `READINESS GATES` | `get_pods_all_namespaces_wide.txt` |
| `kubectl get events` | `LAST SEEN` | `get_events.txt` |
| `docker ps` | `CONTAINER ID` | `docker_ps.txt` |
| `docker images` | `DISK USAGE`, `CONTENT SIZE` | `docker_images.txt` |
| `df -h` | `Mounted on` | `df_h.txt` |
| `ss -tulpn` | `Local Address:Port`, `Peer Address:Port` | `ss_tulpn.txt` |

Needs step 4.

### Type 4. Right-aligned columns

Values end under the header's end, so a long value starts left of the header.
Slicing at header starts cuts it, and nothing errors.

| Tool | Right-aligned columns | Capture |
|---|---|---|
| `ps aux`, `ps -ef`, `ps -eo` | PID, PPID, %CPU, VSZ, RSS, TIME | `ps_eo_high_pids.txt` |
| `df -h` | Size, Used, Avail, Use% | `df_h.txt` |
| `lsblk` | MAJ:MIN, SIZE | `lsblk.txt` |
| `docker images` | DISK USAGE, CONTENT SIZE | `docker_images.txt` |
| `free -h` | every number column | `free_h.txt` |

Reproduced. `ps_eo_high_pids.txt` sliced at the header's start offsets:

```
USER         PID    PPID ...
saksham   382356  382349 ...      header starts: USER 0, PID 13, PPID 20

USER = "saksham   382"      <- PID's first digits
PID  = "356  38"            <- PID cut, plus PPID's first digits
```

No error. Needs step 7 (rows vote). Kubectl tables are left-aligned, so none of
them hit this type.

### Also found (not in the 4 types)

- **Headers with no gap between them.** `ss -tulpn`: `Peer Address:PortProcess`.
  `Port` and `Process` touch, so the header alone can't split them (`ss_tulpn.txt`).
- **Row label with no header.** `free -h`: the header has 6 names, rows have 7
  cells (`Mem:`, `Swap:`), and the `Swap:` row has fewer (`free_h.txt`).
- **Prefix characters in the first column.** `lsblk` draws a tree (`├─nvme0n1p1`),
  and `systemctl list-units` indents every line by 2 spaces.

### Mapping to the MVP order

| Type | Step that handles it |
|---|---|
| 2 | 1-3 (header boundaries) and 6 (overlap check) |
| 3 | 4 (multi-word headers) |
| 4 | 7 (rows vote) |
| "also found" | decide per case: fail loudly or support |
