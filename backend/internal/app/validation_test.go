package app

import (
	"math"
	"testing"

	"github.com/gerry-sabar/byfood/internal/ports"
)

// --- normalize & ISBN helpers ---

func TestNormalizeISBN(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"978-0-321-12521-7", "9780321125217"},
		{"0 321 14653 0", "0321146530"},
		{"  978 0 13 235088 4  ", "9780132350884"},
		{"0-8044-2957-x", "080442957X"}, // just check uppercasing of x
	}
	for _, tc := range cases {
		if got := normalizeISBN(tc.in); got != tc.want {
			t.Fatalf("normalizeISBN(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestIsValidISBN(t *testing.T) {
	valid := []string{
		"0306406152",        // valid ISBN-10 (classic example)
		"9780306406157",     // matching ISBN-13
		"9780132350884",     // Clean Code
		"978-0-321-12521-7", // with hyphens (normalize then validate)
	}
	for _, s := range valid {
		if !isValidISBN(s) {
			t.Fatalf("expected valid ISBN: %q", s)
		}
	}

	invalid := []string{
		"", "123", "abcdefghijk", "9780132350885", // bad checksum
		"0306406153", // bad checksum for 10
	}
	for _, s := range invalid {
		if isValidISBN(s) {
			t.Fatalf("expected invalid ISBN: %q", s)
		}
	}
}

func TestHasMax2Decimals(t *testing.T) {
	type tc struct {
		v    float64
		want bool
	}
	cases := []tc{
		{10, true},
		{10.2, true},
		{10.23, true},
		{10.2300000000001, true}, // epsilon-friendly
		{10.234, false},
		{-1.001, false},
		{math.NaN(), false},
		{math.Inf(1), false},
		{math.Inf(-1), false},
	}
	for _, c := range cases {
		if got := hasMax2Decimals(c.v); got != c.want {
			t.Fatalf("hasMax2Decimals(%v) = %v, want %v", c.v, got, c.want)
		}
	}
}

// --- validateAndNormalizeCreate ---

func TestValidateAndNormalizeCreate_OK(t *testing.T) {
	in := ports.CreateBookInput{
		Title:           "  Clean Code  ",
		Author:          " Robert C. Martin ",
		ISBN:            "978-0-13-235088-4",
		PublicationYear: 2008,
		Price:           33.50,
	}
	out, err := validateAndNormalizeCreate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Title != "Clean Code" {
		t.Fatalf("Title not trimmed/kept: %q", out.Title)
	}
	if out.Author != "Robert C. Martin" {
		t.Fatalf("Author not trimmed/kept: %q", out.Author)
	}
	if out.ISBN != "9780132350884" {
		t.Fatalf("ISBN not normalized: %q", out.ISBN)
	}
	if out.PublicationYear != 2008 {
		t.Fatalf("PublicationYear altered: %v", out.PublicationYear)
	}
	if out.Price != 33.50 {
		t.Fatalf("Price altered: %v", out.Price)
	}
}

func TestValidateAndNormalizeCreate_Errors(t *testing.T) {
	long := func(n int) string {
		s := make([]byte, n)
		for i := range s {
			s[i] = 'a'
		}
		return string(s)
	}
	in := ports.CreateBookInput{
		Title:           " ",        // required
		Author:          long(81),   // > 80
		ISBN:            "bad-isbn", // invalid
		PublicationYear: 123,        // not 4 digits
		Price:           12.345,     // > 2 decimals
	}
	_, err := validateAndNormalizeCreate(in)
	if err == nil {
		t.Fatal("expected validation error")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("want *ValidationError, got %T", err)
	}
	// assert some fields present
	if ve.Fields["title"] == "" {
		t.Fatalf("missing title error")
	}
	if ve.Fields["author"] == "" {
		t.Fatalf("missing author error")
	}
	if ve.Fields["isbn"] == "" {
		t.Fatalf("missing isbn error")
	}
	if ve.Fields["publication_year"] == "" {
		t.Fatalf("missing publication_year error")
	}
	if ve.Fields["price"] == "" {
		t.Fatalf("missing price error")
	}
	// ensure first error per field kept (no overwrite)
	in2 := ports.CreateBookInput{
		Title:           "", // required
		Author:          "", // required
		ISBN:            "", // required
		PublicationYear: 0,  // required
		Price:           -1, // < 0
	}
	_, err2 := validateAndNormalizeCreate(in2)
	if err2 == nil {
		t.Fatal("expected validation error")
	}
	ve2 := err2.(*ValidationError)
	if ve2.Fields["price"] == "" ||
		ve2.Fields["title"] == "" ||
		ve2.Fields["author"] == "" ||
		ve2.Fields["isbn"] == "" ||
		ve2.Fields["publication_year"] == "" {
		t.Fatalf("expected all field errors, got: %+v", ve2.Fields)
	}
}

// --- validateAndNormalizeUpdate ---

func TestValidateAndNormalizeUpdate_OK_Partial(t *testing.T) {
	title := "  New Title  "
	isbn := "978-0-321-12521-7"
	price := 12.30
	in := ports.UpdateBookInput{
		Title:           &title,
		ISBN:            &isbn,
		Price:           &price,
		PublicationYear: nil, // omitted → unchanged/ignored
		// Author nil → unchanged/ignored
	}
	out, err := validateAndNormalizeUpdate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Title == nil || *out.Title != "New Title" {
		t.Fatalf("Title not trimmed: %v", out.Title)
	}
	if out.ISBN == nil || *out.ISBN != "9780321125217" {
		t.Fatalf("ISBN not normalized: %v", out.ISBN)
	}
	if out.Price == nil || *out.Price != 12.30 {
		t.Fatalf("Price changed: %v", out.Price)
	}
	if out.Author != nil {
		t.Fatalf("Author should remain nil, got: %v", out.Author)
	}
	if out.PublicationYear != nil {
		t.Fatalf("PublicationYear should remain nil when omitted, got: %v", *out.PublicationYear)
	}
}

func TestValidateAndNormalizeUpdate_UpdatePublicationYearOnly(t *testing.T) {
	yr := 2020
	in := ports.UpdateBookInput{
		PublicationYear: &yr,
	}
	out, err := validateAndNormalizeUpdate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.PublicationYear == nil || *out.PublicationYear != 2020 {
		t.Fatalf("PublicationYear not kept: %v", out.PublicationYear)
	}
	// others should stay nil
	if out.Title != nil || out.Author != nil || out.ISBN != nil || out.Price != nil {
		t.Fatalf("unexpected non-nil fields in partial update: %+v", out)
	}
}

func TestValidateAndNormalizeUpdate_Errors(t *testing.T) {
	badt := " "
	bada := " "
	badi := "bad"
	badp := 1.239
	bady := 123 // not 4 digits
	in := ports.UpdateBookInput{
		Title:           &badt,
		Author:          &bada,
		ISBN:            &badi,
		Price:           &badp,
		PublicationYear: &bady,
	}
	_, err := validateAndNormalizeUpdate(in)
	if err == nil {
		t.Fatal("expected validation error")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("want *ValidationError, got %T", err)
	}
	if ve.Fields["title"] == "" ||
		ve.Fields["author"] == "" ||
		ve.Fields["isbn"] == "" ||
		ve.Fields["price"] == "" ||
		ve.Fields["publication_year"] == "" {
		t.Fatalf("missing field errors: %+v", ve.Fields)
	}
}

// --- ValidationError helpers ---

func TestValidationError_ErrorAndString(t *testing.T) {
	v := &ValidationError{}
	v.add("title", "bad")
	v.add("title", "should-be-ignored") // first one sticks
	v.add("price", "too many decimals")
	if v.Error() != "validation error" {
		t.Fatalf("Error() = %q", v.Error())
	}
	s := v.String()
	// non-deterministic field order, so just check substrings
	if !(contains(s, "title: bad") && contains(s, "price: too many decimals")) {
		t.Fatalf("String() content: %q", s)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool { return (len(sub) == 0) || (indexOf(s, sub) >= 0) })()
}
func indexOf(s, sub string) int {
	// tiny inline contains to keep deps zero; fine for tests
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
