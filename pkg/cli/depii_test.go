package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDepiiMultiline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple uuid",
			input:    "X-Request-ID: ef1bb310-6f06-40c7-a153-046683106cc5",
			expected: "X-Request-ID: {UUID-0}",
		},
		{
			name:     "same uuid twice",
			input:    "ID1: ef1bb310-6f06-40c7-a153-046683106cc5 ID2: ef1bb310-6f06-40c7-a153-046683106cc5",
			expected: "ID1: {UUID-0} ID2: {UUID-0}",
		},
		{
			name:     "different uuids",
			input:    "ID1: ef1bb310-6f06-40c7-a153-046683106cc5 ID2: a1234567-89ab-cdef-0123-456789abcdef",
			expected: "ID1: {UUID-0} ID2: {UUID-1}",
		},
		{
			name:     "uppercase uuid",
			input:    "ID: EF1BB310-6F06-40C7-A153-046683106CC5",
			expected: "ID: {UUID-0}",
		},
		{
			name:     "same uuid different case",
			input:    "ID1: ef1bb310-6f06-40c7-a153-046683106cc5 ID2: EF1BB310-6F06-40C7-A153-046683106CC5",
			expected: "ID1: {UUID-0} ID2: {UUID-0}",
		},
		{
			name:     "simple alfanum20",
			input:    "token: abcdefghij0123456789",
			expected: "token: {ALFANUM20-0}",
		},
		{
			name:     "same alfanum20 twice",
			input:    "token1: abcdefghij0123456789 token2: abcdefghij0123456789",
			expected: "token1: {ALFANUM20-0} token2: {ALFANUM20-0}",
		},
		{
			name:     "different alfanum20s",
			input:    "token1: abcdefghij0123456789 token2: zyxwvutsrq9876543210",
			expected: "token1: {ALFANUM20-0} token2: {ALFANUM20-1}",
		},
		{
			name:     "mixed uuid and alfanum20",
			input:    "request: ef1bb310-6f06-40c7-a153-046683106cc5 token: abcdefghij0123456789",
			expected: "request: {UUID-0} token: {ALFANUM20-0}",
		},
		{
			name:     "multiple of each",
			input:    "req1: ef1bb310-6f06-40c7-a153-046683106cc5 tok1: abcdefghij0123456789 req2: a1234567-89ab-cdef-0123-456789abcdef tok2: zyxwvutsrq9876543210",
			expected: "req1: {UUID-0} tok1: {ALFANUM20-0} req2: {UUID-1} tok2: {ALFANUM20-1}",
		},
		{
			name:     "no pii",
			input:    "hello world 123",
			expected: "hello world 123",
		},
		{
			name:     "uuid with surrounding text",
			input:    "Request ID is ef1bb310-6f06-40c7-a153-046683106cc5 please",
			expected: "Request ID is {UUID-0} please",
		},
		{
			name:     "alfanum20 case sensitive",
			input:    "token: abcdefghij0123456789 TOKEN: ABCDEFGHIJ0123456789",
			expected: "token: {ALFANUM20-0} TOKEN: ABCDEFGHIJ0123456789",
		},
		{
			name:     "alfanum20 exactly 20 chars",
			input:    "x: 12345678901234567890",
			expected: "x: {ALFANUM20-0}",
		},
		{
			name:     "alfanum20 not triggered on 19 chars",
			input:    "x: 1234567890123456789",
			expected: "x: 1234567890123456789",
		},
		{
			name:     "alfanum20 not triggered on 21 chars",
			input:    "x: 123456789012345678901",
			expected: "x: 123456789012345678901",
		},
		{
			name:     "multiline with state preservation",
			input:    "Line 1: ef1bb310-6f06-40c7-a153-046683106cc5 with token abcdefghij0123456789\nLine 2: ef1bb310-6f06-40c7-a153-046683106cc5 same UUID, token zyxwvutsrq9876543210\nUUID: a1234567-89ab-cdef-0123-456789abcdef",
			expected: "Line 1: {UUID-0} with token {ALFANUM20-0}\nLine 2: {UUID-0} same UUID, token {ALFANUM20-1}\nUUID: {UUID-1}",
		},
		{
			name:     "long domain 3 parts",
			input:    "api.mts.ru",
			expected: "{DOMAIN-0}",
		},
		{
			name:     "long domain 4 parts",
			input:    "console.yandex.cloud",
			expected: "{DOMAIN-0}",
		},
		{
			name:     "short domain 2 parts not replaced",
			input:    "ya.ru google.com",
			expected: "ya.ru google.com",
		},
		{
			name:     "same domain twice",
			input:    "api.mts.ru and api.mts.ru",
			expected: "{DOMAIN-0} and {DOMAIN-0}",
		},
		{
			name:     "different domains",
			input:    "api.mts.ru console.yandex.cloud",
			expected: "{DOMAIN-0} {DOMAIN-1}",
		},
		{
			name:     "domain case insensitive",
			input:    "API.MTS.RU and api.mts.ru",
			expected: "{DOMAIN-0} and {DOMAIN-0}",
		},
		{
			name:     "domain with hyphen",
			input:    "api.my-server.ru",
			expected: "{DOMAIN-0}",
		},
		{
			name:     "4 domains",
			input:    "x.api.my-server.ru",
			expected: "{DOMAIN-0}",
		},
		{
			name:     "mixed uuid, alfanum20, and domain",
			input:    "Request to api.mts.ru with ID ef1bb310-6f06-40c7-a153-046683106cc5 and token abcdefghij0123456789",
			expected: "Request to {DOMAIN-0} with ID {UUID-0} and token {ALFANUM20-0}",
		},
		{
			name:     "url",
			input:    "https://api.my-server.ru/index.html",
			expected: "https://{DOMAIN-0}/index.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := depii(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
