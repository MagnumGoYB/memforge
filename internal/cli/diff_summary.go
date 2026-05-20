package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	baseconfig "github.com/MagnumGOYB/memforge/internal/config"
	"github.com/spf13/cobra"
)

type diffSummaryPayload struct {
	Files     int               `json:"files"`
	Added     int               `json:"added"`
	Deleted   int               `json:"deleted"`
	Summaries []fileDiffSummary `json:"summaries"`
}

type fileDiffSummary struct {
	Path    string `json:"path"`
	Added   int    `json:"added"`
	Deleted int    `json:"deleted"`
}

func newDiffSummaryCmd(streams Streams) *cobra.Command {
	var fromFile string
	var rootOverride string
	cmd := &cobra.Command{
		Use:   "diff-summary",
		Short: "Summarize a git diff locally",
		RunE: func(cmd *cobra.Command, _ []string) error {
			settings, err := baseconfig.LoadBase(cmd)
			if err != nil {
				return invalidError("%v", err)
			}
			var data []byte
			if strings.TrimSpace(fromFile) != "" {
				data, err = os.ReadFile(fromFile)
				if err != nil {
					return userError("read --from file: %v", err)
				}
			} else {
				data, err = runGitNumstat(rootOverride)
				if err != nil {
					return userError("git diff --numstat: %v", err)
				}
			}
			payload := summarizeNumstat(string(data))
			if settings.Format == baseconfig.FormatJSON {
				return internalError(writeJSON(streams.Stdout, payload))
			}
			_, err = fmt.Fprintf(streams.Stdout, "%d files changed, %d insertions(+), %d deletions(-)\n", payload.Files, payload.Added, payload.Deleted)
			return internalError(err)
		},
	}
	cmd.Flags().StringVar(&fromFile, "from", "", "read git --numstat diff from file")
	cmd.Flags().StringVar(&rootOverride, "root", "", "git root override")
	return cmd
}

func runGitNumstat(root string) ([]byte, error) {
	args := []string{"diff", "--numstat"}
	cmd := exec.Command("git", args...)
	if strings.TrimSpace(root) != "" {
		cmd.Dir = root
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return nil, err
	}
	return out, nil
}

func summarizeNumstat(data string) diffSummaryPayload {
	payload := diffSummaryPayload{Summaries: []fileDiffSummary{}}
	for _, line := range strings.Split(strings.ReplaceAll(data, "\r\n", "\n"), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		added := parseNumstatCount(fields[0])
		deleted := parseNumstatCount(fields[1])
		path := strings.Join(fields[2:], " ")
		payload.Files++
		payload.Added += added
		payload.Deleted += deleted
		payload.Summaries = append(payload.Summaries, fileDiffSummary{Path: path, Added: added, Deleted: deleted})
	}
	sort.SliceStable(payload.Summaries, func(i, j int) bool {
		return payload.Summaries[i].Path < payload.Summaries[j].Path
	})
	return payload
}

func parseNumstatCount(value string) int {
	if value == "-" {
		return 0
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return n
}
