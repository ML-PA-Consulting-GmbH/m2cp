package f

// This module contains helpers in a functional style

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	htmlrenderer "github.com/yuin/goldmark/renderer/html"
)

// MaybeToString converts a pointer value to string, returning fallback if nil
func MaybeToString(value any, fallback string) string {
	if value == nil {
		return fallback
	}

	switch v := value.(type) {
	case *bool:
		if v == nil {
			return fallback
		}
		if *v {
			return "true"
		}
		return "false"
	case *string:
		if v == nil {
			return fallback
		}
		return *v
	case *int:
		if v == nil {
			return fallback
		}
		return fmt.Sprintf("%d", *v)
	case *float64:
		if v == nil {
			return fallback
		}
		return fmt.Sprintf("%f", *v)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func Shorten(str string, maxLen int) string {
	if len(str) > maxLen {
		return str[:maxLen-2] + ".."
	}
	return str
}

func TerminalColor(str string, color string) string {
	const colorReset = "\x1b[0m"
	return color + str + colorReset
}

func TerminalGrey(str string) string {
	return TerminalColor(str, "\x1b[30m")
}

func TerminalYellow(str string) string {
	return TerminalColor(str, "\x1b[33m")
}

func TerminalRed(str string) string {
	return TerminalColor(str, "\x1b[31m")
}

func TerminalCyan(str string) string {
	return TerminalColor(str, "\x1b[36m")
}

func Warning(format string, args ...any) {
	str := fmt.Sprintf(format, args...)
	width := len(str)
	if width > 80 {
		width = 80
	} else if width < 20 {
		width = 20
	}

	// Build a centered ASCII banner like
	// -----[ WARNING ]-----
	label := "[ WARNING ]"
	if len(label) >= width {
		// Fallback: just truncate or pad the label to the desired width
		if len(label) > width {
			label = label[:width]
		}
		fmt.Println(TerminalYellow(label))
		fmt.Println(TerminalYellow(str))
		fmt.Println(TerminalYellow(label))
		return
	}

	// Compute how many dashes we can place around the label.
	// width = left + len(label) + right
	rem := width - len(label)
	left := rem / 2
	right := rem - left
	header := strings.Repeat("─", left) + label + strings.Repeat("─", right)
	footer := strings.Repeat("─", width)

	wrapped := SoftWrap(str, width)
	wrappedLines := strings.Split(wrapped, "\n")

	fmt.Println(TerminalYellow(header))
	for _, wl := range wrappedLines {
		fmt.Println(TerminalYellow(wl))
	}
	fmt.Println(TerminalYellow(footer))
}

func terminalWidth() int {
	const defaultWidth = 80

	// Prefer tput cols when available and working on a real terminal.
	cmd := exec.Command("tput", "cols")
	cmd.Stdin = os.Stdin
	if out, err := cmd.Output(); err == nil {
		trimmed := strings.TrimSpace(string(out))
		if cols, err := strconv.Atoi(trimmed); err == nil && cols > 0 {
			return cols
		}
	}

	// Fallback to COLUMNS environment variable.
	if colsEnv := os.Getenv("COLUMNS"); colsEnv != "" {
		if cols, err := strconv.Atoi(colsEnv); err == nil && cols > 0 {
			return cols
		}
	}

	return defaultWidth
}

// Title prints an emphasized title banner across the console width and returns the title.
// It uses the detected terminal width (tput cols or COLUMNS) when available, otherwise falls back to 80 columns.
func Title(str string) string {
	width := terminalWidth()

	// Ensure the banner is at least slightly wider than the title text.
	minWidth := len(str) + 4
	if width < minWidth {
		width = minWidth
	}

	line := strings.Repeat("─", width)
	var titleLine string = str
	if len(titleLine) > width {
		titleLine = titleLine[:width]
	}

	var buf bytes.Buffer
	buf.WriteString(TerminalCyan(line))
	buf.WriteString("\n")
	buf.WriteString(TerminalCyan(titleLine))
	buf.WriteString("\n")
	buf.WriteString(TerminalCyan(line))
	buf.WriteString("\n")

	return buf.String()
}

// Section prints a compact separator + title line, suitable for marking steps in CLI output.
func Section(format string, args ...any) {
	sep := strings.Repeat("─", 80)
	title := fmt.Sprintf(format, args...)
	fmt.Printf("\n%s\n%s\n%s\n", sep, title, sep)
}

func ToTable(headers []string, data [][]string) string {
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.Header(headers)
	table.Bulk(data)
	table.Render()
	return buf.String()
}

// SecondsToHuman converts seconds (float64) to a human-readable duration string
func SecondsToHuman(seconds float64) string {
	if seconds < 0 {
		return "0s"
	}

	// Handle very small durations
	if seconds < 1 {
		ms := seconds * 1000
		return fmt.Sprintf("%.0fms", ms)
	}

	// Handle seconds only
	if seconds < 60 {
		return fmt.Sprintf("%.1fs", seconds)
	}

	// Handle minutes and seconds
	if seconds < 3600 {
		minutes := int(seconds / 60)
		secs := int(seconds) % 60
		if secs == 0 {
			return fmt.Sprintf("%dm", minutes)
		}
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}

	// Handle hours, minutes, and seconds
	hours := int(seconds / 3600)
	minutes := int(seconds/60) % 60
	secs := int(seconds) % 60

	if minutes == 0 && secs == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	if secs == 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
}

func MarkdownToHTML(markdownContent string) string {
	// Treat empty/whitespace-only content as empty HTML
	if strings.TrimSpace(markdownContent) == "" {
		return ""
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
		goldmark.WithRendererOptions(
			// Allow raw HTML in markdown content if present
			htmlrenderer.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(markdownContent), &buf); err != nil {
		return markdownContent
	}
	return buf.String()
}

func SoftWrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	var currentLine strings.Builder
	for _, w := range words {
		if currentLine.Len() == 0 {
			currentLine.WriteString(w)
			continue
		}
		if currentLine.Len()+1+len(w) > width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(w)
		} else {
			currentLine.WriteString(" ")
			currentLine.WriteString(w)
		}
	}
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

func Confirm(message string) bool {
	fmt.Print(message + " (y/N): ")
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
