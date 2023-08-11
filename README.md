# PaperCrypt

A tool to secure access to your sensitive data.
Give a piece of paper to a friend (or trusted third party) and tell them to keep it safe,
you can use it to restore your secrets if you lose access to your computer.

## Installation

TODO(2023-08-11): Add installation instructions.
 
## Usage

Generating a new key:

A 24 word mnemonic phrase is suitable for our purposes,
but you can use any string of words or characters.

Generate one with your tool of choice,
you can use [seedtool-cli](https://github.com/BlockchainCommons/seedtool-cli)
or, if you have it, [Bitwarden](https://bitwarden.com/)'s passphrase generator.

For seedtool-cli, see the [usage instructions](https://github.com/BlockchainCommons/seedtool-cli/blob/master/Docs/MANUAL.md#bip39),

and run

```bash
seedtool --out bip39 --count 32
```

to generate a 24 word mnemonic phrase.

Save your data as a JSON file, let's call it `data.json`.

Then, run

```bash
papercrypt generate --in-file data.json --out-file output.pdf
```

to generate the file containing your data, and the decryption instructions.

The program will ask you for an encryption key,
for which you can use your mnemonic phrase from earlier.
