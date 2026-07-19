package cli

import (
	stderrors "errors"
	"fmt"
	"io"

	"github.com/caramelang/caramel/pkg/highlight"
	diagnostics "github.com/rp1s/digreyt"
)

// PrintError writes every error in err, rendering diagnostics with source context.
func PrintError(w io.Writer, err error) error {
	if err == nil {
		return nil
	}

	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		writeErrors := make([]error, 0, len(joined.Unwrap()))
		for _, nested := range joined.Unwrap() {
			writeErrors = append(writeErrors, PrintError(w, nested))
		}
		return stderrors.Join(writeErrors...)
	}

	var arena *diagnostics.Arena
	if stderrors.As(err, &arena) {
		return arena.PrintWith(w, highlight.NewDiagnosticRenderer())
	}

	_, writeErr := fmt.Fprintln(w, err)
	return writeErr
}
