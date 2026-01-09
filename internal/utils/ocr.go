package utils

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/Dragodui/diploma-server/internal/models"

	"github.com/otiai10/gosseract/v2"
)

// ExtractTextFromImage extracts text from an image using Tesseract OCR
func ExtractTextFromImage(imagePath string, languages ...string) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	// Default languages: English, Russian, Ukrainian, Polish, Belarusian
	langs := "eng+rus+ukr+pol+bel"
	if len(languages) > 0 {
		langs = strings.Join(languages, "+")
	}
	client.SetLanguage(langs)
	client.SetImage(imagePath)

	return client.Text()
}

// ParseReceipt parses receipt text and extracts structured data
func ParseReceipt(text string) *models.OCRResult {
	result := &models.OCRResult{
		RawText: text,
		Items:   []models.OCRItem{},
	}

	lines := strings.Split(text, "\n")

	result.Vendor = extractVendor(lines)
	result.Date = extractDate(text)
	result.Total = extractTotal(text)
	result.Items = extractItems(lines)
	result.Confidence = calculateConfidence(result)

	return result
}

// extractVendor extracts store name from the first lines of receipt
func extractVendor(lines []string) string {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 3 && len(trimmed) < 100 {
			// Skip lines with dates and numbers only
			if !isDateLine(trimmed) && !isOnlyNumbers(trimmed) {
				return trimmed
			}
		}
	}
	return ""
}

// extractDate extracts date from receipt text
func extractDate(text string) string {
	// Various date formats
	datePatterns := []string{
		`(\d{2}[./]\d{2}[./]\d{4})`,                  // DD.MM.YYYY or DD/MM/YYYY
		`(\d{4}[.-]\d{2}[.-]\d{2})`,                  // YYYY-MM-DD or YYYY.MM.DD
		`(\d{2}[./]\d{2}[./]\d{2})`,                  // DD.MM.YY or DD/MM/YY
		`(\d{2}\s+\w+\s+\d{4})`,                      // DD Month YYYY
		`(?i)date[:\s]*(\d{2}[./]\d{2}[./]\d{2,4})`, // Date: DD.MM.YYYY
		`(?i)дата[:\s]*(\d{2}[./]\d{2}[./]\d{2,4})`, // Date (Russian): DD.MM.YYYY
	}

	for _, pattern := range datePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// extractTotal extracts total amount from receipt text
func extractTotal(text string) float64 {
	// Patterns for total amount (multiple languages)
	totalPatterns := []string{
		`(?i)(?:total|итого|всего|suma|razem|сума|загалом|do\s+zapłaty)[:\s]*[=]?\s*(\d+[.,]?\d*)`,
		`(?i)(?:к\s+оплате|до\s+сплати|amount)[:\s]*[=]?\s*(\d+[.,]?\d*)`,
		`(?i)(?:сумма|kwota)[:\s]*[=]?\s*(\d+[.,]?\d*)`,
		`(?i)(?:grand\s+total)[:\s]*[=]?\s*(\d+[.,]?\d*)`,
	}

	var maxAmount float64

	for _, pattern := range totalPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				amountStr := strings.Replace(match[1], ",", ".", 1)
				amountStr = strings.TrimSpace(amountStr)
				if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
					if amount > maxAmount {
						maxAmount = amount
					}
				}
			}
		}
	}

	return maxAmount
}

// extractItems extracts line items from receipt lines
func extractItems(lines []string) []models.OCRItem {
	var items []models.OCRItem

	// Patterns for items: name ... price or name quantity x price
	itemPatterns := []*regexp.Regexp{
		// Name ... quantity x price = total
		regexp.MustCompile(`^(.+?)\s+(\d+(?:[.,]\d+)?)\s*[xх*]\s*(\d+[.,]?\d*)\s*[=]?\s*(\d+[.,]?\d*)$`),
		// Name ... price (at end of line)
		regexp.MustCompile(`^(.+?)\s{2,}(\d+[.,]\d{2})$`),
		// Name price (no separator)
		regexp.MustCompile(`^([A-Za-zА-Яа-яЁёІіЇїЄє\s\-]+)\s+(\d+[.,]\d{2})$`),
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) < 5 {
			continue
		}

		// Skip service lines
		if isServiceLine(trimmed) {
			continue
		}

		for _, pattern := range itemPatterns {
			matches := pattern.FindStringSubmatch(trimmed)
			if len(matches) >= 3 {
				name := strings.TrimSpace(matches[1])
				if len(name) < 2 {
					continue
				}

				item := models.OCRItem{Name: name}

				if len(matches) >= 5 {
					// Format: name quantity x price = total
					if qty, err := parseFloat(matches[2]); err == nil {
						item.Quantity = qty
					}
					if price, err := parseFloat(matches[4]); err == nil {
						item.Price = price
					}
				} else if len(matches) >= 3 {
					// Format: name price
					if price, err := parseFloat(matches[2]); err == nil {
						item.Price = price
						item.Quantity = 1
					}
				}

				if item.Price > 0 {
					items = append(items, item)
				}
				break
			}
		}
	}

	return items
}

// calculateConfidence calculates OCR result confidence score
func calculateConfidence(result *models.OCRResult) float64 {
	score := 0.0
	checks := 0.0

	// Check vendor presence
	checks++
	if result.Vendor != "" {
		score++
	}

	// Check date presence
	checks++
	if result.Date != "" {
		score++
	}

	// Check total presence
	checks++
	if result.Total > 0 {
		score++
	}

	// Check items presence
	checks++
	if len(result.Items) > 0 {
		score++
	}

	// Check if items total matches overall total
	if len(result.Items) > 0 && result.Total > 0 {
		checks++
		itemsTotal := 0.0
		for _, item := range result.Items {
			itemsTotal += item.Price
		}
		// Allow 10% tolerance
		if itemsTotal > 0 && abs(itemsTotal-result.Total)/result.Total < 0.1 {
			score++
		}
	}

	if checks == 0 {
		return 0
	}
	return score / checks
}

// Helper functions

func isDateLine(s string) bool {
	re := regexp.MustCompile(`^\d{2}[./]\d{2}[./]\d{2,4}`)
	return re.MatchString(s)
}

func isOnlyNumbers(s string) bool {
	re := regexp.MustCompile(`^[\d\s.,]+$`)
	return re.MatchString(s)
}

func isServiceLine(s string) bool {
	lower := strings.ToLower(s)
	serviceWords := []string{
		"total", "итого", "всего", "suma", "razem", "сума",
		"cash", "card", "наличн", "картой", "сдача", "change",
		"tax", "vat", "ндс", "pdv", "thank", "спасибо", "дякуємо",
		"receipt", "чек", "fiscal", "фіскальний",
	}
	for _, word := range serviceWords {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return false
}

func parseFloat(s string) (float64, error) {
	s = strings.Replace(s, ",", ".", 1)
	s = strings.TrimSpace(s)
	return strconv.ParseFloat(s, 64)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
