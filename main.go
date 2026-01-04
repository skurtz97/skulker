// Package main contains all of the code for the `skulker` program. The skulker
// program is a program that cleans up system files and caches as well as the
// user home directory using a variety of customizable cleanup processes. This
// is all presented via a TUI interface provided by bubbletea.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var (
	dryRun bool
)

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	warn      = lipgloss.AdaptiveColor{Light: "#F25D94", Dark: "#FF5F87"}

	titleStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true).
			MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	specialStyle = lipgloss.NewStyle().Foreground(special)
	warnStyle    = lipgloss.NewStyle().Foreground(warn)
	subtleStyle  = lipgloss.NewStyle().Foreground(subtle)
)

type TaskID int

const (
	TaskDNFCache TaskID = iota
	TaskDNFAutoRemove
	TaskFlatpak
	TaskJournal
	TaskUserCache
	TaskOldKernels
)

type CleanerTask struct {
	ID          TaskID
	Name        string
	Description string
	Check       func() (int64, string, error)
	Run         func() error
	Selected    bool
	ScanResult  string
	Size        int64
}

var tasks = []*CleanerTask{
	{
		ID:          TaskDNFCache,
		Name:        "DNF Metadata & Cache",
		Description: "Removes cached package data. Safe to run.",
		Check:       checkDNFCache,
		Run:         runDNFCache,
	},
	{
		ID:          TaskDNFAutoRemove,
		Name:        "DNF Autoremove",
		Description: "Removes orphaned dependencies.",
		Check:       checkDNFAutoRemove,
		Run:         runDNFAutoRemove,
	},
	{
		ID:          TaskFlatpak,
		Name:        "Unused Flatpak Runtimes",
		Description: "Removes runtimes not used by any installed app.",
		Check:       checkFlatpak,
		Run:         runFlatpak,
	},
	{
		ID:          TaskJournal,
		Name:        "Systemd Journal",
		Description: "Vacuums system logs older than 7 days.",
		Check:       checkJournal,
		Run:         runJournal,
	},
	{
		ID:          TaskOldKernels,
		Name:        "Old Linux Kernels",
		Description: "Removes all but the running + 1 latest kernel.",
		Check:       checkKernels,
		Run:         runKernels,
	},
	{
		ID:          TaskUserCache,
		Name:        "User Cache (~/.cache)",
		Description: "Cleans temporary user application caches.",
		Check:       checkUserCache,
		Run:         runUserCache,
	},
}

//-----------------------------------------------------------------------------
// Helper Implementations
//-----------------------------------------------------------------------------

// func getRealHomeDir finds the real home directory of the user running
// the skulker program. We need a special function to do this because skulker
// expects to be run as root via `sudo` or similar, which means that a naiive
// check of $HOME might return /root instead of the underlying users actual
// home directory. TO get the correct home directory, we check the value of
// SUDO_USER and get the home directory of that user. If SUDO_USER is empty
// or undefined, then we get the current user (the effective user executing the
// program) home directory instead.
func getRealHomeDir() (string, error) {
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser != "" {
		u, err := user.Lookup(sudoUser)
		if err != nil {
			return "", err
		}
		return u.HomeDir, nil
	}
	return os.UserHomeDir()
}

// func dirSize calculates the total size in bytes of the directory at the
// specified path.
//
// It recursively traverses the directory tree using filepath.WalkDir.
//
// Note: This implementation suppresses permission errors or file access errors
// encountered during traversal, continuing the walk and returning the size of
// accessible files only.
func dirSize(path string) (int64, error) {
	var size int64
	// WalkDir is more efficient than Walk as it avoids calling os.Lstat on
	// every node.
	err := filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		// If an error occurs (e.g., permission denied), return nil to ignore
		// it and continue walking the rest of the tree.
		if err != nil {
			return nil
		}
		// If the entry is a file (not a directory), add its size to the total.
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				size += info.Size()
			}
		}
		return nil
	})
	return size, err
}

