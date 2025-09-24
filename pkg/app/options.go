package app

import (
	cliflag "github.com/strayca7/siam/staging/src/component-base/cli/flag"
)

// CliOptions abstracts configuration options for reading parameters from the command line.
type CliOptions interface {
	// Flags returns a cliflag.NamedFlagSets structure that contains all command line flags grouped by name.
	Flags() (fss cliflag.NamedFlagSets)
	// Validate asks the options to check if there is any error in the configuration.
	Validate() []error
}

// CompleteableOptions is an interface for options that can be completed.
type CompleteableOptions interface {
	Complete() error
}

// PrintableOptions is an interface for options that can be printed.
type PrintableOptions interface {
	String() string
}
