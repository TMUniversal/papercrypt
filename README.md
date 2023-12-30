<p align="center">
  <h3 align="center">PaperCrypt</h3>
    <p align="center">Printable Backup Documents</p>
    <p align="center">
      <a href="https://goreportcard.com/report/github.com/TMUniversal/papercrypt"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/TMUniversal/papercrypt" /></a>
      <a href="https://pkg.go.dev/github.com/TMUniversal/papercrypt"><img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/TMUniversal/papercrypt.svg" /></a>
      <a href="https://github.com/TMUniversal/papercrypt/releases"><img alt="GitHub release (latest SemVer)" src="https://img.shields.io/github/v/release/TMUniversal/papercrypt?sort=semver" /></a>
      <a href="https://github.com/TMUniversal/papercrypt"><img alt="GitHub" src="https://img.shields.io/github/license/TMUniversal/papercrypt" /></a>
    </p>
</p>

---

PaperCrypt is a Go-based command-line tool designed to enhance the security of your sensitive data through the
generation of printable backup documents.
These documents, referred to as "PaperCrypt" Documents, combine the robust
encryption capabilities of the [OpenPGP](https://gopenpgp.org/)
with the resilience and simplicity of a physical hardcopy.
This ensures the confidentiality and integrity of your data,
while also providing a physical backup that 's not susceptible to digital threats.

> Please note that to decrypt the data from a PaperCrypt Document, you will need the original passphrase used during the
> encryption process.
> Treat this passphrase with care; if it is lost, the data will be irrecoverable.

## Features

- **Encryption**: PaperCrypt utilizes the [GopenPGP](https://gopenpgp.org/) library to apply advanced cryptographic
  techniques, securing your data with industry-standard algorithms.

- **PDF Generation**: PaperCrypt can output PDF documents ready for printing, that include all the information needed to
  recover the encrypted data.

- **Data Integrity**: To verify the integrity of the data, PaperCrypt embeds checksums within the encrypted data section
  of its documents. This ensures that the data remains unaltered during backup and restoration processes.

- **Offline Security**: By generating printable backup documents, PaperCrypt offers an offline solution to
  safeguard your sensitive data against online threats, as well as an option to store your data in an off-site
  location. This provides a layer of security, as it ensures that your data remains safe and accessible even in the
  event of a catastrophic failure, malicious attack, or natural disaster.

## Installation

### Pre-built binaries

Pre-built binaries for PaperCrypt are available for download from
the [releases](https://github.com/TMUniversal/papercrypt/releases) page.

> The pre-built binaries are preferred over manual installation, as they include version information. The GitHub
> releases also included signatures for the binaries.

#### Homebrew (Linux, macOS)

```bash
brew tap TMUniversal/homebrew-tap
brew install papercrypt
```

#### Scoop (Windows)

Make sure you have [scoop](https://scoop.sh/) installed,
alongside `git` (`scoop install git`) to be able to add the bucket.

```bash
scoop bucket add tmuniversal https://github.com/tmuniversal/scoop-bucket.git
scoop install papercrypt
```

### From source (manual)

1. **Install Go**: Ensure you have Go installed. If not, you can download it from [here](https://go.dev/dl/).

2. **Clone the Repository**: Clone the repository to your local machine:

```bash
git clone https://github.com/TMUniversal/papercrypt.git
```

3. **Navigate to the Directory**:

```bash
cd papercrypt
```

4. **Build the binary**:

Build using [`task`](https://taskfile.dev/#/installation):

```bash
task build
```

or through [`goreleaser`](https://goreleaser.com/install/):

```bash
goreleaser build --snapshot --clean --single-target
```

### From source (go install)

> This method is not recommended, as it won't include the version information in the binary.

1. **Ensure Go is installed**: See [step 1 from manual installation](#from-source-manual).
2. **Install PaperCrypt**: Run the following command to install PaperCrypt:

```bash
go install github.com/TMUniversal/papercrypt@latest
```

### Running with Docker

You can also run PaperCrypt using Docker, with the following command:

```bash
docker run --rm -it -v $(pwd):/data ghcr.io/tmuniversal/papercrypt:latest
```

With `-v $(pwd):/data` mounting the current working directory as `/data` in the container,
allowing the container to read and write to host storage.

On Windows, the command is slightly different:

```bash
docker run --rm -it -v ${PWD}:/data ghcr.io/tmuniversal/papercrypt:latest
```

Note that `-t` is required so that the program can prompt for a passphrase.

### Verifying artifacts

First, you'll need to download the archive and signature file (`.sig`) for your version from
the [releases page](https://github.com/TMUniversal/papercrypt/releases), pay attention to the
version (`papercrypt version`), your OS and architecture. You will also need the public key ([`cosign.pub`]).

The pre-built binaries are signed through [`cosign`](https://github.com/sigstore/cosign#installation).

To verify the signature, you can run the following command:

```bash
cosign verify-blob \
  --key cosign.pub \
  --signature papercrypt_$(uname -s)_$(uname -m).tar.gz.sig \
  papercrypt_$(uname -s)_$(uname -m).tar.gz
```

and for Windows:

```bash
cosign verify-blob \
  --key cosign.pub \
  --signature papercrypt_Windows_x86_64.zip.sig \
  papercrypt_Windows_x86_64.zip
```

## Usage

General notes:

- `--in` and `--out` can be omitted, in which case `stdin` and `stdout` are used.
- This means `papercrypt decode --in - --out - < qr.txt > data.json` is equivalent
  to `papercrypt decode < qr.txt > data.json`
- Commands, as well as their flags, can be abbreviated to their shortest unique prefix:
  - `papercrypt generate` can be abbreviated to `papercrypt g`
- that is `papercrypt generate --in data.json --out output.pdf` can be abbreviated
  to `papercrypt g -i data.json -o output.pdf`

### Generating a key phrase

A 24 word mnemonic phrase is suitable for real-world use,
but you can use any string of words or characters.

Generate one with your tool of choice,
you can run:

```bash
papercrypt generateKey --words 24 --out mnemonic.txt
```

to generate a 24 word mnemonic phrase.

[![key example](examples/demo/key.gif)](examples/)

#### The passphrase sheet

PaperCrypt is able to generate a printable _Phrase Sheet_,
which is a two-page document containing 135 words from the EFF large word list,
chosen with a seeded random number generator.

If no seed is passed to the command, one will be generated using the system's entropy source.

[Example](examples/phrase.pdf):

```bash
papercrypt phraseSheet --out phrase-sheet.pdf ExampleAbcA=
```

Here, `ExampleAbcA=` is the base64-encoded seed, which is used to generate the word list.
The seed will is also present on the generated PDF document,
so you can regenerate the same word list later, even if you allowed the seed to be chosen at random.

Using the phrase sheet, you can select a number of words from to form your mnemonic phrase.

### Generating a PaperCrypt document

Save your data as a file, `data.json`, for example:

```json
{
  "bitwarden": {
    "backup_location": "https://your-bucket.s3.amazonaws.com/bitwarden-backup.tar.gz",
    "access_key_id": "YOUR_S3_ACCESS_KEY_ID",
    "secret_key": "YOUR_S3_SECRET_KEY",
    "encryption_key": "your-backup-encryption-key",
    "admin_token": "your-bitwarden-admin-token"
  }
}
```

Then, run

```bash
papercrypt generate --in data.json --out output.pdf
```

to generate the file containing your data, and the decryption instructions.

The program then asks you for an encryption key,
for which you can use your mnemonic phrase from earlier.

> You can also pass the data through `stdin`, simply omit the `--in` flag.
> The caveat is that, when on Windows, you can't be prompted for your passphrase,
> so you would have to pass it with the `--passphrase` flag.

[![generate example](examples/demo/generate.gif)](examples/output.pdf)

Please see the [examples](examples) directory for the generated PDF files.

### Restoring a PaperCrypt document

To restore your data from a PaperCrypt document,
you must first re-construct the document from the printed copy.
This can be done either by saving the QR code as an image file,
and [passing it to the command-line](#using-the-qr-code),
or by copy-pasting the text from the printed document (would have to run [OCR](https://www.adobe.com/acrobat/guides/what-is-ocr.html "optical character recognition")).

#### Using the QR code

Save the QR code as an image file (a screenshot should do), for example `qr.png`.

Then, run

```bash
papercrypt qr --in qr.png --out data.txt
```

#### Decoding from text

Once you have the text from the printed document,

<details>
<summary>Click to expand</summary>

which should look something like this:

```
# PaperCrypt Version: 1.0.7-next
# Content Serial: FVCUW7
# Purpose: Example Sheet
# Comment: Regular PDF Example
# Date: Sat, 23 Sep 2023 14:07:34.051057700 CEST
# Content Length: 362
# Content CRC-24: f6fd74
# Content CRC-32: 13938adb
# Content SHA-256: Z8h1aiYWzS6OCGzArbLBxex2ROQ9Do2/wga55qmlt4I=
# Header CRC-32: b19162f6


 1: C3 2E 04 09 03 08 16 C6 62 5A D9 78 6F 63 E0 D0 B8 8F 47 BE F5 1B 365028
 2: B2 BE C1 DC 71 FC 1C C5 D4 0A 0D 32 FC D1 32 E1 52 A5 5A 0C 62 84 E4B7A0
 3: 6E F7 87 20 D2 C0 77 01 EB 82 C2 E5 B3 B6 28 5F 97 D8 35 48 42 1B B6B934
 4: C6 9B B9 F0 18 B9 DC 19 F9 89 E5 14 F6 EC 9E 7D 39 7E E5 48 E4 27 1F43CC
 5: 4A 81 7A 1D 6B 24 89 AE B1 91 80 C9 C6 60 49 F4 29 A7 3B 89 42 3C CA8180
 6: BE 89 6A 43 B3 E5 89 8F 94 21 1F 07 65 BA 19 75 92 21 B7 8D 27 DE 433210
 7: 76 CF F1 A4 52 9B B2 81 64 DC FB 15 5B C4 2B EB B6 CD 3F 0C 0A 93 4B4CE1
 8: 14 3D 47 0D 91 06 90 60 9B D4 B6 14 88 E9 24 3A D7 97 53 02 49 F0 5BB0D9
 9: AE A5 B2 D3 15 7E 61 5D 67 15 AD 01 05 5F EE 4C 7B B1 B4 98 19 91 F11A64
10: 37 91 0B A0 06 8C 0C 2D 34 7B A4 21 BD 45 F3 5D 2D DD F6 DD 22 B9 DDC18B
11: 4B 18 38 B0 93 38 19 92 E3 F1 37 7A 97 E8 AE 8D 1A 7B A4 3A 9D F4 BDDBDF
12: C0 6C 3A 27 05 AB AF C4 E9 DD 33 6A 74 6B F2 09 14 06 2D 21 60 A2 ADC47E
13: D5 79 96 69 5D DE BB AC E2 F8 B5 3F DD E6 72 CB D8 7C C2 8A B9 69 F0281A
14: 4F 48 BA 54 10 94 36 02 3F 3F D8 67 8D E1 89 C7 A3 C3 D3 F0 97 6B 5649F8
15: E7 F9 47 67 60 8B 7D 83 7E D1 71 56 E0 62 2B 8F 3B 38 62 86 01 1C FCB09F
16: 85 7B 2D A1 42 5C 2B 8E AA 48 39 30 E5 73 F3 EE 8E E0 F3 E2 9A F5 4B29AF
17: 07 92 46 67 23 BB B6 A6 A4 68 256181
18: F6FD74
```

Note the two empty lines between the header and the data.

</details>

Decode and decompress the data:

```bash
papercrypt decode --in data.txt --out data.json
```

The command prompts for the encryption key, which you could specify with `--passphrase|-P`:

```bash
papercrypt decode -i data.txt -o data.json -P "super-secret-key"
```

### Full pipeline

[![demo](examples/demo/demo.gif)](examples/output.pdf)

## Contributing

Contributions to PaperCrypt are welcomed and encouraged! If you have suggestions for improvements, bug fixes, or new
features, please feel free to submit a pull request.
Refer to [CONTRIBUTING.md](CONTRIBUTING.md) for more information.

## License

PaperCrypt is licensed under the terms of the GNU Affero General Public License, version 3.0 or
later ([GNU AGPL-3.0-or-later](COPYING)).

[![License Logo](https://www.gnu.org/graphics/agplv3-with-text-162x68.png)](https://www.gnu.org/licenses/agpl-3.0.en.html)

## Acknowledgments

PaperCrypt is developed leveraging the power of Go and a suite of dependable open source libraries.
We extend our gratitude to the developers behind
[GopenPGP](https://github.com/ProtonMail/gopenpgp), [GoFPDF](https://github.com/jung-kurt/gofpdf),
and other foundational components.

[`cosign.pub`]: https://github.com/TMUniversal/papercrypt/blob/main/cosign.pub
