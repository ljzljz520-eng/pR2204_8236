package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"examvault/internal/domain"
	"examvault/internal/flow"
	"examvault/internal/report"
)

type Command struct {
	Name string
	Args []string
}

type Runner struct{ Service *flow.Service }

func ParseArgs(args []string) (Command, error) {
	if len(args) == 0 {
		return Command{}, errors.New("command required")
	}
	name := strings.ToLower(strings.TrimSpace(args[0]))
	allowed := map[string]bool{"register": true, "review": true, "confirm": true, "update": true, "publish": true, "archive": true, "search": true, "import": true, "refresh": true}
	if !allowed[name] {
		return Command{}, fmt.Errorf("unknown command %q", name)
	}
	return Command{Name: name, Args: append([]string(nil), args[1:]...)}, nil
}

func (r Runner) Run(args []string) (string, error) {
	command, err := ParseArgs(args)
	if err != nil {
		return "", err
	}
	switch command.Name {
	case "register":
		return r.register(command.Args)
	case "review":
		return r.review(command.Args)
	case "confirm":
		return r.confirm(command.Args)
	case "update":
		return r.update(command.Args)
	case "publish":
		return r.publish(command.Args)
	case "archive":
		return r.archive(command.Args)
	case "search":
		return r.search(command.Args)
	case "import":
		return r.importRows(command.Args)
	case "refresh":
		return r.refresh(command.Args)
	default:
		return "", errors.New("unsupported command")
	}
}

func (r Runner) register(args []string) (string, error) {
	if len(args) < 5 {
		return "", errors.New("register id title permission payload owner")
	}
	rec, err := r.Service.Register(args[0], args[1], args[2], args[3], args[4])
	if err != nil {
		return "", err
	}
	return FormatResult(rec), nil
}

func (r Runner) review(args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("review id reviewer")
	}
	rec, err := r.Service.Review(args[0], args[1])
	return FormatResult(rec), err
}

func (r Runner) confirm(args []string) (string, error) {
	if len(args) < 1 {
		return "", errors.New("confirm id")
	}
	rec, err := r.Service.Confirm(args[0])
	return FormatResult(rec), err
}

func (r Runner) update(args []string) (string, error) {
	if len(args) < 3 {
		return "", errors.New("update id permission revision")
	}
	revision, err := strconv.Atoi(args[2])
	if err != nil {
		return "", err
	}
	rec, err := r.Service.UpdatePermission(args[0], args[1], revision)
	return FormatResult(rec), err
}

func (r Runner) publish(args []string) (string, error) {
	if len(args) < 1 {
		return "", errors.New("publish id")
	}
	rec, err := r.Service.Publish(args[0])
	return FormatResult(rec), err
}

func (r Runner) archive(args []string) (string, error) {
	if len(args) < 1 {
		return "", errors.New("archive id")
	}
	rec, err := r.Service.Archive(args[0])
	return FormatResult(rec), err
}

func (r Runner) search(args []string) (string, error) {
	query := ""
	if len(args) > 0 {
		query = args[0]
	}
	records, err := r.Service.Search(flow.SearchOptions{Query: query})
	if err != nil {
		return "", err
	}
	return report.EncodeSearch(report.SearchSummary(query, records)), nil
}

func (r Runner) importRows(args []string) (string, error) {
	if len(args) < 1 {
		return "", errors.New("import rows")
	}
	rows := flow.ParseRows(args[0])
	result := r.Service.Import(rows)
	return report.EncodeImport(report.ImportSummary(result, r.Service.ImportDigest(rows))), nil
}

func (r Runner) refresh(args []string) (string, error) {
	if len(args) < 1 {
		return "", errors.New("refresh id")
	}
	rec, err := r.Service.Refresh(args[0])
	return FormatResult(rec), err
}

func FormatResult(record domain.Record) string {
	return fmt.Sprintf("%s|%s|%s|%s|r%d", record.ID, record.Title, record.Permission, record.Status, record.Revision)
}

func ParseKeyValues(args []string) map[string]string {
	values := make(map[string]string)
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" {
			values[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return values
}

func CommandFromRequest(request Request) (Command, error) {
	if err := request.Validate(); err != nil {
		return Command{}, err
	}
	args := make([]string, 0, 4)
	if request.ID != "" {
		args = append(args, request.ID)
	}
	if request.Permission != "" {
		args = append(args, request.Permission)
	}
	if request.Revision > 0 {
		args = append(args, strconv.Itoa(request.Revision))
	}
	if request.Payload != "" {
		args = append(args, request.Payload)
	}
	return Command{Name: request.Command, Args: args}, nil
}

func ParseRequest(args []string) (Request, error) {
	command, err := ParseArgs(args)
	if err != nil {
		return Request{}, err
	}
	request := NewRequest(command.Name, "")
	if len(command.Args) > 0 {
		request.ID = command.Args[0]
	}
	if command.Name == "update" {
		if len(command.Args) < 3 {
			return Request{}, errors.New("update requires id, permission, revision")
		}
		request.Permission = command.Args[1]
		request.Revision, err = strconv.Atoi(command.Args[2])
		if err != nil {
			return Request{}, err
		}
	}
	return request, nil
}

func FormatRecords(records []domain.Record) []string {
	lines := make([]string, 0, len(records))
	for _, record := range records {
		lines = append(lines, FormatResult(record))
	}
	return lines
}

func FormatError(err error) string {
	if err == nil {
		return ""
	}
	return "error: " + err.Error()
}

func FormatUsageError(command string) string {
	if strings.TrimSpace(command) == "" {
		return "command required; use: " + Usage()
	}
	return "unknown command " + command + "; use: " + Usage()
}

func ValidateArguments(command Command) error {
	switch command.Name {
	case "register":
		if len(command.Args) != 5 {
			return errors.New("register expects five arguments")
		}
	case "review":
		if len(command.Args) != 2 {
			return errors.New("review expects two arguments")
		}
	case "update":
		if len(command.Args) != 3 {
			return errors.New("update expects three arguments")
		}
	case "search", "confirm", "publish", "archive", "refresh":
		if len(command.Args) < 1 {
			return errors.New("command expects an id")
		}
	case "import":
		if len(command.Args) != 1 {
			return errors.New("import expects rows")
		}
	default:
		return errors.New("unsupported command")
	}
	return nil
}
