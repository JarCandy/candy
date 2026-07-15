module github.com/CandyCrafts/candy

go 1.25.6

require (
	github.com/k0kubun/pp v2.4.0+incompatible
	github.com/rp1s/colorista v0.0.0-20260708184842-4c40c097e1bc
	github.com/rp1s/digreyt v0.0.0-20260715015800-e45b5ec0f059
)

require (
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rp1s/lipa v0.0.0
	golang.org/x/sys v0.44.0 // indirect
)

replace github.com/rp1s/lipa => ./pkg/lipa