// func runCmd is just a simple wrapper around exec.Command which runs
// shell commands.
func runCmd(name string, args ...string) error {
	if dryRun {
		return nil
	}
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// ------------------------------------------------------------------------------
// Check Implementations
// ------------------------------------------------------------------------------
func checkDNFCache() (int64, string, error) {
	// Fedora typically stores cache in /var/cache/dnf or /var/cache/libdnf5
	size, err := dirSize("/var/cache/libdnf5")
	if err != nil || size == 0 {
		size, _ = dirSize("/var/cache/dnf")
	}
	return size, humanize.Bytes(uint64(size)), nil
}

func checkDNFAutoRemove() (int64, string, error) {
	return 0, "Checks for orphans", nil
}

func checkFlatpak() (int64, string, error) {
	_, err := exec.LookPath("flatpak")
	if err != nil {
		return 0, "Not installed", nil
	}
	return 0, "Removes unused runtimes", nil
}

func checkJournal() (int64, string, error) {
	out, err := exec.Command("journalctl", "--disk-usage").Output()
	if err != nil {
		return 0, "Unknown", nil
	}
	s := string(out)
	parts := strings.Fields(s)
	for _, p := range parts {
		if strings.Contains(p, "M") || strings.Contains(p, "G") || strings.Contains(p, "B") {
			return 0, p, nil
		}
	}
	return 0, "Unknown", nil
}

func checkKernels() (int64, string, error) {
	out, err := exec.Command("rpm", "-q", "kernel-core").Output()
	if err != nil {
		return 0, "Unknown", nil
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	count := 0
	for _, l := range lines {
		if l != "" {
			count++
		}
	}
	if count > 2 {
		return 0, fmt.Sprintf("%d kernels installed", count), nil
	}
	return 0, fmt.Sprintf("%d kernels (Safe)", count), nil
}

func checkUserCache() (int64, string, error) {
	home, err := getRealHomeDir()
	if err != nil {
		return 0, "", err
	}
	cachePath := filepath.Join(home, ".cache")
	size, err := dirSize(cachePath)
	return size, humanize.Bytes(uint64(size)), nil
}

func runDNFCache() error {
	return runCmd("dnf", "clean", "all")
}

func runDNFAutoRemove() error {
	return runCmd("dnf", "autoremove", "-y")
}

func runFlatpak() error {
	return runCmd("flatpak", "uninstall", "--unused", "-y")
}

func runJournal() error {
	return runCmd("journalctl", "--vacuum-time=7d")
}

func runKernels() error {
	script := "dnf remove $(dnf repoquery --installonly --latest-limit=-2 -q) -y"
	if dryRun {
		return nil
	}
	return exec.Command("bash", "-c", script).Run()
}

func runUserCache() error {
	home, err := getRealHomeDir()
	if err != nil {
		return err
	}
	cachePath := filepath.Join(home, ".cache")

	if dryRun {
		return nil
	}

	entries, err := os.ReadDir(cachePath)
	if err != nil {
		return err
	}

	for _, e := range entries {
		p := filepath.Join(cachePath, e.Name())
		os.RemoveAll(p)
	}
	return nil
}

type appState int

const (
	stateScanning appState = iota
	stateSelecting
	stateCleaning
	stateDone
)

type model struct {
	state       appState
	spinner     spinner.Model
	scanIndex   int
	cleanIndex  int
	form        *huh.Form
	selectedIDs []TaskID
	width       int
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(special)
	selectedIDs := []TaskID{}
	return model{
		state:       stateScanning,
		spinner:     s,
		scanIndex:   0,
		selectedIDs: selectedIDs,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, scanNextCmd(0))
}

type scanMsg struct {
	index int
	size  int64
	info  string
}

type cleanMsg struct {
	index int
	err   error
}

func scanNextCmd(index int) tea.Cmd {
	return func() tea.Msg {
		if index >= len(tasks) {
			return nil
		}
		t := tasks[index]
		size, info, _ := t.Check()
		return scanMsg{index, size, info}
	}
}

func cleanNextCmd(index int) tea.Cmd {
	return func() tea.Msg {
		// Simulate delay for visual clarity
		time.Sleep(500 * time.Millisecond)

		//Find next selected task starting from index
		for i := index; i < len(tasks); i++ {
			if tasks[i].Selected {
				err := tasks[i].Run()
				return cleanMsg{i, err}
			}
		}
		return cleanMsg{-1, nil}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width

	// SCANNING PHASE
	case scanMsg:
		tasks[msg.index].Size = msg.size
		tasks[msg.index].ScanResult = msg.info

		m.scanIndex++
		if m.scanIndex < len(tasks) {
			return m, scanNextCmd(m.scanIndex)
		}

		// Scanning finished, switch to Selection state
		m.state = stateSelecting

		// Pre-select all by default
		for _, t := range tasks {
			m.selectedIDs = append(m.selectedIDs, t.ID)
		}

		// Build Huh options
		opts := []huh.Option[TaskID]{}
		for _, t := range tasks {
			label := fmt.Sprintf("%s (%s)", t.Name, t.ScanResult)
			opts = append(opts, huh.NewOption(label, t.ID))
		}

		m.form = huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[TaskID]().
					Title("Select Items to Clean").
					Description("Press to toggle, [Enter] to confirm").
					Options(opts...).
					Value(&m.selectedIDs),
			),
		).WithTheme(huh.ThemeDracula())

		return m, m.form.Init()

	// SPINNER TICK
	case spinner.TickMsg:
		if m.state == stateScanning || m.state == stateCleaning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	// CLEANING PHASE
	case cleanMsg:
		if msg.index == -1 {
			m.state = stateDone
			return m, tea.Quit
		}
		m.cleanIndex = msg.index + 1
		return m, cleanNextCmd(m.cleanIndex)
	}

	// Handle Form Updates
	if m.state == stateSelecting {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
			if m.form.State == huh.StateCompleted {
				// Map selectedIDs back to tasks
				for _, t := range tasks {
					t.Selected = false
					for _, id := range m.selectedIDs {
						if t.ID == id {
							t.Selected = true
							break
						}
					}
				}
				m.state = stateCleaning
				m.cleanIndex = 0
				return m, tea.Batch(m.spinner.Tick, cleanNextCmd(0))
			}
		}
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	s := strings.Builder{}

	s.WriteString(titleStyle.Render("Skulker"))
	s.WriteString(("\n\n"))

	if dryRun {
		s.WriteString(warnStyle.Render("DRY RUN MODE ENABLED"))
		s.WriteString("\n\n")
	}

	switch m.state {
	case stateScanning:
		s.WriteString((fmt.Sprintf("%s Scanning system resources...(%d/%d)\n\n", m.spinner.View(), m.scanIndex+1, len(tasks))))
		for i, t := range tasks {
			check := " "
			if i < m.scanIndex {
				check = "v"
			}
			s.WriteString(subtleStyle.Render(fmt.Sprintf(" %s %s", check, t.Name)) + "\n")
		}
	case stateSelecting:
		s.WriteString(m.form.View())

	case stateCleaning:
		s.WriteString("Cleaning in progress...\n\n")
		for i, t := range tasks {
			if !t.Selected {
				continue
			}

			if i < m.cleanIndex {
				s.WriteString(specialStyle.Render(" ✓ "+t.Name) + "\n")
				s.WriteString("\n")
			} else if i == m.cleanIndex {
				s.WriteString(fmt.Sprintf(" %s %s", m.spinner.View(), t.Name))
				s.WriteString("\n")
			} else {
				s.WriteString(lipgloss.NewStyle().Faint(true).Render(" - " + t.Name))
				s.WriteString("\n")
			}
		}

	case stateDone:
		s.WriteString(specialStyle.Render("All operations completed successfully!") + "\n")
	}

	return itemStyle.Render(s.String())

}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "fclean",
		Short: "Interactive Fedora System Cleaner",
		Long: `An interactive system clean for performing routine maintenence
		on a Fedora system. This includes cleaning up package caches, orphans,
		trimming the journal, removing old kernels, and doing some (careful) 
		maintenence of the user home directory.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Sudo check
			if os.Geteuid() != 0 && !dryRun {
				fmt.Println(lipgloss.NewStyle().Foreground(warn).Render("Error: This tool requires root priviliges to clean system files"))
				fmt.Println("Please run with sudo.")
				os.Exit(1)
			}

			p := tea.NewProgram(initialModel())
			if _, err := p.Run(); err != nil {
				fmt.Printf("Alas, there's been an error: %v", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Simulate actions without deleting")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
