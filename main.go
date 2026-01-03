// THe main package contains everything for the fclean program.
// This is currently a simple cleanup program that could also have
// been just as easily written as a shell script. However, I'd like to
// take some free time I've had lately to play around with one of my
// favorite languages (Go) and make something a little different to the
// backend network serving and requesting applications I've used it for in the
// past, so this is my excuse.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

/*----------------- Global Flags----------------------------------------------*/

var (
	dryRun bool
)

/*----------------- Styling --------------------------------------------------*/

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
)

/*----------------- Domain Logic----------------------------------------------*/

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

/*----------------- Helper Functions -----------------------------------------*/

func getRealHomeDir() (string, error) {
	/* If using sudo, $HOME might be /root, we want real users' home */
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

func runCmd(name string, args ...string) error {
	if dryRun {
		return nil
	}
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
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

/*----------------- Check Implementations ------------------------------------*/

/*----------------- Run Implementations --------------------------------------*/

/*----------------- Bubble Tea UI --------------------------------------------*/
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

func (m model) Update(msg bubbletea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
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

/*----------------- Main Execution--------------------------------------------*/

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

		},
	}

}
