package mods

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const modPortalBase = "https://mods.factorio.com"

// builtinMods are shipped with the Factorio server and not downloadable from the mod portal.
var builtinMods = map[string]bool{
	"base":           true,
	"space-age":      true,
	"elevated-rails": true,
	"quality":        true,
}

// Manager handles downloading mods from the Factorio mod portal.
type Manager struct {
	modsDir         string
	modListFile     string
	username        string
	token           string
	factorioVersion string
	httpClient      *http.Client
}

func NewManager(modsDir, modListFile, username, token, factorioVersion string) *Manager {
	return &Manager{
		modsDir:         modsDir,
		modListFile:     modListFile,
		username:        username,
		token:           token,
		factorioVersion: factorioVersion,
		httpClient:      &http.Client{Timeout: 10 * time.Minute},
	}
}

// SyncMods downloads any mods listed in mod-list.json that are not yet present in modsDir.
// Returns (downloaded count, list of failed mod names, fatal error).
// If credentials are not configured, returns (0, nil, nil) and logs a warning.
func (m *Manager) SyncMods(ctx context.Context) (int, []string, error) {
	if m.username == "" || m.token == "" {
		log.Println("mods: FACTORIO_MOD_PORTAL_USER / FACTORIO_MOD_PORTAL_TOKEN не заданы, синхронизация пропущена")
		return 0, nil, nil
	}

	list, err := m.readModList()
	if err != nil {
		return 0, nil, fmt.Errorf("чтение mod-list.json: %w", err)
	}

	downloaded := 0
	var failures []string

	for _, entry := range list.Mods {
		if !entry.Enabled {
			continue
		}
		if builtinMods[entry.Name] {
			continue
		}

		present, err := m.modAlreadyPresent(entry.Name)
		if err != nil {
			return downloaded, failures, fmt.Errorf("сканирование папки модов: %w", err)
		}
		if present {
			continue
		}

		log.Printf("mods: скачиваю %s...", entry.Name)
		if err := m.downloadMod(ctx, entry.Name); err != nil {
			log.Printf("mods: ошибка загрузки %s: %v", entry.Name, err)
			failures = append(failures, entry.Name)
			continue
		}
		log.Printf("mods: %s скачан", entry.Name)
		downloaded++
	}

	return downloaded, failures, nil
}

// ── internal ──────────────────────────────────────────────────────────────────

type modList struct {
	Mods []modListEntry `json:"mods"`
}

type modListEntry struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type modInfo struct {
	Releases []modRelease `json:"releases"`
}

type modRelease struct {
	DownloadURL string      `json:"download_url"`
	FileName    string      `json:"file_name"`
	Version     string      `json:"version"`
	InfoJSON    modInfoJSON `json:"info_json"`
}

type modInfoJSON struct {
	FactorioVersion string `json:"factorio_version"`
}

func (m *Manager) readModList() (*modList, error) {
	data, err := os.ReadFile(m.modListFile)
	if err != nil {
		return nil, err
	}
	var list modList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// modAlreadyPresent checks if any file named "{modName}_*.zip" exists in modsDir.
func (m *Manager) modAlreadyPresent(modName string) (bool, error) {
	entries, err := os.ReadDir(m.modsDir)
	if err != nil {
		return false, err
	}
	prefix := strings.ToLower(modName) + "_"
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".zip") {
			continue
		}
		if strings.HasPrefix(strings.ToLower(e.Name()), prefix) {
			return true, nil
		}
	}
	return false, nil
}

func (m *Manager) downloadMod(ctx context.Context, modName string) error {
	// Fetch mod release list from the portal
	apiURL := fmt.Sprintf("%s/api/mods/%s", modPortalBase, url.PathEscape(modName))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("запрос к mod portal: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mod portal вернул %d для %q", resp.StatusCode, modName)
	}

	var info modInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return fmt.Errorf("декодирование ответа: %w", err)
	}

	release := latestRelease(info.Releases, m.factorioVersion)
	if release == nil {
		return fmt.Errorf("нет релиза для Factorio %s", m.factorioVersion)
	}

	// Build download URL with auth
	downloadURL := fmt.Sprintf("%s%s?username=%s&token=%s",
		modPortalBase,
		release.DownloadURL,
		url.QueryEscape(m.username),
		url.QueryEscape(m.token),
	)

	req2, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}

	resp2, err := m.httpClient.Do(req2)
	if err != nil {
		return fmt.Errorf("скачивание архива: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		return fmt.Errorf("скачивание вернуло %d", resp2.StatusCode)
	}

	dest := filepath.Join(m.modsDir, release.FileName)
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("создание файла: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp2.Body); err != nil {
		os.Remove(dest) // удаляем неполный файл
		return fmt.Errorf("запись файла: %w", err)
	}

	return nil
}

// latestRelease returns the most recent release compatible with the given Factorio major version (e.g. "2.0").
func latestRelease(releases []modRelease, factorioVersion string) *modRelease {
	var best *modRelease
	for i := range releases {
		r := &releases[i]
		if r.InfoJSON.FactorioVersion != factorioVersion {
			continue
		}
		if best == nil || compareVersions(r.Version, best.Version) > 0 {
			best = r
		}
	}
	return best
}

// compareVersions returns positive if a > b (semver comparison, 3 components).
func compareVersions(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		var numA, numB int
		if i < len(partsA) {
			fmt.Sscanf(partsA[i], "%d", &numA) //nolint:errcheck
		}
		if i < len(partsB) {
			fmt.Sscanf(partsB[i], "%d", &numB) //nolint:errcheck
		}
		if numA != numB {
			return numA - numB
		}
	}
	return 0
}
