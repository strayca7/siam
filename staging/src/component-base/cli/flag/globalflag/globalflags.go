package globalflag

import "github.com/spf13/pflag"

// AddGlobalFlags adds help flags to the specified FlagSet object.
// name is the binary name of the application without space, like "apiserver".
// Now it only adds help flag.
//
// In Kubernetes, func AddGlobalFlags(fs *pflag.FlagSet, name string, opts ...logs.Option)
func AddGlobalFlags(fs *pflag.FlagSet, name string) {
	fs.BoolP("help", "h", false, "Help for the "+name+" command")
}
