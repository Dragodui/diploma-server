package models

// OCRItem represents a single line item from a receipt
type OCRItem struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity,omitempty"`
	Price    float64 `json:"price"`
}

// OCRResult contains structured data extracted from a receipt
type OCRResult struct {
	Vendor     string    `json:"vendor"`     // Store/company name
	Date       string    `json:"date"`       // Receipt date
	Total      float64   `json:"total"`      // Total amount
	Items      []OCRItem `json:"items"`      // List of items
	RawText    string    `json:"raw_text"`   // Raw text for debugging
	Confidence float64   `json:"confidence"` // Recognition confidence (0-1)
}

// OCRRequest for API request
type OCRRequest struct {
	ImageURL string `json:"image_url" validate:"required,url"`
}
