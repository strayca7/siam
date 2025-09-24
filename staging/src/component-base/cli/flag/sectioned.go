package flag

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	usageFmt = "Usage:\n  %s\n"
)

// PrintSections prints the given names flag sets in sections, with the maximal given column number.
// If cols is zero, lines are not wrapped. From k8s.io/component-base/cli/flag/sectioned.go
func PrintSections(w io.Writer, fss NamedFlagSets, cols int) {
	for _, name := range fss.Order {
		fs := fss.FlagSets[name]
		if !fs.HasFlags() {
			continue
		}

		wideFS := pflag.NewFlagSet("", pflag.ExitOnError)
		wideFS.AddFlagSet(fs)

		var zzz string
		if cols > 24 {
			zzz = strings.Repeat("z", cols-24)
			wideFS.Int(zzz, 0, strings.Repeat("z", cols-24))
		}

		var buf bytes.Buffer
		fmt.Fprintf(&buf, "\n%s flags:\n\n%s", strings.ToUpper(name[:1])+name[1:], wideFS.FlagUsagesWrapped(cols))

		if cols > 24 {
			i := strings.Index(buf.String(), zzz)
			lines := strings.Split(buf.String()[:i], "\n")
			fmt.Fprint(w, strings.Join(lines[:len(lines)-1], "\n"))
			fmt.Fprintln(w)
		} else {
			fmt.Fprint(w, buf.String())
		}
	}
}

// SetUsageAndHelpFunc sets both uasage and help function for the command.
// Print the flag sets we need instead of all of them. From k8s.io/component-base/cli/flag/sectioned.go
func SetUsageAndHelpFunc(cmd *cobra.Command, fss NamedFlagSets, cols int) {
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
		PrintSections(cmd.OutOrStderr(), fss, cols)
		return nil
	})
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		PrintSections(cmd.OutOrStdout(), fss, cols)
	})
}
