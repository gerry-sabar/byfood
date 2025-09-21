package app

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/gerry-sabar/byfood/internal/ports"
)

type ValidationError struct {
	Fields map[string]string `json:"fields"`
}

func (v *ValidationError) Error() string { return "validation error" }
func (v *ValidationError) add(field, msg string) {
	if v.Fields == nil {
		v.Fields = map[string]string{}
	}
	// keep first error per field (simple UX)
	if _, exists := v.Fields[field]; !exists {
		v.Fields[field] = msg
	}
}
func (v *ValidationError) ok() bool { return len(v.Fields) == 0 }

var (
	reIsbn10 = regexp.MustCompile(`^\d{9}[\dX]$`)
	reIsbn13 = regexp.MustCompile(`^\d{13}$`)
)

// ---- Year validation ----
const (
	minYear = 1000
	maxYear = 9999
)

func isValidFourDigitYearInt(y int) bool {
	return y >= minYear && y <= maxYear
}

func normalizeISBN(s string) string {
	// remove spaces/hyphens, uppercase X
	return strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "-", ""))
}

func isValidISBN10(s string) bool {
	if !reIsbn10.MatchString(s) {
		return false
	}
	sum := 0
	for i := 0; i < 9; i++ {
		sum += (10 - i) * int(s[i]-'0')
	}
	last := s[9]
	if last == 'X' {
		sum += 10
	} else {
		sum += int(last - '0')
	}
	return sum%11 == 0
}

func isValidISBN13(s string) bool {
	if !reIsbn13.MatchString(s) {
		return false
	}
	sum := 0
	for i := 0; i < 12; i++ {
		d := int(s[i] - '0')
		if i%2 == 0 {
			sum += d
		} else {
			sum += d * 3
		}
	}
	check := (10 - (sum % 10)) % 10
	return check == int(s[12]-'0')
}

func isValidISBN(s string) bool {
	n := normalizeISBN(s)
	if len(n) == 10 {
		return isValidISBN10(n)
	}
	if len(n) == 13 {
		return isValidISBN13(n)
	}
	return false
}

func hasMax2Decimals(n float64) bool {
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return false
	}
	scaled := n * 100
	rounded := math.Round(scaled)
	return math.Abs(scaled-rounded) < 1e-9
}

/* ------------ Public validators used by service ------------ */

func validateAndNormalizeCreate(in ports.CreateBookInput) (ports.CreateBookInput, error) {
	errs := &ValidationError{}

	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		errs.add("title", "Title is required")
	} else if len(in.Title) > 120 {
		errs.add("title", "Title must be ≤ 120 characters")
	}

	in.Author = strings.TrimSpace(in.Author)
	if in.Author == "" {
		errs.add("author", "Author is required")
	} else if len(in.Author) > 80 {
		errs.add("author", "Author must be ≤ 80 characters")
	}

	in.ISBN = strings.TrimSpace(in.ISBN)
	if in.ISBN == "" {
		errs.add("isbn", "ISBN is required")
	} else if !isValidISBN(in.ISBN) {
		errs.add("isbn", "Invalid ISBN (must be ISBN-10 or ISBN-13)")
	} else {
		in.ISBN = normalizeISBN(in.ISBN) // store normalized
	}

	// ---- PublicationYear (create) ----
	if in.PublicationYear == 0 {
		errs.add("publication_year", "Publication year is required")
	} else if !isValidFourDigitYearInt(in.PublicationYear) {
		errs.add("publication_year", "Publication year must be a 4-digit number")
	}

	if in.Price < 0 {
		errs.add("price", "Price must be ≥ 0")
	} else if in.Price > 1_000_000 {
		errs.add("price", "Price is too large")
	} else if !hasMax2Decimals(in.Price) {
		errs.add("price", "Max 2 decimal places")
	}

	if !errs.ok() {
		return in, errs
	}
	return in, nil
}

func validateAndNormalizeUpdate(in ports.UpdateBookInput) (ports.UpdateBookInput, error) {
	errs := &ValidationError{}

	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		if t == "" {
			errs.add("title", "Title is required")
		} else if len(t) > 120 {
			errs.add("title", "Title must be ≤ 120 characters")
		} else {
			*in.Title = t
		}
	}

	if in.Author != nil {
		a := strings.TrimSpace(*in.Author)
		if a == "" {
			errs.add("author", "Author is required")
		} else if len(a) > 80 {
			errs.add("author", "Author must be ≤ 80 characters")
		} else {
			*in.Author = a
		}
	}

	if in.ISBN != nil {
		s := strings.TrimSpace(*in.ISBN)
		if s == "" {
			errs.add("isbn", "ISBN is required")
		} else if !isValidISBN(s) {
			errs.add("isbn", "Invalid ISBN (must be ISBN-10 or ISBN-13)")
		} else {
			ns := normalizeISBN(s)
			*in.ISBN = ns
		}
	}

	// ---- PublicationYear (update) ----
	if in.PublicationYear != nil {
		y := *in.PublicationYear
		if !isValidFourDigitYearInt(y) {
			errs.add("publication_year", "Publication year must be a 4-digit number")
		}
	}

	if in.Price != nil {
		p := *in.Price
		if p < 0 {
			errs.add("price", "Price must be ≥ 0")
		} else if p > 1_000_000 {
			errs.add("price", "Price is too large")
		} else if !hasMax2Decimals(p) {
			errs.add("price", "Max 2 decimal places")
		}
	}

	if !errs.ok() {
		return in, errs
	}
	return in, nil
}

/* Optional: helper to pretty print (useful in logs) */
func (v *ValidationError) String() string {
	var b strings.Builder
	for k, m := range v.Fields {
		fmt.Fprintf(&b, "%s: %s; ", k, m)
	}
	return strings.TrimSpace(b.String())
}
