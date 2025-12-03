// package utils

// import (
// 	"encoding/json"
// 	"io"
// 	"net/http"
// 	"os"
// 	"regexp"
// 	"strconv"
// 	"strings"

// 	"github.com/otiai10/gosseract/v2"
// )

// func extractTextFromImage(imagePath string) (string, error) {
// 	client := gosseract.NewClient()
// 	defer client.Close()

// 	client.SetLanguage("pol") // польский
// 	client.SetImage(imagePath)

// 	return client.Text()
// }

// func parseInvoice(text string) (title string, amount float64) {
// 	// Парсинг суммы (разные форматы)
// 	re := regexp.MustCompile(`(?i)suma|total|razem:?\s*(\d+[,.]?\d*)`)
// 	matches := re.FindStringSubmatch(text)
// 	if len(matches) > 1 {
// 		amountStr := strings.Replace(matches[1], ",", ".", 1)
// 		amount, _ = strconv.ParseFloat(amountStr, 64)
// 	}

// 	// Парсинг title (первая строка или после "Tytuł:")
// 	lines := strings.Split(text, "\n")
// 	if len(lines) > 0 {
// 		title = strings.TrimSpace(lines[0])
// 	}

// 	return
// }

// // HTTP handler
// func uploadInvoice(w http.ResponseWriter, r *http.Request) {
// 	file, _, _ := r.FormFile("invoice")
// 	defer file.Close()

// 	// Сохрани временно
// 	tempPath := "/tmp/invoice.jpg"
// 	out, _ := os.Create(tempPath)
// 	io.Copy(out, file)
// 	out.Close()

// 	// OCR
// 	text, _ := extractTextFromImage(tempPath)
// 	title, amount := parseInvoice(text)

// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"title":    title,
// 		"amount":   amount,
// 		"raw_text": text,
// 	})
// }

package utils