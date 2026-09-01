## [3.5.1](https://github.com/TMUniversal/papercrypt/compare/v3.5.0...v3.5.1) (2026-09-01)


### Bug Fixes

* **decompression:** avoid LimitReader overflow at max int limit ([a73356d](https://github.com/TMUniversal/papercrypt/commit/a73356df070ade1b8343e430dd9d30b40f1f373c))
* **envelope:** reject reserved header bits in ParseHeader ([8087486](https://github.com/TMUniversal/papercrypt/commit/80874866de69dd2585194068cb25d1341f0686ce))
* **file_format:** accept any SerializeBinary width in DeserializeBinary ([f79ca7f](https://github.com/TMUniversal/papercrypt/commit/f79ca7f05c1b8914667ca8b24c008d2876751980))
* **file_format:** correct base32 label in serial error message ([833a98c](https://github.com/TMUniversal/papercrypt/commit/833a98c314548d6fefcdbf345977dfdc32fd2a5a))
* **file_format:** guard DecodeData against nil document ([dfe2a39](https://github.com/TMUniversal/papercrypt/commit/dfe2a393bc43cc941b739829355400b042aee025))
* **file_format:** never underflow serial buffer before slicing ([7350db3](https://github.com/TMUniversal/papercrypt/commit/7350db30bb22ddf9053ebeff9280f5585ea92155))
* **file_format:** nil-guard GetText before dereferencing p ([81aadbe](https://github.com/TMUniversal/papercrypt/commit/81aadbecdada23276ac63191d620f1258c7a3726))
* **file_format:** reject unrepresentable versions on binary marshal ([95673e5](https://github.com/TMUniversal/papercrypt/commit/95673e506dc48772108649a8c6bbeaece5becd50))
* **file_format:** treat absent CRC and date headers as required-field errors ([c5152e8](https://github.com/TMUniversal/papercrypt/commit/c5152e8a43daeda09b09d8984a04af21b4f4f96d))
* **file_format:** verify every data line number is consecutive ([89f1dfd](https://github.com/TMUniversal/papercrypt/commit/89f1dfdd0769385e68089c5ec098ca11960264d9))
* **terminal:** close tty on all paths and drop duplicate prompt ([5093516](https://github.com/TMUniversal/papercrypt/commit/50935162762de286becdfd06343fbc8c36cb1827))
* **terminal:** prompt before /dev/tty password read ([c883273](https://github.com/TMUniversal/papercrypt/commit/c883273d42ac792ce4d7468f2b0be0b91db96e5d))


### Performance Improvements

* **file_format:** streamline hex serialization and line parsing ([a45ee62](https://github.com/TMUniversal/papercrypt/commit/a45ee627082059c83f09b8e30dd8fd0ec10e9086))

# [3.5.0](https://github.com/TMUniversal/papercrypt/compare/v3.4.1...v3.5.0) (2026-08-29)


### Bug Fixes

* **envelope:** use ErrInvalidType for wrong-type headers ([e277e53](https://github.com/TMUniversal/papercrypt/commit/e277e5350c06e58b95adad44ef64a849beee6a3c))


### Features

* **envelope:** replace PCE magic with PC + base32 info header ([927eed7](https://github.com/TMUniversal/papercrypt/commit/927eed773d596eb52ba28a01bb5ac729bdc947a8))
* **scan:** add --unlimited flag to bypass decompressed size cap ([e88ed8f](https://github.com/TMUniversal/papercrypt/commit/e88ed8f385d5292babeb0cf10e6a36fbe3e103a2))
* **scan:** hint at --unlimited when the decompressed size limit is hit ([dfc9553](https://github.com/TMUniversal/papercrypt/commit/dfc9553f2f55a3cc4bc6cd12abcd168482470e76))
* update binary formats ([ee38acd](https://github.com/TMUniversal/papercrypt/commit/ee38acdc2472d67de94057ebaeb7ae2206a99218))

## [3.4.1](https://github.com/TMUniversal/papercrypt/compare/v3.4.0...v3.4.1) (2026-08-28)


### Bug Fixes

* **binary:** do not prepend `v` prefix in version ([5c80cd7](https://github.com/TMUniversal/papercrypt/commit/5c80cd713ebb04cc496576e12d16eb8e8485ab54))

# [3.4.0](https://github.com/TMUniversal/papercrypt/compare/v3.3.0...v3.4.0) (2026-08-28)


### Features

* add project url to doc comment ([cb26f66](https://github.com/TMUniversal/papercrypt/commit/cb26f66f27d53917623ce2b7381e111f5d4f9966))

# [3.3.0](https://github.com/TMUniversal/papercrypt/compare/v3.2.0...v3.3.0) (2026-08-28)


### Features

* make checksum errors more informative ([bdff797](https://github.com/TMUniversal/papercrypt/commit/bdff7979e566410c928d2e6a5e1804abe724288d))

# [3.2.0](https://github.com/TMUniversal/papercrypt/compare/v3.1.2...v3.2.0) (2026-08-28)


### Bug Fixes

* remove double base45 encoding in QR payload ([f45a3b2](https://github.com/TMUniversal/papercrypt/commit/f45a3b2f3c94306cbaef0e9e516f749870aa46bd))
* return ErrBinaryTruncated instead of panicking on truncated binary ([4deffa5](https://github.com/TMUniversal/papercrypt/commit/4deffa58f025e118874de95e11741762c32d673e))
* trim whitespace from envelope input in scan --from-binary ([0a45d89](https://github.com/TMUniversal/papercrypt/commit/0a45d89260f8c3bcf9468a5c9a401afc11bac373))
* update docs comment positioning ([86ac7c1](https://github.com/TMUniversal/papercrypt/commit/86ac7c1944c35a34ff98c713e61c58d82510ffa4))


### Features

* add docs at page bottom ([f7f0e7a](https://github.com/TMUniversal/papercrypt/commit/f7f0e7a4a135a4ca076aa778d616ba2e2cde0f9d))
* **binary:** add 3-byte version field and fix timestamp precision ([19ec21f](https://github.com/TMUniversal/papercrypt/commit/19ec21fda082373104922991d488f3f88d636f31))
* **pdf:** add program name and version to footer ([cad7ffe](https://github.com/TMUniversal/papercrypt/commit/cad7ffe9b59d090e6b01aafd84d9a4a9abe4af15))
* replace JSON QR container with compact binary format ([f75cf79](https://github.com/TMUniversal/papercrypt/commit/f75cf79b2f4ebbdb89aee57f394ca8fbbb7d332c))
* switch to qr-code with base45 encoding ([cb613a7](https://github.com/TMUniversal/papercrypt/commit/cb613a7cf42db53c9e7681f0926252dd3c6e5ee5))

## [3.1.2](https://github.com/TMUniversal/papercrypt/compare/v3.1.1...v3.1.2) (2026-08-26)


### Bug Fixes

* trim tag prefix from version string ([f65d95a](https://github.com/TMUniversal/papercrypt/commit/f65d95a0148f27a5a8e0ce06ce849cbb522da71f))

# [3.1.0](https://github.com/TMUniversal/papercrypt/compare/v3.0.0...v3.1.0) (2026-08-25)


### Bug Fixes

* update gettext for new parameters ([5940b15](https://github.com/TMUniversal/papercrypt/commit/5940b159d1d6e68491e498054776f65b9e2f8f00))


### Features

* remove content crc-32 in favour of existing sha256 ([64ecfa1](https://github.com/TMUniversal/papercrypt/commit/64ecfa1a85a5760a248ad0985f64c4b9bf7cd707))
* remove crc24 from header ([bd9ea38](https://github.com/TMUniversal/papercrypt/commit/bd9ea384bcad91ba5b1a405a141d2f07011c4a3d))
* skip passphrase input in raw mode ([bed42af](https://github.com/TMUniversal/papercrypt/commit/bed42afa5d3d0f695185061f455024b460a11949))

# [2.1.0](https://github.com/TMUniversal/papercrypt/compare/v2.0.3...v2.1.0) (2026-08-25)


### Bug Fixes

* convert colour model to remove 16-bit depth ([21a5ce1](https://github.com/TMUniversal/papercrypt/commit/21a5ce100ca8fee3d9bcf5b6d66c43112543f5ac))
* **deps:** undo accidental papercrypt v1 upgrade ([5c25deb](https://github.com/TMUniversal/papercrypt/commit/5c25deb28e024d3c1ac10b78dc9df002c9776490))
* limit gzip decoding from scanned code ([bd91145](https://github.com/TMUniversal/papercrypt/commit/bd91145c9f43abb97fb58b6aa8c46998acdc9a6c))
* set version support to v3 ([a78489b](https://github.com/TMUniversal/papercrypt/commit/a78489bdadca12d8b89b4608acf24e8a1de4bc2b))
* update copyright notice ([65ea0c5](https://github.com/TMUniversal/papercrypt/commit/65ea0c573619b4af90045a9a9be78a60185e6627))
* update version handling ([a1e211e](https://github.com/TMUniversal/papercrypt/commit/a1e211e95db1f38134d88f65dba9f3d58ccbb718))


### Features

* add encoding modes ([ad5d612](https://github.com/TMUniversal/papercrypt/commit/ad5d6128353251e8b84def7386f38036a5cf93cb))
* add timestamp format for marshalled document ([6af36d4](https://github.com/TMUniversal/papercrypt/commit/6af36d48a910a32573fae614e78c6802e2079dc5))
* compress 2d code contents ([6944ba7](https://github.com/TMUniversal/papercrypt/commit/6944ba7443fd5e3044bdfdbf48e838a64f79ca7a))
* disable payload limit via flag ([309cd98](https://github.com/TMUniversal/papercrypt/commit/309cd98d1bdcad3315d83bb35d088d585d123591))
* do not compress data in raw mode ([727c9a7](https://github.com/TMUniversal/papercrypt/commit/727c9a7faf1db1dd574b7aecb5cea584be65b002))
* remove extra encoding and validation header ([b5a23ab](https://github.com/TMUniversal/papercrypt/commit/b5a23ab264d05617ddefa77e0896f7ed11b76f08))
* remove inner compression layer from encrypted mode ([df5ba0b](https://github.com/TMUniversal/papercrypt/commit/df5ba0b4dce3ad8435ddbb08024827085ee79f32))
* update json format ([0d55fef](https://github.com/TMUniversal/papercrypt/commit/0d55fef48eb8cf671b651d3d7696995de4dfe9b0))


### Performance Improvements

* remove image-to-image copy ([9783e62](https://github.com/TMUniversal/papercrypt/commit/9783e62dc863173bae193a25d548a86b99bfa36a))

## [2.0.3](https://github.com/TMUniversal/papercrypt/compare/v2.0.2...v2.0.3) (2025-09-28)


### Bug Fixes

* **deps:** update module github.com/caarlos0/go-version to v0.2.2 ([7ae85c4](https://github.com/TMUniversal/papercrypt/commit/7ae85c4e170a83b9bbcb140218637d257e13b0ef))
* **deps:** update module github.com/muesli/mango-cobra to v1.3.0 ([6ed0cdf](https://github.com/TMUniversal/papercrypt/commit/6ed0cdfe3bc25e22be9e92a6d56e9b9c4f77b643))
* **deps:** update module github.com/spf13/cobra to v1.10.1 ([48d0a3f](https://github.com/TMUniversal/papercrypt/commit/48d0a3f772c384cab8689782605e1bc59467503a))
* **deps:** update module github.com/stretchr/testify to v1.11.1 ([f9361d9](https://github.com/TMUniversal/papercrypt/commit/f9361d941dbd9ea205246fec2b1589d30dfa2f9c))
* **deps:** update module golang.org/x/term to v0.35.0 ([00a4339](https://github.com/TMUniversal/papercrypt/commit/00a4339d3d3c23d4d7ad81748cc7e95c773991c4))

## [2.0.2](https://github.com/TMUniversal/papercrypt/compare/v2.0.1...v2.0.2) (2025-08-12)


### Bug Fixes

* **deps:** update module github.com/caarlos0/log to v0.4.8 ([f30215e](https://github.com/TMUniversal/papercrypt/commit/f30215e477ef7c10dec4a85be4a35fd63351cf47))
* **deps:** update module golang.org/x/term to v0.30.0 ([dd5822b](https://github.com/TMUniversal/papercrypt/commit/dd5822b3d44f6eccf862cf3baf5564f3acfab9af))
* **scan:** close input file only if it is not stdin ([1c0af37](https://github.com/TMUniversal/papercrypt/commit/1c0af37f602bd8b81a468399439aebde4b02173f))
* **tty:** do not close twice ([ec4c908](https://github.com/TMUniversal/papercrypt/commit/ec4c90839ca1dec2d1768659bc1709cc052b03dc))

## [2.0.1](https://github.com/TMUniversal/papercrypt/compare/v2.0.0...v2.0.1) (2024-08-07)


### Bug Fixes

* **go:** update go mod to reflect version update ([45cee4f](https://github.com/TMUniversal/papercrypt/commit/45cee4f6bd378b7173f8c6a868c901f8f206f308))

# [2.0.0](https://github.com/TMUniversal/papercrypt/compare/v1.3.0...v2.0.0) (2024-08-07)

### Bug Fixes

* **compat:** update v1 date parser ([a66c115](https://github.com/TMUniversal/papercrypt/commit/a66c115c20dc2a47cd078b47bcd49a4b4876878d))
* **decode:** normalize line endings before version detection ([722e047](https://github.com/TMUniversal/papercrypt/commit/722e047a35992c678efb8130788593d0c22a8272))


### Features

* replace qr code with aztec to hold more data ([8f6ef58](https://github.com/TMUniversal/papercrypt/commit/8f6ef58be5403f52c2b18516922de8ce237551d4))
* update file format to support raw data ([d970476](https://github.com/TMUniversal/papercrypt/commit/d9704768510c8ca6b83fe8508abf5bc2f1a94bc5))

### BREAKING CHANGES

* update file format

# [1.3.0](https://github.com/TMUniversal/papercrypt/compare/v1.2.7...v1.3.0) (2024-07-26)


### Features

* **cmd:** [qr] add json flags to support external qr code readers ([ec3d52f](https://github.com/TMUniversal/papercrypt/commit/ec3d52f783eaa029d06639fc0962e39f0f4380c7))

## [1.2.7](https://github.com/TMUniversal/papercrypt/compare/v1.2.6...v1.2.7) (2024-07-25)


### Bug Fixes

* **deps:** update module github.com/caarlos0/log to v0.4.6 ([a97bd10](https://github.com/TMUniversal/papercrypt/commit/a97bd10b52c45f0d67030264cb1b9f22c7ab46d4))

## [1.2.6](https://github.com/TMUniversal/papercrypt/compare/v1.2.5...v1.2.6) (2024-07-06)


### Bug Fixes

* **deps:** update module golang.org/x/term to v0.22.0 ([6fe9fe4](https://github.com/TMUniversal/papercrypt/commit/6fe9fe48febe6b342089c941fa40b2cccc8f9db7))

## [1.2.5](https://github.com/TMUniversal/papercrypt/compare/v1.2.4...v1.2.5) (2024-06-27)


### Bug Fixes

* **deps:** update module github.com/spf13/cobra to v1.8.1 ([baaa89e](https://github.com/TMUniversal/papercrypt/commit/baaa89ed3ede0bfe08364aa1488f44e238d11fe5))

## [1.2.4](https://github.com/TMUniversal/papercrypt/compare/v1.2.3...v1.2.4) (2024-06-07)


### Bug Fixes

* **deps:** update module golang.org/x/term to v0.21.0 ([9f46b6e](https://github.com/TMUniversal/papercrypt/commit/9f46b6e4fd3b3c9219bccc5dc01b8e304ad2ccc3))

## [1.2.3](https://github.com/TMUniversal/papercrypt/compare/v1.2.2...v1.2.3) (2024-05-28)


### Bug Fixes

* **deps:** update module github.com/caarlos0/log to v0.4.5 ([003a2e3](https://github.com/TMUniversal/papercrypt/commit/003a2e3ea1c1e409ce68eac91a05fc1d4f178e66))
* **deps:** update module github.com/charmbracelet/lipgloss to v0.11.0 ([7daafc9](https://github.com/TMUniversal/papercrypt/commit/7daafc9ddd3a25f51719a36e797111f2a73fb3c8))

## [1.2.2](https://github.com/TMUniversal/papercrypt/compare/v1.2.1...v1.2.2) (2024-05-18)


### Bug Fixes

* **deps:** update module golang.org/x/term to v0.20.0 ([6a206f5](https://github.com/TMUniversal/papercrypt/commit/6a206f534b9e9c68c4a7c9b3b84a4b488d8606cc))

## [1.2.1](https://github.com/TMUniversal/papercrypt/compare/v1.2.0...v1.2.1) (2024-04-19)


### Bug Fixes

* **deps:** update module golang.org/x/term to v0.19.0 ([33b1c0e](https://github.com/TMUniversal/papercrypt/commit/33b1c0ec2fea9c9a21b465dd4b5d7c9565dd0629))

# [1.2.0](https://github.com/TMUniversal/papercrypt/compare/v1.1.8...v1.2.0) (2024-04-02)


### Features

* **generate:** underlay every second line in text output ([db33b4e](https://github.com/TMUniversal/papercrypt/commit/db33b4ea010cb97b9dc856ce043e4374f8a86ba4))

## [1.1.8](https://github.com/TMUniversal/papercrypt/compare/v1.1.7...v1.1.8) (2024-03-21)


### Bug Fixes

* **deps:** update module github.com/charmbracelet/lipgloss to v0.10.0 ([c072b39](https://github.com/TMUniversal/papercrypt/commit/c072b39abe06c73726006130da37536e70673b59))
* **deps:** update module golang.org/x/term to v0.18.0 ([19ba04a](https://github.com/TMUniversal/papercrypt/commit/19ba04a429a7f8009da385c740cda330d302d022))

## [1.1.7](https://github.com/TMUniversal/papercrypt/compare/v1.1.6...v1.1.7) (2024-02-08)


### Bug Fixes

* **deps:** update module github.com/protonmail/gopenpgp/v2 to v2.7.5 ([186f158](https://github.com/TMUniversal/papercrypt/commit/186f15850b28a60ea4ffef3c988f629f4ec180a4))
* **deps:** update module golang.org/x/term to v0.17.0 ([5e30e86](https://github.com/TMUniversal/papercrypt/commit/5e30e866fc9eb42a197ea2c7193a3ca25993e3cf))

## [1.1.6](https://github.com/TMUniversal/papercrypt/compare/v1.1.5...v1.1.6) (2024-01-04)


### Bug Fixes

* **deps:** update module golang.org/x/term to v0.16.0 ([d0b01bf](https://github.com/TMUniversal/papercrypt/commit/d0b01bfa0a1571f9d3ad80b673db56d336581121))

## [1.1.5](https://github.com/TMUniversal/papercrypt/compare/v1.1.4...v1.1.5) (2023-12-21)


### Bug Fixes

* **deps:** update module golang.org/x/term to v0.15.0 ([da3a76f](https://github.com/TMUniversal/papercrypt/commit/da3a76f03aac751afc0aea89e25857ae01719436))

## [1.1.4](https://github.com/TMUniversal/papercrypt/compare/v1.1.3...v1.1.4) (2023-11-21)


### Bug Fixes

* **deps:** update module github.com/protonmail/gopenpgp/v2 to v2.7.4 ([d8236f4](https://github.com/TMUniversal/papercrypt/commit/d8236f495105c1dff909e3fddf1b8c339d455418))
* **deps:** update module github.com/spf13/cobra to v1.8.0 ([cef8075](https://github.com/TMUniversal/papercrypt/commit/cef8075af6383f6668b3dd8de5869f7c252c2dc5))
* **deps:** update module golang.org/x/term to v0.14.0 ([0f10fca](https://github.com/TMUniversal/papercrypt/commit/0f10fca7546f769afbe80fab14d6146d4194ed0d))

## [1.1.3](https://github.com/TMUniversal/papercrypt/compare/v1.1.2...v1.1.3) (2023-10-22)


### Bug Fixes

* **deps:** update module github.com/caarlos0/log to v0.4.3 ([93ab9b7](https://github.com/TMUniversal/papercrypt/commit/93ab9b7a421c4d86acea22c21f0b30a680576b48))
* **deps:** update module github.com/caarlos0/log to v0.4.4 ([ad7f49a](https://github.com/TMUniversal/papercrypt/commit/ad7f49a7859b56f29dfd295a6eaf195a0cf25abc))
* **deps:** update module github.com/charmbracelet/lipgloss to v0.9.0 ([2c89529](https://github.com/TMUniversal/papercrypt/commit/2c895294d0e65e182b5529cc566fbc516cfaf5f7))
* **deps:** update module github.com/charmbracelet/lipgloss to v0.9.1 ([7f2d939](https://github.com/TMUniversal/papercrypt/commit/7f2d9392d199b0ab7744328460aac638accab678))

## [1.1.2](https://github.com/TMUniversal/papercrypt/compare/v1.1.1...v1.1.2) (2023-10-05)


### Bug Fixes

* **deps:** update module golang.org/x/term to v0.13.0 ([7651b8c](https://github.com/TMUniversal/papercrypt/commit/7651b8c0c8114016577e90215c046e5507329581))

## [1.1.1](https://github.com/TMUniversal/papercrypt/compare/v1.1.0...v1.1.1) (2023-09-28)


### Bug Fixes

* close files after error ([33d87e2](https://github.com/TMUniversal/papercrypt/commit/33d87e2a334b05686667f2a1920934ecd9ced895))

# [1.1.0](https://github.com/TMUniversal/papercrypt/compare/v1.0.6...v1.1.0) (2023-09-23)


### Bug Fixes

* **goreleaser:** add missing git-describe details to binary ([2fbeb10](https://github.com/TMUniversal/papercrypt/commit/2fbeb1032f0d7e8f3c4305132750cae73a12f204))


### Features

* compress secret data before encryption ([b733303](https://github.com/TMUniversal/papercrypt/commit/b733303a4950f4d0fb630adade7480544cbb029c))

# [1.0.6](https://github.com/TMUniversal/papercrypt/compare/v1.0.5...v1.0.6) (2023-09-22)

# [1.0.5](https://github.com/TMUniversal/papercrypt/compare/v1.0.4...v1.0.5) (2023-09-21)

# [1.0.4](https://github.com/TMUniversal/PaperCrypt/compare/v1.0.3...v1.0.4) (2023-09-21)

# [1.0.3](https://github.com/TMUniversal/PaperCrypt/compare/v1.0.2...v1.0.3) (2023-09-21)


### Bug Fixes

* **generate:** place commands with fail-over in parentheses ([2689fd7](https://github.com/TMUniversal/PaperCrypt/commit/2689fd7d2b6dd5d1a3c378e3b52652c05360e435))

# [1.0.2](https://github.com/TMUniversal/PaperCrypt/compare/v1.0.1...v1.0.2) (2023-09-21)

# [1.0.1](https://github.com/TMUniversal/PaperCrypt/compare/v1.0.0...v1.0.1) (2023-09-21)

# [1.0.0](https://github.com/TMUniversal/PaperCrypt/compare/v1.0.0-beta3...v1.0.0) (2023-09-20)


### Bug Fixes

* **cmd/generate:** allow reading stdin ([f95ca1b](https://github.com/TMUniversal/PaperCrypt/commit/f95ca1b5f8fdd1b9e23a3d3b73dfb72e6fdd011d))
