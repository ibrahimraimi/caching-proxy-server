package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type CacheEntry struct {
	Body       []byte        `json:"body"`
	Headers    http.Header   `json:"headers"`
	StatusCode int           `json:"status_code"`
	Timestamp  time.Time     `json:"timestamp"`
	TTL        time.Duration `json:"ttl"`
}

type Cache struct {
	entries map[string]*CacheEntry
	mutex   sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]*CacheEntry),
	}
}

func (c *Cache) Get(key string) (*CacheEntry, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Since(entry.Timestamp) > entry.TTL {
		delete(c.entries, key)
		return nil, false
	}

	return entry, true
}

func (c *Cache) Set(key string, entry *CacheEntry) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.entries[key] = entry
}

func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

func (c *Cache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.entries)
}

type RequestLog struct {
	Method       string
	Path         string
	Status       string
	CacheHit     bool
	Timestamp    time.Time
	ResponseTime time.Duration
}

type model struct {
	server    *ProxyServer
	port      int
	origin    string
	status    string
	cacheSize int
	requests  []RequestLog
	selected  int
	quitting  bool
	width     int
	height    int
}

func NewModel(port int, origin string) *model {
	return &model{
		port:     port,
		origin:   origin,
		status:   "Starting...",
		requests: make([]RequestLog, 0),
		selected: 0,
		width:    80,
		height:   24,
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		startServer(m.port, m.origin),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.requests)-1 {
				m.selected++
			}
		case "c":
			if m.server != nil {
				m.server.cache.Clear()
				m.cacheSize = 0
			}
		case "r":
			return m, refreshData()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case serverStartedMsg:
		m.server = msg.server
		m.status = "Running"
		return m, tick()
	case serverErrorMsg:
		m.status = fmt.Sprintf("Error: %s", msg.error)
		return m, tea.Quit
	case cacheUpdateMsg:
		m.cacheSize = msg.size
	case requestLogMsg:
		m.requests = append(m.requests, msg.request)
		if len(m.requests) > 100 {
			m.requests = m.requests[1:]
		}
		if m.selected >= len(m.requests) {
			m.selected = len(m.requests) - 1
		}
	case tickMsg:
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return m, nil
}

func (m *model) View() string {
	if m.quitting {
		return "\n  See you later! 👋\n\n"
	}

	var s strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		MarginBottom(1)

	s.WriteString(headerStyle.Render("🚀 Caching Proxy Server"))
	s.WriteString("\n\n")

	statusStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#04B575"))

	s.WriteString(statusStyle.Render("Status: "))
	s.WriteString(m.status)
	s.WriteString("\n")

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	s.WriteString(infoStyle.Render(fmt.Sprintf("Port: %d | Origin: %s | Cache Size: %d entries", m.port, m.origin, m.cacheSize)))
	s.WriteString("\n\n")

	controlsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true)

	s.WriteString(controlsStyle.Render("Controls: "))
	s.WriteString("↑/↓ Navigate | c Clear Cache | r Refresh | q Quit")
	s.WriteString("\n\n")

	if len(m.requests) > 0 {
		s.WriteString(lipgloss.NewStyle().Bold(true).Render("Recent Requests:"))
		s.WriteString("\n")

		header := lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().Width(20).Render("Time"),
			lipgloss.NewStyle().Width(8).Render("Method"),
			lipgloss.NewStyle().Width(30).Render("Path"),
			lipgloss.NewStyle().Width(8).Render("Status"),
			lipgloss.NewStyle().Width(8).Render("Cache"),
			lipgloss.NewStyle().Width(12).Render("Response Time"),
		)
		s.WriteString(headerStyle.Copy().Background(lipgloss.Color("#3C3C3C")).Render(header))
		s.WriteString("\n")

		start := 0
		if len(m.requests) > m.height-15 {
			start = len(m.requests) - (m.height - 15)
		}

		for i := start; i < len(m.requests); i++ {
			req := m.requests[i]
			isSelected := i == m.selected

			rowStyle := lipgloss.NewStyle()
			if isSelected {
				rowStyle = rowStyle.Background(lipgloss.Color("#7D56F4")).Foreground(lipgloss.Color("#FAFAFA"))
			}

			cacheStatus := "MISS"
			cacheColor := lipgloss.Color("#FF6B6B")
			if req.CacheHit {
				cacheStatus = "HIT"
				cacheColor = lipgloss.Color("#04B575")
			}

			row := lipgloss.JoinHorizontal(
				lipgloss.Left,
				lipgloss.NewStyle().Width(20).Render(req.Timestamp.Format("15:04:05")),
				lipgloss.NewStyle().Width(8).Render(req.Method),
				lipgloss.NewStyle().Width(30).Render(truncateString(req.Path, 28)),
				lipgloss.NewStyle().Width(8).Render(req.Status),
				lipgloss.NewStyle().Width(8).Foreground(cacheColor).Render(cacheStatus),
				lipgloss.NewStyle().Width(12).Render(req.ResponseTime.String()),
			)

			s.WriteString(rowStyle.Render(row))
			s.WriteString("\n")
		}
	} else {
		s.WriteString(lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#626262")).Render("No requests yet..."))
		s.WriteString("\n")
	}

	// Footer
	s.WriteString("\n")
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true)

	s.WriteString(footerStyle.Render("Press 'q' to quit"))

	return s.String()
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

type serverStartedMsg struct {
	server *ProxyServer
}

type serverErrorMsg struct {
	error string
}

type cacheUpdateMsg struct {
	size int
}

type requestLogMsg struct {
	request RequestLog
}

type tickMsg time.Time

