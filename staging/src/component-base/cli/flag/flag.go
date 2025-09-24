package flag

import (
	goflag "flag"
	"strings"
	"sync"

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/strayca7/siam/pkg/logger"
)

var mu sync.Mutex

// NamedFlagSets stores named flag sets in the order of calling FlagSet.
type NamedFlagSets struct {
	// Order is an ordered list of flag set names.
	Order []string
	// FlagSets stores the flag sets by name.
	FlagSets map[string]*pflag.FlagSet
}

// NamedFlagSets.FlagSet returns the *pflag.FlagSet associated with the name.
// If the name does not exist, a new pflag.FlagSet will be created and adds to FlagSets.
func (nfs *NamedFlagSets) FlagSet(name string) *pflag.FlagSet {
	if nfs.FlagSets == nil {
		nfs.FlagSets = map[string]*pflag.FlagSet{}
	}
	if _, ok := nfs.FlagSets[name]; !ok {
		mu.Lock()
		nfs.FlagSets[name] = pflag.NewFlagSet(name, pflag.ExitOnError)
		mu.Unlock()
		// make sure that the name of the FlagSet will only be added to the order after the FlagSet is added to the map
		nfs.Order = append(nfs.Order, name)
	}
	return nfs.FlagSets[name]
}

// WordSepNormalizeFunc changes all flags that contain "_" separators.
func WordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		name = strings.ReplaceAll(name, "_", "-")
	}
	return pflag.NormalizedName(name)
}

func InitFlags(flags *pflag.FlagSet) {
	flags.SetNormalizeFunc(WordSepNormalizeFunc)
	flags.AddGoFlagSet(goflag.CommandLine)
}

// PrintFlags logs the flags in the flagset.
func PrintFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		logger.L().Debug("FLAG", zap.String("--"+flag.Name, flag.Value.String()))
	})
}
