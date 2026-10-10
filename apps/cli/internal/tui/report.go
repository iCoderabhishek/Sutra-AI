package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/iCoderabhishek/Sutra-AI/internal/api"
)

// Renders an agent's final answer. Every report has the same layout:
//
//	● COMPLETED  ·  confidence ●●○ medium
//	Headline
//	Summary
//	KEY FINDINGS ──────
//	◆ Title
//	  › point
//	SOURCES ───────────
//	  01  Name · date
//	      url
//	TAKEAWAY ──────────
//	▌ takeaway

func renderReport(r *api.Report, width int) string {
	width = max(width, 30)
	wrap := lipgloss.NewStyle().Width(width)

	parts := []string{
		outcomeBadge(r.Outcome) + styleMuted.Render("  ·  ") + confidenceMeter(r.Confidence),
		"",
		styleReportHeadline.Width(width).Render(sanitize(r.Headline)),
		wrap.Render(styleText.Render(sanitize(r.Summary))),
	}

	if len(r.Findings) > 0 {
		parts = append(parts, "", sectionRule("KEY FINDINGS", width))
		for i, f := range r.Findings {
			if i > 0 {
				parts = append(parts, "")
			}
			parts = append(parts, styleReportFinding.Render("◆ "+sanitize(f.Title)))
			for _, p := range f.Points {
				parts = append(parts, hanging(styleReportBullet.Render("  › "), 4, sanitize(p), width))
			}
		}
	}

	if len(r.Sources) > 0 {
		parts = append(parts, "", sectionRule("SOURCES", width))
		for i, s := range r.Sources {
			line := styleMuted.Render(fmt.Sprintf("  %02d  ", i+1)) + styleText.Render(sanitize(s.Name))
			if s.Date != nil && *s.Date != "" {
				line += styleMuted.Render(" · " + sanitize(*s.Date))
			}
			parts = append(parts, line)
			if s.URL != nil && *s.URL != "" {
				parts = append(parts, styleMuted.Render("      "+truncate(sanitize(*s.URL), width-6)))
			}
		}
	}

	if r.Takeaway != "" {
		parts = append(parts, "", sectionRule("TAKEAWAY", width),
			styleReportTakeaway.Width(width-2).Render(sanitize(r.Takeaway)))
	}

	return strings.Join(parts, "\n")
}

// renderPlainAnswer shows answers saved before the report format existed,
// turning markdown into the same visual language instead of raw ## and **.
func renderPlainAnswer(text string, width int) string {
	width = max(width, 30)
	var out []string
	for _, line := range strings.Split(sanitize(text), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "":
			if len(out) > 0 && out[len(out)-1] != "" {
				out = append(out, "")
			}
		case mdHeading.MatchString(line):
			title := strings.ToUpper(stripInline(mdHeading.ReplaceAllString(line, "")))
			out = append(out, sectionRule(title, width))
		case mdBullet.MatchString(line):
			item := stripInline(mdBullet.ReplaceAllString(line, ""))
			out = append(out, hanging(styleReportBullet.Render("  › "), 4, item, width))
		default:
			out = append(out, lipgloss.NewStyle().Width(width).Render(styleText.Render(stripInline(line))))
		}
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n")
}

// sectionRule draws "LABEL ─────" across the width.
func sectionRule(label string, width int) string {
	label = truncate(label, max(width-4, 4))
	return styleCardLabel.Render(label) + " " + styleMuted.Render(strings.Repeat("─", max(width-lipgloss.Width(label)-1, 2)))
}

// hanging wraps text after a prefix and indents the continuation lines.
func hanging(prefix string, indent int, text string, width int) string {
	body := lipgloss.NewStyle().Width(max(width-indent, 10)).Foreground(colorText).Render(text)
	lines := strings.Split(body, "\n")
	pad := strings.Repeat(" ", indent)
	for i := range lines {
		if i == 0 {
			lines[i] = prefix + lines[i]
		} else {
			lines[i] = pad + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}

func outcomeBadge(outcome string) string {
	switch outcome {
	case "partial":
		return styleStatusPaused.Render("◐ PARTIAL")
	case "declined":
		return lipgloss.NewStyle().Foreground(colorDanger).Render("⊘ DECLINED")
	default:
		return styleStatusActive.Render("● COMPLETED")
	}
}

func confidenceMeter(level string) string {
	filled := map[string]int{"high": 3, "medium": 2, "low": 1}[level]
	if filled == 0 {
		filled, level = 2, "medium"
	}
	return styleMuted.Render("confidence ") +
		styleReportBullet.Render(strings.Repeat("●", filled)) +
		styleMuted.Render(strings.Repeat("○", 3-filled)+" "+level)
}
