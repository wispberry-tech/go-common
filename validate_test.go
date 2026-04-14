package common

import (
	"testing"
)

type testStruct struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8"`
	UUID     string `json:"uuid" validate:"required,uuid"`
}

type displayTagStruct struct {
	Email string `json:"email" validate:"required,email" display:"Email Address"`
	Age   int    `json:"age" validate:"required,gte=18" display:"Your Age"`
}

func TestFormatValidationErrors_Required(t *testing.T) {
	err := Validate.Struct(testStruct{})
	resp := FormatValidationErrors(err)

	if resp.Error != "Validation failed" {
		t.Errorf("Error = %s, want 'Validation failed'", resp.Error)
	}
	if len(resp.Details) != 4 {
		t.Fatalf("expected 4 errors, got %d", len(resp.Details))
	}

	// All fields should have "is required" messages.
	for _, detail := range resp.Details {
		if detail.Message == "" {
			t.Errorf("empty message for field %s", detail.Field)
		}
	}
}

func TestFormatValidationErrors_Email(t *testing.T) {
	s := testStruct{Email: "notanemail", Name: "Alice", Password: "12345678", UUID: "550e8400-e29b-41d4-a716-446655440000"}
	err := Validate.Struct(s)
	if err == nil {
		t.Fatal("expected validation error")
	}
	resp := FormatValidationErrors(err)

	if len(resp.Details) != 1 {
		t.Fatalf("expected 1 error, got %d", len(resp.Details))
	}
	if resp.Details[0].Field != "email" {
		t.Errorf("field = %s, want email", resp.Details[0].Field)
	}
	want := "email must be a valid email address"
	if resp.Details[0].Message != want {
		t.Errorf("message = %q, want %q", resp.Details[0].Message, want)
	}
}

func TestFormatValidationErrors_MinMax(t *testing.T) {
	s := testStruct{Email: "a@b.com", Name: "A", Password: "short", UUID: "550e8400-e29b-41d4-a716-446655440000"}
	err := Validate.Struct(s)
	if err == nil {
		t.Fatal("expected validation error")
	}
	resp := FormatValidationErrors(err)

	if len(resp.Details) != 2 {
		t.Fatalf("expected 2 errors, got %d: %+v", len(resp.Details), resp.Details)
	}
}

func TestFormatValidationErrors_UnknownTag(t *testing.T) {
	type custom struct {
		URL string `validate:"required,url"`
	}
	s := custom{URL: "not-a-url"}
	err := Validate.Struct(s)
	if err == nil {
		t.Fatal("expected validation error")
	}
	resp := FormatValidationErrors(err)
	if len(resp.Details) != 1 {
		t.Fatalf("expected 1 error, got %d", len(resp.Details))
	}
	want := "URL must be a valid URL"
	if resp.Details[0].Message != want {
		t.Errorf("message = %q, want %q", resp.Details[0].Message, want)
	}
}

func TestFormatValidationErrorsFor_DisplayTag(t *testing.T) {
	s := displayTagStruct{Email: "bad", Age: 10}
	err := Validate.Struct(s)
	if err == nil {
		t.Fatal("expected validation error")
	}
	resp := FormatValidationErrorsFor(err, s)

	if len(resp.Details) != 2 {
		t.Fatalf("expected 2 errors, got %d: %+v", len(resp.Details), resp.Details)
	}

	messageMap := map[string]string{}
	for _, d := range resp.Details {
		messageMap[d.Field] = d.Message
	}

	if msg, ok := messageMap["email"]; !ok {
		t.Error("missing email error")
	} else if msg != "Email Address must be a valid email address" {
		t.Errorf("email message = %q, want display tag name", msg)
	}

	if msg, ok := messageMap["age"]; !ok {
		t.Error("missing age error")
	} else if msg != "Your Age must be greater than or equal to 18" {
		t.Errorf("age message = %q, want display tag name", msg)
	}
}

func TestFormatValidationErrorsFor_Pointer(t *testing.T) {
	s := &displayTagStruct{Email: "bad", Age: 10}
	err := Validate.Struct(s)
	if err == nil {
		t.Fatal("expected validation error")
	}
	resp := FormatValidationErrorsFor(err, s)

	// Should still resolve display tags through pointer.
	found := false
	for _, d := range resp.Details {
		if d.Field == "email" && d.Message == "Email Address must be a valid email address" {
			found = true
		}
	}
	if !found {
		t.Errorf("display tag not resolved through pointer: %+v", resp.Details)
	}
}

func TestFormatValidationErrors_ExpandedTags(t *testing.T) {
	tests := []struct {
		name    string
		v       any
		wantMsg string
	}{
		{
			"oneof",
			struct {
				Status string `validate:"required,oneof=active inactive"`
			}{Status: "deleted"},
			"Status must be one of: active inactive",
		},
		{
			"alpha",
			struct {
				Code string `validate:"required,alpha"`
			}{Code: "123"},
			"Code must contain only letters",
		},
		{
			"alphanum",
			struct {
				Code string `validate:"required,alphanum"`
			}{Code: "abc-123"},
			"Code must contain only letters and numbers",
		},
		{
			"numeric",
			struct {
				Amount string `validate:"required,numeric"`
			}{Amount: "abc"},
			"Amount must be numeric",
		},
		{
			"ip",
			struct {
				Addr string `validate:"required,ip"`
			}{Addr: "not-an-ip"},
			"Addr must be a valid IP address",
		},
		{
			"ipv4",
			struct {
				Addr string `validate:"required,ipv4"`
			}{Addr: "not-an-ip"},
			"Addr must be a valid IPv4 address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate.Struct(tt.v)
			if err == nil {
				t.Fatal("expected validation error")
			}
			resp := FormatValidationErrors(err)
			if len(resp.Details) != 1 {
				t.Fatalf("expected 1 error, got %d: %+v", len(resp.Details), resp.Details)
			}
			if resp.Details[0].Message != tt.wantMsg {
				t.Errorf("message = %q, want %q", resp.Details[0].Message, tt.wantMsg)
			}
		})
	}
}

func TestCamelCaseToWords(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ClientUUID", "Client UUID"},
		{"firstName", "first Name"},
		{"Name", "Name"},
		{"ID", "ID"},
		{"userID", "user ID"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := camelCaseToWords(tt.input)
			if got != tt.want {
				t.Errorf("camelCaseToWords(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
