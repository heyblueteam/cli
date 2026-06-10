package webhooks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen for webhook deliveries locally",
	Example: `  blue webhooks listen --port 8080
  blue webhooks listen --port 8080 --secret whsec_123`,
	RunE: runListen,
}

var (
	listenPort   int
	listenPath   string
	listenSecret string
)

func init() {
	listenCmd.Flags().IntVar(&listenPort, "port", 8080, "Port to listen on")
	listenCmd.Flags().StringVar(&listenPath, "path", "/", "Path to listen on")
	listenCmd.Flags().StringVar(&listenSecret, "secret", "", "Optional webhook signing secret for verification")
}

func runListen(cmd *cobra.Command, args []string) error {
	mux := http.NewServeMux()
	mux.HandleFunc(listenPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		fmt.Printf("\n%s %s\n", time.Now().Format(time.RFC3339), r.RemoteAddr)
		if listenSecret != "" {
			signature := r.Header.Get("X-Signature")
			if signaturesMatch(body, listenSecret, signature) {
				fmt.Println("Signature: verified")
			} else {
				fmt.Println("Signature: invalid")
			}
		}

		var formatted interface{}
		if err := json.Unmarshal(body, &formatted); err == nil {
			out, _ := json.MarshalIndent(formatted, "", "  ")
			fmt.Println(string(out))
		} else {
			fmt.Println(string(body))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := fmt.Sprintf(":%d", listenPort)
	fmt.Printf("Listening for Blue webhooks on http://localhost:%d%s\n", listenPort, listenPath)
	return http.ListenAndServe(addr, mux)
}
