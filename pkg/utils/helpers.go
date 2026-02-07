// Copyright vfcastr 2025
// SPDX-License-Identifier: MPL-2.0

package utils

import "strings"

// ToBoolPtr converts a bool to a pointer to bool
func ToBoolPtr(b bool) *bool {
	return &b
}

// ToStringPtr converts a string to a pointer to string
func ToStringPtr(s string) *string {
	return &s
}

// ToIntPtr converts an int to a pointer to int
func ToIntPtr(i int) *int {
	return &i
}

// SplitAndTrim splits a comma-separated string and trims spaces
func SplitAndTrim(s string) []string {
	var result []string
	if s == "" {
		return result
	}
	for _, part := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
