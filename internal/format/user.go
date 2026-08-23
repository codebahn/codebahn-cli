package format

import (
	"encoding/json"
	"fmt"

	"github.com/codebahn/codebahn-cli/internal/output"
)

func init() {
	Register("get_my_user_info", formatUserInfo)
}

func formatUserInfo(raw json.RawMessage, _ any, p *output.Printer) error {
	var user struct {
		Login    string `json:"login"`
		FullName string `json:"full_name"`
		Email    string `json:"email"`
	}
	if err := json.Unmarshal(raw, &user); err != nil {
		return err
	}

	line := fmt.Sprintf("Logged in as %s", user.Login)
	if user.FullName != "" {
		line += fmt.Sprintf(" (%s)", user.FullName)
	}
	p.Text(line)
	if user.Email != "" {
		p.Text(fmt.Sprintf("Email: %s", user.Email))
	}
	return nil
}
