# AGENTS.md

Guidance for working in this repo. Compact by design — omit anything already obvious from filenames or `go doc`.

## Verification

Use `task` for all verification; do not substitute raw `go test`/`go vet` for the wrappers below.

- `task fmt` — gofumpt the whole tree
- `task lint` — golangci-lint (config: `.golangci.yaml`, v2)
- `task build`
- `task test` — vet + unit + e2e + raw + cleanup (CI parity)
- `task test:unit` — `-short -race -coverpkg=./...`; tune via `TEST_OPTIONS`, `SOURCE_FILES`, `TEST_PATTERN`
    - Focused:
      `task test:unit SOURCE_FILES=./file_format/envelope/... TEST_PATTERN=TestGzipCompressorRejectsOversizedOutput`
- `task test:unit:full` — same without `-short`
- `task ci` (setup + build + test), `task test:fuzz`, `task cover`, `task reltest`
- E2E requires `pdftoppm` (macOS: `brew install poppler`); `task test` flows a PDF through pdftoppm → `scan` → `decode`.

Pre-commit hook: `task dev` installs `.git/hooks/pre-commit` (runs `gofumpt` + `golangci-lint run --new --fix`). Not
installed by default.

## Style

- Comments: `why` only, never restate what code does. The revive config deliberately drops `exported`/`package-comments`
  doc rules so "what" comments can be removed.
- Every `.go` file carries the AGPL license header — copy from a neighboring file for new files.
- Lint rules that fail in surprising ways:
    - revive `redefines-builtin-id` is ON — no params/vars named `max`, `min`, `any`, etc.
    - revive `error-strings` is ON — error messages start lowercase, no trailing punctuation.
    - `golines` is a formatter with a short line budget and will not auto-wrap long literals — wrap long `errors.New`/
      `fmt.Errorf` args manually (see the `ErrDecompressedSizeExceeded` var block).
    - forbidigo bans `ioutil.*`; depguard bans `github.com/pkg/errors` (use stdlib `errors`).
    - tagliatelle requires snake_case yaml/json tags.
    - gosec `G304` is excluded only for `internal/filesystem.go` and `cmd/decode_test.go`.

## Architecture

- Entrypoint `papercrypt.go` sets go-embedded assets (fonts, LICENSE, EFF word list, THIRD_PARTY.md) onto `cmd` package
  pointers, then calls `cmd.Execute()`.
- `file_format`: binary container wire format v5 — magic `PC`, format version byte `05` (`CurrentBinaryFormatVersion`;
  decode rejects any other byte). Table in README. Package-level functions (`MarshalBinary`, `UnmarshalBinary`,
  `UnmarshalEnvelope`, `SerializeBinary`, `DeserializeBinary`, `DeserializeText`, `DecodeData`, `GetText`, `GetPDF`)
  drive the pipeline; split across `binary_*`, `text_*`, `pdf_*`, `json.go`, `decode.go` and `format_handler.go`.
- `file_format/envelope`: `Wrap`/`Unwrap` with an injectable `ContentEncoder` (currently Base45), gzip only when it
  shrinks the payload. Header = `PC` + base36 (info) + base36 (version) + base45 (CRC-32) + base45 (payload) —
  documented in README; keep in sync.
- Decompression capped at 1 GiB (`maxDecompressedSize`); `scan --unlimited` disables it. On a cap hit,
  `envelope.ErrDecompressedSizeExceeded` fires and scan appends a `use --unlimited` hint.
- `codematrix` = QR encode (boombuler/barcode) / decode (gozxing); `pdf` = gofpdf with embedded Noto Sans/Inconsolata.

## Tracked artifacts

- `examples/*.pdf` are committed; regenerate via `task docs:examples` (requires `pdfcpu`) after envelope/container
  format changes. The checked-in PDFs predate the base36 envelope header and carry old-format QRs.
- `coverage.txt`, `dist/`, `bin/`, `manpages/`, `completions/` are generated; `task clean` removes them. `task test`
  leaves no residue.

## Compatibility

- Software major v3 decodes only v3 documents (README); distinct from the container wire format byte above (`05`). Keep
  envelope/container wire formats backward compatible within the branch; the base36 header alphabet and the 1 GiB cap
  are recent changes.
