package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"infra/internal/command"
	"infra/internal/executor"
	"infra/internal/k8s"
	"log"
)

var (
	commandExecutor  = command.NewExecutor()
	kubeConfig, _    = k8s.GetKubeConfig()
	ctrlClient       = k8s.NewCtrlClient(kubeConfig)
	etcdAlarm        bool
	etcdHealth       bool
	etcdPerf         bool
	rkeSystemdHealth bool
	prometheusAlarm  bool
	containerdHealth bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "This commands helps you to understand the problem in your cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		executors := make([]executor.Executor, 0)
		if etcdHealth {
			executors = append(executors, executor.NewEtcdHealthExecutor(ctrlClient, commandExecutor))
		}
		if etcdAlarm {
			executors = append(executors, executor.NewEtcdAlarmExecutor(ctrlClient, commandExecutor))
		}
		if etcdPerf {
			executors = append(executors, executor.NewEtcdPerfExecutor(ctrlClient, commandExecutor))
		}
		if rkeSystemdHealth {
			executors = append(executors, executor.NewRkeSystemdHealthExecutor(ctrlClient, commandExecutor))
		}
		if prometheusAlarm {
			executors = append(executors, executor.NewPrometheusAlarmExecutor(ctrlClient, commandExecutor))
		}
		if containerdHealth {
			executors = append(executors, executor.NewContainerdHealthExecutor(ctrlClient, commandExecutor))
		}
		if len(executors) == 0 {
			p := tea.NewProgram(initialModel(&executors))
			if _, err := p.Run(); err != nil {
				log.Fatalln(err)
			}
		}
		executor.NewPool(executors).Run(cmd.Context())
	},
}

func init() {
	doctorCmd.Flags().BoolVarP(&etcdHealth, "etcd-health", "", false, "Check the health of etcd")
	doctorCmd.Flags().BoolVarP(&etcdAlarm, "etcd-alarm", "", false, "Check the alarm of etcd")
	doctorCmd.Flags().BoolVarP(&etcdPerf, "etcd-perf", "", false, "Check the performance of etcd")
	doctorCmd.Flags().BoolVarP(&rkeSystemdHealth, "rke-systemd-health", "", false, "Check the health of rke systemd")
	doctorCmd.Flags().BoolVarP(&prometheusAlarm, "prometheus-alarm", "", false, "Check the alarm of prometheus")
	doctorCmd.Flags().BoolVarP(&containerdHealth, "containerd-health", "", false, "Check the health of containerd")
	rootCmd.AddCommand(doctorCmd)
}

type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
	quitting bool
	execs    *[]executor.Executor
}

func initialModel(execs *[]executor.Executor) model {
	return model{
		choices: []string{
			"Etcd Health",
			"Etcd Alarm",
			"Etcd Perf",
			"Prometheus Alerts",
			"RKE Systemd Health",
			"Containerd Health",
		},
		selected: make(map[int]struct{}),
		execs:    execs,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		case "enter":
			m.quitting = true
			for i := range m.selected {
				switch i {
				case 0:
					*m.execs = append(*m.execs, executor.NewEtcdHealthExecutor(ctrlClient, commandExecutor))
				case 1:
					*m.execs = append(*m.execs, executor.NewEtcdAlarmExecutor(ctrlClient, commandExecutor))
				case 2:
					*m.execs = append(*m.execs, executor.NewEtcdPerfExecutor(ctrlClient, commandExecutor))
				case 3:
					*m.execs = append(*m.execs, executor.NewPrometheusAlarmExecutor(ctrlClient, commandExecutor))
				case 4:
					*m.execs = append(*m.execs, executor.NewRkeSystemdHealthExecutor(ctrlClient, commandExecutor))
				case 5:
					*m.execs = append(*m.execs, executor.NewContainerdHealthExecutor(ctrlClient, commandExecutor))
				}
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Which one do you want to check?\n\n"
	for i, choice := range m.choices {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}
	s += "\nPress ctrl+c or q to quit.\n"
	return s
}
