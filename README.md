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
you can use [seedtool-cli](https://github.com/BlockchainCommons/seedtool-cli):

```bash
seedtool --out bip39 --count 32
```

or use this tool directly:

```bash
papercrypt generateKey --words 24 --out mnemonic.txt
```

to generate a 24 word mnemonic phrase.

Save your data as a JSON file, let's call it `data.json`.

Example `data.json` file:

```json
{
  "bitwarden": {
    "backup_localtion": "https://your-bucket.s3.amazonaws.com/bitwarden-backup.tar.gz.aes",
    "access_key_id": "YOUR_S3_ACCESS_KEY_ID",
    "secret_key": "YOUR_S3_SECRET_KEY",
    "encryption_key": "your-backup-encryption-key",
    "admin_token": "your-bitwarden-admin-token"
  }
}
```

Then, run

```bash
papercrypt generate --in-file data.json --out-file output.txt.gpg
```

to generate the file containing your data, and the decryption instructions.

The program will ask you for an encryption key,
for which you can use your mnemonic phrase from earlier.

![example.png](example.png)
