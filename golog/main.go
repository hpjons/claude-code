// GoLog draws and stores graphic sediment logs. Run it and it opens in the browser.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"golog/internal/server"
	"golog/internal/store"
	"golog/web"
)

const defaultAddr = "127.0.0.1:7341"

func main() {
	addr := flag.String("addr", defaultAddr, "address to listen on (keep it on 127.0.0.1)")
	dataDir := flag.String("data", "", "folder for the GoLog database (default: your user data folder)")
	noBrowser := flag.Bool("no-browser", false, "don't open the browser")
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("GoLog", server.Version)
		return
	}
	if err := run(*addr, *dataDir, !*noBrowser); err != nil {
		fmt.Fprintln(os.Stderr, "GoLog could not start:", err)
		if runtime.GOOS == "windows" {
			// Started by double-clicking, the window would close before the message can be read.
			fmt.Fprintln(os.Stderr, "Press Enter to close this window.")
			bufio.NewReader(os.Stdin).ReadString('\n')
		}
		os.Exit(1)
	}
}

func run(addr, dataDir string, browser bool) error {
	if dataDir == "" {
		dir, err := defaultDataDir()
		if err != nil {
			return err
		}
		dataDir = dir
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	dbPath := filepath.Join(dataDir, "golog.db")
	st, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	defer st.Close()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		// GoLog may already be running: then just show it.
		if addr == defaultAddr && alreadyRunning("http://"+addr) {
			fmt.Println("GoLog is already running at http://" + addr)
			if browser {
				openBrowser("http://" + addr)
			}
			return nil
		}
		if addr != defaultAddr {
			return err
		}
		fmt.Printf("Another program is using %s and does not answer. It may be a GoLog that was paused\n"+
			"with Ctrl+Z instead of stopped with Ctrl+C (in that terminal, type fg and then press Ctrl+C).\n"+
			"Starting on another address for now.\n\n", addr)
		if ln, err = net.Listen("tcp", "127.0.0.1:0"); err != nil {
			return err
		}
	}
	url := "http://" + ln.Addr().String()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	srv := &http.Server{
		Handler:           server.New(st, web.Files(), stop),
		ReadHeaderTimeout: 10 * time.Second,
	}

	fmt.Printf("GoLog %s\n  open:     %s\n  database: %s\n  stop:     Ctrl+C here, or \"Stop GoLog\" on the Projects page\n", server.Version, url, dbPath)
	fmt.Println("Keep this window open while you use GoLog.")
	if browser {
		openBrowser(url)
	}

	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()
	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		fmt.Println("\nGoLog has stopped. Your work is saved.")
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdown)
	}
	return nil
}

// defaultDataDir is ~/.local/share/GoLog on Linux, ~/Library/Application Support/GoLog on
// macOS and %AppData%\GoLog on Windows.
func defaultDataDir() (string, error) {
	if runtime.GOOS == "linux" {
		if x := os.Getenv("XDG_DATA_HOME"); x != "" {
			return filepath.Join(x, "GoLog"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share", "GoLog"), nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "GoLog"), nil
}

func alreadyRunning(url string) bool {
	c := http.Client{Timeout: 2 * time.Second}
	resp, err := c.Get(url + "/api/health")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		fmt.Println("Could not open the browser; open", url, "yourself.")
		return
	}
	go cmd.Wait()
}
