package settings

import (
	"encoding/json"
	"fmt"
	"os"
)

// UpdatePasswords writes the generated password into server-settings.json:
//   - game_password  — пароль для входа игроков на сервер
//   - rcon_password  — пароль RCON (дублируем для надёжности, помимо rconpw-файла)
func UpdatePasswords(settingsFile, password string) error {
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		return fmt.Errorf("reading server settings: %w", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("parsing server settings: %w", err)
	}

	m["game_password"] = password
	m["rcon_password"] = password

	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling server settings: %w", err)
	}

	if err := os.WriteFile(settingsFile, out, 0644); err != nil {
		return fmt.Errorf("writing server settings: %w", err)
	}

	return nil
}
