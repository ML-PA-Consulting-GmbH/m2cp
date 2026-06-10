package model

import (
	"fmt"
	"m2cpcli/format"
	"m2cpcli/graphql"
	"m2cpcli/tools"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"
)

var pushModelCmd = &cobra.Command{
	Use:   "push [modelJsonFilePath uploadMessage]",
	Short: "Push a model to the store",
	Args:  cobra.ExactArgs(2),
	RunE:  runPushModelCmd,
}

// PushModelResult is the structured payload returned by `push`. It
// is rendered as JSON when --json is in effect and as a short human
// summary (plus the full assertion text, when present) otherwise.
type PushModelResult struct {
	Revision       int    `json:"revision" yaml:"revision"`
	ModelAssertion string `json:"modelAssertion,omitempty" yaml:"modelAssertion,omitempty"`
}

func init() {
	modelCmd.AddCommand(pushModelCmd)

	pushModelCmd.Flags().BoolP("update", "u", false,
		"update the model if it already exists "+
			"(errors if content is identical to the current revision)")
	pushModelCmd.Flags().BoolP("idempotent", "i", false,
		"ensure the model is in the requested state with a single "+
			"call; always returns a revision -- whether the model was "+
			"created, updated, or already at the requested content. "+
			"Implies --update.")
	pushModelCmd.Flags().Bool("dry-run", false,
		"scan existing revisions, diff each against the model, and "+
			"report whether it would be reused (already present) or "+
			"pushed as a NEW revision -- without pushing anything.")
}

// duplicateRevisionRe extracts the existing revision number from
// the appstore's duplicate-rejection message, e.g.:
//
//	"Model is a duplicate of model 'foo' revision '1'"
//
// In idempotent mode that error is treated as success (the on-store
// content already matches what we'd have pushed), and the recovered
// revision is returned to the caller.
var duplicateRevisionRe = regexp.MustCompile(`duplicate of model '[^']+' revision '(\d+)'`)

func runPushModelCmd(cmd *cobra.Command, args []string) error {
	modelJsonFilePath := args[0]
	uploadMessage := args[1]

	body, err := tools.ReadLocalFile(modelJsonFilePath)
	if err != nil {
		return err
	}
	modelJson := string(body)

	update, err := cmd.Flags().GetBool("update")
	if err != nil {
		return err
	}
	idempotent, err := cmd.Flags().GetBool("idempotent")
	if err != nil {
		return err
	}
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}
	// -i implies -u: idempotent ensure is a strict superset of update.
	if idempotent {
		update = true
	}

	// The revision scan runs for both --idempotent and --dry-run: it's
	// how we decide whether a matching revision already exists.
	//
	// The appstore's own duplicate check is order-sensitive on the snap
	// list, so a re-push of an already-present model (snaps in a
	// different order) is treated as new and mints a fresh revision.
	// Scanning client-side and comparing semantically works around that.
	//
	// In --idempotent mode it's best-effort: on any error we fall
	// through to the push rather than blocking it. In --dry-run mode any
	// error is fatal (there's nothing to fall through to).
	if idempotent || dryRun {
		want, fpErr := parseToPushModel(modelJson)
		if fpErr != nil {
			if dryRun {
				return fmt.Errorf("dry-run: %w", fpErr)
			}
			fmt.Fprintf(cmd.ErrOrStderr(),
				"warning: idempotent precheck skipped (%v); proceeding to push\n", fpErr)
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(),
				"scanning existing revisions of %q (ascending), diffing each against the model to push...\n", want.Model)
			rev, ok, findErr := findMatchingRevision(cmd.Context(), want, cmd.ErrOrStderr())
			switch {
			case findErr != nil:
				if dryRun {
					return fmt.Errorf("dry-run scan failed: %w", findErr)
				}
				fmt.Fprintf(cmd.ErrOrStderr(),
					"warning: idempotent precheck failed (%v); proceeding to push\n", findErr)
			case ok && dryRun:
				fmt.Fprintf(cmd.ErrOrStderr(),
					"dry-run: model already present as revision %d; no push would occur\n", rev)
				return nil
			case ok:
				fmt.Fprintf(cmd.ErrOrStderr(),
					"idempotent: reusing revision %d (no push)\n", rev)
				return format.PrintFormattedOutput(cmd,
					PushModelResult{Revision: rev},
					customPushFormatter)
			case dryRun:
				fmt.Fprintf(cmd.ErrOrStderr(),
					"dry-run: no existing revision matches; a NEW revision would be pushed\n")
				return nil
			default:
				fmt.Fprintf(cmd.ErrOrStderr(),
					"idempotent: no existing revision matches; pushing a new one\n")
			}
		}
	}

	mutationString := `mutation(
	$modelJson: String!
	$update: Boolean!
	$uploadMessage: String!
) {
	result: pushModelAssertionV2(
		input: {
			modelJson: $modelJson
			uploadMessage: $uploadMessage
			update: $update
		}
	) {
		modelAssertion
		revision
	}
}`

	client, req := graphql.PrepareClientAndRequest(cmd.Context(), mutationString)
	req.Var("modelJson", modelJson)
	req.Var("update", update)
	req.Var("uploadMessage", uploadMessage)

	var pushResult struct {
		Result graphql.PushModelAssertionOutput `json:"result"`
	}
	err = client.Run(cmd.Context(), req, &pushResult)
	if err != nil {
		// In idempotent mode, "duplicate of revision N" means
		// the store already has exactly what we'd push -- nothing
		// to do, return the existing revision.
		if idempotent {
			if m := duplicateRevisionRe.FindStringSubmatch(err.Error()); len(m) == 2 {
				rev, _ := strconv.Atoi(m[1])
				return format.PrintFormattedOutput(cmd,
					PushModelResult{Revision: rev},
					customPushFormatter)
			}
		}
		return err
	}

	return format.PrintFormattedOutput(cmd,
		PushModelResult{
			Revision:       pushResult.Result.Revision,
			ModelAssertion: pushResult.Result.ModelAssertion,
		},
		customPushFormatter)
}

func customPushFormatter(r PushModelResult) (string, error) {
	s := fmt.Sprintf("Pushed model at revision %d", r.Revision)
	if r.ModelAssertion != "" {
		s += ":\n\n" + r.ModelAssertion
	}
	return s, nil
}
