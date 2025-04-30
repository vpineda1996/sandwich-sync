module github.com/vpineda1996/sandwich-sync

go 1.23.0

toolchain go1.23.5

require (
	github.com/Rhymond/go-money v1.0.14
	github.com/icco/lunchmoney v0.4.1
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/samber/lo v1.50.0
	github.com/spf13/cobra v1.9.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.25.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
)

replace github.com/icco/lunchmoney v0.4.1 => github.com/vpineda1996/lunchmoney v0.5.0
