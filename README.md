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

TODO(2023-08-11): Add installation instructions.

### From source

1. **Install Go**: Ensure you have Go installed. If not, you can download it from [here](https://go.dev/dl/).

2. **Clone the Repository**: Clone the repository to your local machine:

```bash
git clone https://github.com/TMUniversal/papercrypt.git
```

3. **Navigate to the Directory**:

```bash
cd papercrypt
```

4. **Install Dependencies**: Run the following command to install the required dependencies:

```bash
go mod tidy
```

5. **Build PaperCrypt**:

```bash
go generate ./... && \
go build -o bin/papercrypt
```

or using [`task`](https://taskfile.dev/#/installation):

```bash
task build
```

## Usage

Generating a new key:

A 24 word mnemonic phrase is suitable for real-world use,
but you can use any string of words or characters.

Generate one with your tool of choice,
you can use [seedtool-cli](https://github.com/BlockchainCommons/seedtool-cli):

```bash
seedtool --out bip39 --count 32
```

or use this tool directly:

```bash
papercrypt generateKey --words 24 --out mnemonic.txt
```

to generate a 24 word mnemonic phrase.

Save your data as a JSON file: `data.json`, for example.

Example `data.json` file:

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

Please see the [examples](examples) directory for the generated PDF files.

## Contributing

Contributions to PaperCrypt are welcomed and encouraged! If you have suggestions for improvements, bug fixes, or new
features, please feel free to submit a pull request.

## License

PaperCrypt is licensed under the terms of the GNU Affero General Public License, version 3.0 or
later ([GNU AGPL-3.0-or-later](COPYING)).

[![License Logo](https://www.gnu.org/graphics/agplv3-with-text-162x68.png)](https://www.gnu.org/licenses/agpl-3.0.en.html)

## Acknowledgments

PaperCrypt is developed leveraging the power of Go and a suite of dependable open source libraries.
We extend our gratitude to the developers behind
[GopenPGP](https://github.com/ProtonMail/gopenpgp), [GoFPDF](https://github.com/jung-kurt/gofpdf),
and other foundational components.