func startServer(port int, origin string) tea.Cmd {
	return func() tea.Msg {
		proxy, err := NewProxyServer(origin, port)
		if err != nil {
			return serverErrorMsg{error: err.Error()}
		}

		go func() {
			if err := proxy.Start(); err != nil {
				log.Printf("Server error: %v", err)
			}
		}()

		time.Sleep(100 * time.Millisecond)

		return serverStartedMsg{server: proxy}
	}
}

func refreshData() tea.Cmd {
	return func() tea.Msg {
		return tickMsg(time.Now())
	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(time.Now())
	})
}

type ProxyServer struct {
	origin *url.URL
	cache  *Cache
	port   int
	mu     sync.RWMutex
}

func NewProxyServer(originURL string, port int) (*ProxyServer, error) {
	origin, err := url.Parse(originURL)
	if err != nil {
		return nil, fmt.Errorf("invalid origin URL: %v", err)
	}

	return &ProxyServer{
		origin: origin,
		cache:  NewCache(),
		port:   port,
	}, nil
}

func (ps *ProxyServer) generateCacheKey(req *http.Request) string {
	data := fmt.Sprintf("%s:%s:%s", req.Method, req.URL.String(), req.Header.Get("User-Agent"))
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (ps *ProxyServer) handleRequest(w http.ResponseWriter, req *http.Request) {
	start := time.Now()

	cacheKey := ps.generateCacheKey(req)

	if cachedEntry, hit := ps.cache.Get(cacheKey); hit {
		for key, values := range cachedEntry.Headers {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.Header().Set("X-Cache", "HIT")
		w.WriteHeader(cachedEntry.StatusCode)
		w.Write(cachedEntry.Body)
		log.Printf("Cache HIT for %s", req.URL.Path)

		ps.logRequest(req, "200", true, time.Since(start))
		return
	}

	log.Printf("Cache MISS for %s", req.URL.Path)

	ps.logRequest(req, "200", false, time.Since(start))

	proxy := httputil.NewSingleHostReverseProxy(ps.origin)

	req.URL.Host = ps.origin.Host
	req.URL.Scheme = ps.origin.Scheme
	req.Host = ps.origin.Host

	recorder := &responseRecorder{
		ResponseWriter: w,
		statusCode:     200,
		headers:        make(http.Header),
		body:           make([]byte, 0),
	}

	proxy.ServeHTTP(recorder, req)

	if recorder.statusCode >= 200 && recorder.statusCode < 400 {
		cacheEntry := &CacheEntry{
			Body:       recorder.body,
			Headers:    recorder.headers,
			StatusCode: recorder.statusCode,
			Timestamp:  time.Now(),
			TTL:        5 * time.Minute,
		}
		ps.cache.Set(cacheKey, cacheEntry)
		log.Printf("Cached response for %s", req.URL.Path)
	}

	w.Header().Set("X-Cache", "MISS")
}

func (ps *ProxyServer) logRequest(req *http.Request, status string, cacheHit bool, responseTime time.Duration) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	log.Printf("Request: %s %s | Cache: %s | Time: %v",
		req.Method, req.URL.Path,
		map[bool]string{true: "HIT", false: "MISS"}[cacheHit],
		responseTime)
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	headers    http.Header
	body       []byte
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

func (rr *responseRecorder) Write(data []byte) (int, error) {
	rr.body = append(rr.body, data...)
	return rr.ResponseWriter.Write(data)
}

func (rr *responseRecorder) Header() http.Header {
	return rr.headers
}

func (ps *ProxyServer) Start() error {
	http.HandleFunc("/", ps.handleRequest)

	addr := fmt.Sprintf(":%d", ps.port)
	log.Printf("Starting caching proxy server on port %d", ps.port)
	log.Printf("Forwarding requests to: %s", ps.origin.String())
	log.Printf("Cache size: %d entries", ps.cache.Size())

	return http.ListenAndServe(addr, nil)
}

func runTUI(port int, origin string) error {
	p := tea.NewProgram(
		NewModel(port, origin),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		p.Quit()
	}()

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %v", err)
	}

	return nil
}

func main() {
	var (
		port   int
		origin string
		tui    bool
	)

	rootCmd := &cobra.Command{
		Use:   "caching-proxy",
		Short: "A caching proxy server with beautiful TUI",
		Long: `A caching proxy server that forwards requests to origin servers and caches responses.
Features a beautiful terminal user interface (TUI) for monitoring and control.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if origin == "" {
				return fmt.Errorf("origin URL is required")
			}

			if port <= 0 || port > 65535 {
				return fmt.Errorf("port must be between 1 and 65535")
			}

			if tui {
				return runTUI(port, origin)
			} else {
				proxy, err := NewProxyServer(origin, port)
				if err != nil {
					return err
				}

				return proxy.Start()
			}
		},
	}

	clearCmd := &cobra.Command{
		Use:   "clear-cache",
		Short: "Clear the cache",
		Run: func(cmd *cobra.Command, args []string) {
			cache := NewCache()
			cache.Clear()
			fmt.Println("Cache cleared successfully")
		},
	}

	rootCmd.AddCommand(clearCmd)

	rootCmd.Flags().IntVarP(&port, "port", "p", 0, "Port on which the caching proxy server will run")
	rootCmd.Flags().StringVarP(&origin, "origin", "o", "", "URL of the server to which requests will be forwarded")
	rootCmd.Flags().BoolVarP(&tui, "tui", "t", false, "Enable beautiful terminal user interface")

	rootCmd.MarkFlagRequired("port")
	rootCmd.MarkFlagRequired("origin")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
