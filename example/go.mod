module example

go 1.24.1

require (
	github.com/oomol-lab/ssh-forward v0.0.0-20250310111812-cc1e7c2d1876
	github.com/sirupsen/logrus v1.9.3
)

require (
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
)

replace github.com/oomol-lab/ssh-forward v0.0.0-20250310111812-cc1e7c2d1876 => ../
