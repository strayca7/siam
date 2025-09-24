package flag

import "github.com/spf13/pflag"

type versionValue int

const (
	versionFalse versionValue = iota
	versionTrue
	versionRaw
)

const VersionFlagName = "version"

// AddVersionFlag adds a flag for the version of the program to the specified FlagSet.
func AddVersionFlag(fs *pflag.FlagSet) {
	// TODO: judge the VersionFlagName already exists?
	fs.AddFlag(pflag.Lookup(VersionFlagName))
}
