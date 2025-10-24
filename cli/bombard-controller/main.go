package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	hotPink  = lipgloss.Color("#FF00FF")
	darkGray = lipgloss.Color("#767676")
	green    = lipgloss.Color("#00FF00")
	cyan     = lipgloss.Color("#00FFFF")
)

type slider struct {
	label   string
	min     float64
	max     float64
	value   float64
	step    float64
	width   int
	focused bool
}

func newSlider(label string, min, max, initial, step float64) slider {
	return slider{
		label: label,
		min:   min,
		max:   max,
		value: initial,
		step:  step,
		width: 40,
	}
}

func (s *slider) increase() {
	s.value = math.Min(s.max, s.value+s.step)
}

func (s *slider) decrease() {
	s.value = math.Max(s.min, s.value-s.step)
}

func (s slider) view() string {
	// Calculate position
	percent := (s.value - s.min) / (s.max - s.min)
	filled := int(percent * float64(s.width))

	// Build slider bar
	bar := strings.Repeat("━", filled) + "◉" + strings.Repeat("─", s.width-filled)

	// Style based on focus
	labelStyle := lipgloss.NewStyle().Foreground(darkGray).Width(20)
	barStyle := lipgloss.NewStyle().Foreground(darkGray)
	valueStyle := lipgloss.NewStyle().Foreground(darkGray)

	if s.focused {
		labelStyle = labelStyle.Copy().Foreground(hotPink).Bold(true)
		barStyle = barStyle.Copy().Foreground(hotPink)
		valueStyle = valueStyle.Copy().Foreground(hotPink).Bold(true)
	}

	// Format value
	valueStr := fmt.Sprintf("%.2f", s.value)
	if s.step >= 1 {
		valueStr = fmt.Sprintf("%.0f", s.value)
	}

	return fmt.Sprintf("%s %s %s",
		labelStyle.Render(s.label),
		barStyle.Render(bar),
		valueStyle.Render(valueStr),
	)
}

type model struct {
	sliders      []slider
	focused      int
	isBombarding bool
	isPaused     bool
	stats        *stats
	stopChan     chan struct{}
	pauseChan    chan bool
	mu           sync.Mutex
	width        int
	height       int
}
type stats struct {
	totalSent   int
	currentRate float64
	errors      int
	lastError   string
	mu          sync.Mutex
}

func (s *stats) increment() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalSent++
}

func (s *stats) addError(errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errors++
	s.lastError = errMsg
}

func (s *stats) updateRate(rate float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentRate = rate
}

func initialModel() model {
	sliders := []slider{
		newSlider("Concurrency", 10, 200, 100, 10),
		newSlider("Rate (req/s)", 1, 1500, 1000, 10),
		newSlider("Radius", 1, 150, 30, 10),
		newSlider("Transparency", 0.0, 1.0, 1, 0.05),
		newSlider("Time to Grow (ms)", 100, 5000, 400, 50),
	}

	sliders[0].focused = true // Focus on the new concurrency slider

	return model{
		sliders:   sliders,
		focused:   0,
		stats:     &stats{},
		stopChan:  make(chan struct{}),
		pauseChan: make(chan bool, 1), // Buffered channel for pause/resume signals
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.ClearScreen,
		tea.EnterAltScreen,
		m.tickCmd(),
	)
}

