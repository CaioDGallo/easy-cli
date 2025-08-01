package validation

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/CaioDGallo/easy-cli/internal/types"
)

var (
	clientNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`)
	portRegex       = regexp.MustCompile(`^\d+$`)
)

func ValidateClient(client types.Client) error {
	if err := validateClientName(client.Name); err != nil {
		return fmt.Errorf("invalid client name: %w", err)
	}

	if err := validateSMTPInfo(client.SMTPInfo); err != nil {
		return fmt.Errorf("invalid SMTP configuration: %w", err)
	}

	return nil
}

func validateClientName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("client name cannot be empty")
	}

	if len(name) > 50 {
		return fmt.Errorf("client name cannot exceed 50 characters")
	}

	if !clientNameRegex.MatchString(name) {
		return fmt.Errorf("client name contains invalid characters (only alphanumeric, spaces, hyphens, and underscores allowed)")
	}

	return nil
}

func validateSMTPInfo(smtp types.SMTPInfo) error {
	if strings.TrimSpace(smtp.Server) == "" {
		return fmt.Errorf("SMTP server cannot be empty")
	}

	if !portRegex.MatchString(smtp.Port) {
		return fmt.Errorf("SMTP port must be a valid number")
	}

	if strings.TrimSpace(smtp.Username) == "" {
		return fmt.Errorf("SMTP username cannot be empty")
	}

	if strings.TrimSpace(smtp.Password) == "" {
		return fmt.Errorf("SMTP password cannot be empty")
	}

	if _, err := mail.ParseAddress(smtp.DoNotReplyEmail); err != nil {
		return fmt.Errorf("invalid do-not-reply email address: %w", err)
	}

	if _, err := mail.ParseAddress(smtp.DevEmail); err != nil {
		return fmt.Errorf("invalid dev email address: %w", err)
	}

	return nil
}
