package api

import (
	"fmt"
	"strconv"
	"strings"

	"examvault/internal/domain"
)

type Request struct {
	Command    string
	ID         string
	Permission string
	Revision   int
	Payload    string
}

func (q Request) Validate() error {
	if strings.TrimSpace(q.Command) == "" {
		return fmt.Errorf("command is required")
	}
	if q.Command != "search" && strings.TrimSpace(q.ID) == "" {
		return fmt.Errorf("id is required")
	}
	if q.Command == "update" && !domain.ValidPermission(q.Permission) {
		return fmt.Errorf("permission is invalid")
	}
	if q.Command == "update" && q.Revision < 1 {
		return fmt.Errorf("revision is invalid")
	}
	return nil
}

func NewRequest(command, id string) Request {
	return Request{Command: strings.ToLower(strings.TrimSpace(command)), ID: strings.TrimSpace(id)}
}

func (q Request) WithPermission(permission string, revision int) Request {
	q.Permission = strings.TrimSpace(permission)
	q.Revision = revision
	return q
}

func Usage() string { return "register review confirm update publish archive search import refresh" }

func IsMutating(command string) bool {
	switch strings.ToLower(strings.TrimSpace(command)) {
	case "register", "review", "confirm", "update", "publish", "archive", "import", "refresh":
		return true
	default:
		return false
	}
}

func IsReadOnly(command string) bool { return strings.EqualFold(strings.TrimSpace(command), "search") }

func RequireArgument(args []string, index int, name string) (string, error) {
	if index < 0 || index >= len(args) || strings.TrimSpace(args[index]) == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return strings.TrimSpace(args[index]), nil
}

func ParseRevision(value string) (int, error) {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("revision is invalid")
	}
	return parsed, nil
}

func (q Request) Args() []string {
	args := []string{q.Command, q.ID}
	if q.Permission != "" {
		args = append(args, q.Permission)
	}
	if q.Revision > 0 {
		args = append(args, strconv.Itoa(q.Revision))
	}
	if q.Payload != "" {
		args = append(args, q.Payload)
	}
	return args
}