func (m model) tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type tickMsg time.Time

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		return m, m.tickCmd()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.isBombarding {
				close(m.stopChan)
			}
			return m, tea.Quit

		case "up", "k":
			if m.focused < len(m.sliders) {
				m.sliders[m.focused].focused = false
			}
			m.focused--
			if m.focused < 0 {
				m.focused = len(m.sliders) - 1 // Wrap around to the last slider
			}
			m.sliders[m.focused].focused = true

		case "down", "j":
			if m.focused < len(m.sliders) {
				m.sliders[m.focused].focused = false
			}
			m.focused++
			if m.focused >= len(m.sliders) {
				m.focused = 0 // Wrap around to the first slider
			}
			m.sliders[m.focused].focused = true
		case "left", "h":
			if m.focused < len(m.sliders) {
				m.sliders[m.focused].decrease()
			}

		case "right", "l":
			if m.focused < len(m.sliders) {
				m.sliders[m.focused].increase()
			}

		case " ", "enter":
			if !m.isBombarding {
				// Start
				m.isBombarding = true
				m.isPaused = false
				m.stats = &stats{}
				m.stopChan = make(chan struct{})
				m.pauseChan = make(chan bool, 1) // Buffered channel for pause/resume signals
				go m.runBombardment()
			} else {
				// Pause/Resume
				m.isPaused = !m.isPaused
				select {
				case m.pauseChan <- m.isPaused:
				default:
				}
			}

		}
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(hotPink).
		Width(m.width).
		Align(lipgloss.Center).
		MarginTop(2).
		MarginBottom(2)

	b.WriteString(titleStyle.Render("🔥 BOMBARDMENT CONTROLLER 🔥"))
	b.WriteString("\n\n")

	// Connection info
	connStyle := lipgloss.NewStyle().
		Foreground(cyan).
		Width(m.width).
		Align(lipgloss.Center)

	b.WriteString(connStyle.Render("Target: localhost:4000"))
	b.WriteString("\n\n\n")

	// Sliders - centered
	for _, s := range m.sliders {
		line := s.view()
		centered := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(line)
		b.WriteString(centered)
		b.WriteString("\n\n")
	}

	b.WriteString("\n")

	// Button
	buttonText := "▶ START"
	buttonColor := green
	if m.isBombarding && !m.isPaused {
		buttonText = "⏸ PAUSE"
		buttonColor = lipgloss.Color("#FFA500")
	} else if m.isBombarding && m.isPaused {
		buttonText = "▶ RESUME"
		buttonColor = green
	}

	buttonStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(buttonColor).
		Foreground(buttonColor).
		Padding(0, 2).
		Bold(true)

	if m.focused == len(m.sliders) {
		buttonStyle = buttonStyle.Copy().
			BorderForeground(hotPink).
			Foreground(hotPink)
	}

	buttonLine := buttonStyle.Render(buttonText)
	centeredButton := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(buttonLine)
	b.WriteString(centeredButton)
	b.WriteString("\n\n")

	// Stats
	if m.isBombarding || m.stats.totalSent > 0 {
		m.stats.mu.Lock()
		statusText := ""
		if m.isPaused {
			statusText = fmt.Sprintf("⏸ PAUSED | Sent: %d | Errors: %d",
				m.stats.totalSent, m.stats.errors)
		} else {
			statusText = fmt.Sprintf("📊 Sent: %d | Rate: %.0f req/s | Errors: %d",
				m.stats.totalSent, m.stats.currentRate, m.stats.errors)
		}
		m.stats.mu.Unlock()

		statsStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Width(m.width).
			Align(lipgloss.Center)

		b.WriteString(statsStyle.Render(statusText))
		b.WriteString("\n")

		b.WriteString("\n")
	}

	// Help text
	helpLines := []string{
		"↑↓/jk: navigate  |  ←→/hl: adjust  |  Space/Enter: start/pause  |  q: quit",
	}
	helpStyle := lipgloss.NewStyle().Foreground(darkGray).
		Width(m.width).
		Align(lipgloss.Center)

	for _, line := range helpLines {
		b.WriteString(helpStyle.Render(line))
		b.WriteString("\n")
	}

	// Center the entire view vertically
	content := b.String()
	contentHeight := strings.Count(content, "\n")
	topPadding := (m.height - contentHeight) / 2
	if topPadding > 0 {
		return strings.Repeat("\n", topPadding) + content
	}

	return content
}

