package flag

import "github.com/spf13/pflag"

// AddGlobalFlags adds help flags to the specified FlagSet object.
// basename is the binary name of the application without space, like "apiserver".
// Now it only adds help flag.
func AddGlobalFlags(basename string, fs *pflag.FlagSet) {
	fs.BoolP("help", "h", false, "Help for the "+basename+" command")
}
