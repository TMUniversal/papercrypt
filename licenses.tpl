# PaperCrypt Third Party Licenses

This file lists the third party licenses used by PaperCrypt.
It is generated using `go-licenses` (`task docs:third_party`) and is not meant to be edited manually.
{{ range . }}
## {{ .Name }}

* Name: {{ .Name }}
* Version: {{ .Version }}
* License: [{{ .LicenseName }}]({{ .LicenseURL }})

```md
{{ .LicenseText }}
```
{{ end }}
