# Head Command Compatibility

## Summary

Compatible with GNU `head` for single-stream input in line mode (`-n`); byte mode (`-c`) and multiple file operands diverge by design.

## Key Behaviors

```bash
# Default: first 10 lines of the input
$ printf '1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n' | head
1
2
3
4
5
6
7
8
9
10

# -n N: first N lines (N larger than the input yields the whole input)
$ printf 'a\nb\nc\nd\ne\n' | head -n 3
a
b
c

# Single file operand: matches GNU (no header for a lone file)
$ head -n 2 a.txt
1
2

# -c N: leading N bytes (see divergence on the trailing newline)
$ printf 'hello world\n' | head -c 5
hello
```

## Intentional Divergences

- `-c` (byte mode) appends a trailing newline. The leading N bytes are emitted as a single value, which the framework's `[]byte` sink terminates with a newline, so `printf 'hello world\n' | head -c 5` yields `hello\n` (6 bytes) where GNU yields `hello` (5 bytes, no added newline).
- Multiple file operands are concatenated into one byte stream rather than processed independently. GNU `head` applies `-n`/`-c` to *each* file and prefixes each with a `==> NAME <==` header; this implementation emits the first N lines/bytes of the *concatenation* with no headers. For `a.txt` (1..12) followed by `b.txt` (one, two, three), `head -n 2 a.txt b.txt` yields `1\n2` — `b.txt` never contributes and no headers are printed.
- A single file operand matches GNU exactly (GNU also omits the header when given one file).