func (m *model) runBombardment() {
	host := "127.0.0.1"
	port := "4000"
	maxCoord := 511
	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        1000, // Even higher
			MaxIdleConnsPerHost: 500,  // Even higher
			MaxConnsPerHost:     500,  // Even higher
		},
	}

	startTime := time.Now()
	step := 0
	lastStatsUpdate := time.Now()
	paused := false

	// Initialize semaphore with the initial concurrency limit
	concurrencyLimit := int(m.sliders[0].value)
	semaphore := make(chan struct{}, concurrencyLimit)

	for {

		// Handle pause/resume signals
		select {
		case newPausedState := <-m.pauseChan:
			paused = newPausedState
			if !paused {
				// m.debugPaused = false // Removed debug-related line
			}
		default:
			// No new pause/resume signal
		}

		// If currently paused, block until a resume signal is received
		if paused {
			// This loop will block until a value is sent to m.pauseChan
			// effectively pausing the bombardment.
			for paused {
				newPausedState := <-m.pauseChan
				paused = newPausedState
				if !paused {
					// m.debugPaused = false // Removed debug-related line
				}
			}
			// Once resumed, continue the main loop
			continue
		}

		select {
		case <-m.stopChan:
			return
		default:
		}

		// Get current slider values and update semaphore if concurrency changed
		newConcurrencyLimit := int(m.sliders[0].value)
		if newConcurrencyLimit != concurrencyLimit {
			// Recreate semaphore with new size. This is a simplified approach;
			// a more robust solution might involve draining and refilling,
			// or using a package like `semaphore`. For this CLI tool,
			// recreating is acceptable as changes are infrequent.
			close(semaphore)
			semaphore = make(chan struct{}, newConcurrencyLimit)
			concurrencyLimit = newConcurrencyLimit
		}

		targetRate := m.sliders[1].value
		radiusValue := m.sliders[2].value // Single radius slider
		transparencyValue := m.sliders[3].value
		timeToGrow := int(m.sliders[4].value)

		interval := time.Duration(float64(time.Second) / targetRate)

		x := rand.Intn(maxCoord + 1)
		y := rand.Intn(maxCoord + 1)

		now := float64(time.Now().UnixMilli())
		color := hslFromParams(now, x, y, maxCoord)
		radius, trans := getDynamicProperties(step, radiusValue, transparencyValue)
		payload := map[string]interface{}{
			"color":        color,
			"radius":       radius,
			"transparency": trans,
			"time_to_grow": timeToGrow,
		}

		// Acquire a token from the semaphore before launching a goroutine
		semaphore <- struct{}{}

		// Capture model reference for closure
		model := m

		go func(x, y int, payload map[string]interface{}) {
			defer func() { <-semaphore }() // Release token when goroutine finishes

			if err := sendRequest(client, host, port, x, y, payload); err != nil {
				errMsg := fmt.Sprintf("POST /api/%d/%d - %v", x, y, err)
				model.stats.addError(errMsg)
			} else {
				model.stats.increment()
			}
		}(x, y, payload)

		step++

		// Update stats display
		if time.Since(lastStatsUpdate) >= 200*time.Millisecond {
			elapsed := time.Since(startTime).Seconds()
			if elapsed > 0 {
				m.stats.updateRate(float64(m.stats.totalSent) / elapsed)
			}
			lastStatsUpdate = time.Now()
		}

		time.Sleep(interval)
	}
}

func hslFromParams(t float64, x, y, maxCoord int) []int {
	hue := int(math.Mod(t*0.05+float64(x)/float64(maxCoord)*360+float64(y)/float64(maxCoord)*180, 360))
	sat := 85 + int(15*math.Sin(t*0.01+float64(x)/float64(maxCoord)*3.14))
	light := 45 + int(20*math.Cos(t*0.008-float64(y)/float64(maxCoord)*2))
	return []int{hue, sat, light}
}

func getDynamicProperties(step int, radiusValue, transparencyValue float64) (float64, float64) {
	radius := radiusValue // Directly use the slider value
	transparency := transparencyValue

	return radius, transparency
}

func sendRequest(client *http.Client, host, port string, x, y int, payload map[string]interface{}) error {
	url := fmt.Sprintf("http://%s:%s/api/%d/%d", host, port, x, y)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshalling JSON for %s: %v\n", url, err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating new request for %s: %v\n", url, err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		// Only print to console if it's not the specific port exhaustion error
		if !strings.Contains(err.Error(), "can't assign requested address") {
			fmt.Printf("Error sending request to %s: %v\n", url, err)
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		// Read response body for more details on error
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			fmt.Printf("Error reading response body for %s (status %d): %v\n", url, resp.StatusCode, readErr)
			return fmt.Errorf("status %d, error reading response body: %v", resp.StatusCode, readErr)
		}
		bodyString := string(bodyBytes)
		fmt.Printf("Server returned error for %s: Status %d, Body: %s\n", url, resp.StatusCode, bodyString)
		return fmt.Errorf("status %d, body: %s", resp.StatusCode, bodyString)
	}

	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
