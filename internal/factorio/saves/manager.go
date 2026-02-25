package saves

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Manager handles reading and writing Factorio save files
type Manager struct {
	savesDir string
}

func NewManager(savesDir string) *Manager {
	return &Manager{savesDir: savesDir}
}

// LatestSave returns the filename and raw bytes of the most recently modified .zip save
func (m *Manager) LatestSave() (string, []byte, error) {
	entries, err := os.ReadDir(m.savesDir)
	if err != nil {
		return "", nil, fmt.Errorf("reading saves dir: %w", err)
	}

	type saveFile struct {
		name    string
		modTime int64
	}

	var files []saveFile
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".zip" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, saveFile{name: entry.Name(), modTime: info.ModTime().UnixNano()})
	}

	if len(files) == 0 {
		return "", nil, fmt.Errorf("no save files found in %s", m.savesDir)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime > files[j].modTime
	})

	path := filepath.Join(m.savesDir, files[0].name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, fmt.Errorf("reading save file: %w", err)
	}

	return files[0].name, data, nil
}

// CleanAutosaves removes Factorio autosave files (_autosave*.zip).
// Call this before starting the server so the user's uploaded save is picked as the latest.
func (m *Manager) CleanAutosaves() error {
	entries, err := os.ReadDir(m.savesDir)
	if err != nil {
		return fmt.Errorf("reading saves dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "_autosave") && strings.HasSuffix(name, ".zip") {
			path := filepath.Join(m.savesDir, name)
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("removing %s: %w", name, err)
			}
			log.Printf("saves: removed autosave %s", name)
		}
	}
	return nil
}

// Replace removes all existing .zip saves and writes the new one
func (m *Manager) Replace(filename string, data []byte) error {
	entries, err := os.ReadDir(m.savesDir)
	if err != nil {
		return fmt.Errorf("reading saves dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".zip" {
			_ = os.Remove(filepath.Join(m.savesDir, entry.Name()))
		}
	}

	destPath := filepath.Join(m.savesDir, filename)
	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return fmt.Errorf("writing save file: %w", err)
	}

	return nil
}
