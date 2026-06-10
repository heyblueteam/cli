package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}

func printWebhook(w Webhook, showSecret bool) {
	fmt.Printf("%s\n", w.Name)
	fmt.Printf("  ID:       %s\n", w.ID)
	fmt.Printf("  UID:      %s\n", w.UID)
	fmt.Printf("  URL:      %s\n", w.URL)
	fmt.Printf("  Status:   %s\n", w.Status)
	fmt.Printf("  Enabled:  %t\n", w.Enabled)
	fmt.Printf("  Events:   %s\n", listOrAll(w.Events))
	fmt.Printf("  Projects: %s\n", listOrAll(w.ProjectIDs))
	if showSecret && w.Secret != "" {
		fmt.Printf("  Secret:   %s\n", w.Secret)
	}
	fmt.Printf("  Created:  %s\n", w.CreatedAt)
	if w.UpdatedAt != "" {
		fmt.Printf("  Updated:  %s\n", w.UpdatedAt)
	}
}

func listOrAll(items []string) string {
	if len(items) == 0 {
		return "all"
	}
	return strings.Join(items, ", ")
}

func printJSON(value interface{}) error {
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func signatureFor(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func signaturesMatch(body []byte, secret, signature string) bool {
	expected := signatureFor(body, secret)
	expectedBytes, err := hex.DecodeString(expected)
	if err != nil {
		return false
	}
	signatureBytes, err := hex.DecodeString(strings.TrimSpace(signature))
	if err != nil {
		return false
	}
	return hmac.Equal(expectedBytes, signatureBytes)
}
