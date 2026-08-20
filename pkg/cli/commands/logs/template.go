package logs

import (
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
)

const (
	timeLayout      = "2006.01.02 15:04:05.000000"
	tagColor        = color.FgHiBlack
	defaultTemplate = `{{ colorHiBlack (toTimestamp .Time "15:04:05") }} {{ colorHiWhite .Caller }} ` +
		`{{ levelColor (printf "%s:" .Level) }} {{ colorTags .Message }}`
	verboseTemplate = `{{ colorHiBlack (toTimestamp .Time "2006.01.02 15:04:05.00") }} {{ colorHiWhite .Caller }} ` +
		`{{ levelColor (printf "%s:" .Level) }}{{ with leadingTags .Message }} {{ . }}{{ end }}` + "\n" +
		`{{ colorWhite (messageBody .Message) }}`
)

var (
	lineRegexp  = regexp.MustCompile(`^(\d{4}\.\d{2}\.\d{2} \d{2}:\d{2}:\d{2}\.\d+) (\[[^\]]*\]) ([A-Z]): (.*)$`)
	tagRegexp   = regexp.MustCompile(`^(\[[^\]]*=[^\]]*\])(\s*)`)
	levelColors = map[string]color.Attribute{
		"D": color.FgHiBlack,
		"I": color.FgHiGreen,
		"W": color.FgHiYellow,
		"E": color.FgRed,
		"P": color.FgHiRed,
		"F": color.FgHiRed,
	}
)

type Log struct {
	Time    time.Time
	Caller  string
	Level   string
	Message string
	Line    string
}

type LogFormatter struct {
	template *template.Template
}

func NewLogFormatter(verbose bool) (*LogFormatter, error) {
	templateText := defaultTemplate
	if verbose {
		templateText = verboseTemplate
	}

	parsed, err := template.New("log").Funcs(templateFuncs()).Parse(templateText)
	if err != nil {
		return nil, errors.Wrap(err, "parse template")
	}

	return &LogFormatter{template: parsed}, nil
}

func (f *LogFormatter) format(line string) string {
	trimmed := strings.TrimSuffix(line, "\n")

	parsedLog, ok := parseLog(trimmed)
	if !ok {
		return line
	}

	_, known := levelColors[parsedLog.Level]
	if !known {
		return line
	}

	var builder strings.Builder

	err := f.template.Execute(&builder, parsedLog)
	if err != nil {
		return line
	}

	suffix := ""
	if strings.HasSuffix(line, "\n") {
		suffix = "\n"
	}

	return builder.String() + suffix
}

func parseLog(line string) (Log, bool) {
	matches := lineRegexp.FindStringSubmatch(line)
	if matches == nil {
		return Log{}, false
	}

	parsedTime, err := time.Parse(timeLayout, matches[1])
	if err != nil {
		return Log{}, false
	}

	return Log{
		Time:    parsedTime,
		Caller:  matches[2],
		Level:   matches[3],
		Message: matches[4],
		Line:    line,
	}, true
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"toTimestamp":  toTimestamp,
		"levelColor":   levelColorFunc,
		"colorTags":    colorTags,
		"leadingTags":  leadingTags,
		"messageBody":  messageBody,
		"colorBlack":   color.BlackString,
		"colorRed":     color.RedString,
		"colorGreen":   color.GreenString,
		"colorYellow":  color.YellowString,
		"colorBlue":    color.BlueString,
		"colorMagenta": color.MagentaString,
		"colorCyan":    color.CyanString,
		"colorWhite":   color.WhiteString,
		"colorHiWhite": color.HiWhiteString,
		"colorHiBlack": color.HiBlackString,
		"colorHiCyan":  color.HiCyanString,
	}
}

func toTimestamp(value any, layout string, timezone ...string) (string, error) {
	parsedTime, err := toTime(value)
	if err != nil {
		return "", errors.Wrap(err, "to time")
	}

	if len(timezone) > 0 {
		var location *time.Location

		location, err = time.LoadLocation(timezone[0])
		if err != nil {
			return "", errors.Wrap(err, "load location")
		}

		parsedTime = parsedTime.In(location)
	}

	return parsedTime.Format(layout), nil
}

func toTime(value any) (time.Time, error) {
	switch typed := value.(type) {
	case time.Time:
		return typed, nil
	case string:
		parsed, err := time.Parse(time.RFC3339Nano, typed)
		if err == nil {
			return parsed, nil
		}

		parsed, err = time.Parse(timeLayout, typed)
		if err != nil {
			return time.Time{}, errors.Wrap(err, "parse time string")
		}

		return parsed, nil
	default:
		return time.Time{}, errors.Newf("unsupported time value %T", value)
	}
}

func levelColorFunc(level string) string {
	if level == "" {
		return level
	}

	attr, ok := levelColors[strings.ToUpper(level[:1])]
	if !ok {
		return level
	}

	return color.New(attr).Sprint(level)
}

func colorTags(message string) string {
	var builder strings.Builder

	rest := message
	for {
		matches := tagRegexp.FindStringSubmatch(rest)
		if matches == nil {
			break
		}

		builder.WriteString(color.New(tagColor).Sprint(matches[1]))
		builder.WriteString(matches[2])
		rest = rest[len(matches[0]):]
	}

	builder.WriteString(color.WhiteString(rest))

	return builder.String()
}

func leadingTags(message string) string {
	var builder strings.Builder

	rest := message
	for {
		matches := tagRegexp.FindStringSubmatch(rest)
		if matches == nil {
			break
		}

		builder.WriteString(color.New(tagColor).Sprint(matches[1]))
		builder.WriteString(matches[2])
		rest = rest[len(matches[0]):]
	}

	return strings.TrimRight(builder.String(), " ")
}

func messageBody(message string) string {
	rest := message
	for {
		matches := tagRegexp.FindStringSubmatch(rest)
		if matches == nil {
			break
		}

		rest = rest[len(matches[0]):]
	}

	return rest
}
