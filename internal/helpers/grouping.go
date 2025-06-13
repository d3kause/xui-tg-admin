package helpers

import (
	"strings"
	"xui-tg-admin/internal/constants"
)

// GroupSimilarEmails groups emails if the difference in local part length is less than 3 characters
func GroupSimilarEmails(emails []string) []string {
	if len(emails) <= 1 {
		return emails
	}

	if ShouldGroupEmails(emails) {
		return []string{GenerateGroupName(emails)}
	}

	return emails
}

// ShouldGroupEmails checks if emails should be grouped based on length difference < 3 chars
func ShouldGroupEmails(emails []string) bool {
	if len(emails) <= 1 {
		return false
	}

	var localParts []string
	var domain string

	for _, email := range emails {
		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			localParts = append(localParts, email)
			continue
		}

		if domain == "" {
			domain = parts[1]
		} else if domain != parts[1] {
			return false
		}

		localParts = append(localParts, parts[0])
	}

	if len(localParts) < 2 {
		return false
	}

	minLen := len(localParts[0])
	maxLen := len(localParts[0])

	for _, local := range localParts[1:] {
		if len(local) < minLen {
			minLen = len(local)
		}
		if len(local) > maxLen {
			maxLen = len(local)
		}
	}

	return (maxLen - minLen) < constants.MaxLengthDifferenceForGrouping
}

// GenerateGroupName generates a common name for grouped emails by removing numeric suffixes
func GenerateGroupName(emails []string) string {
	if len(emails) <= 1 {
		return emails[0]
	}

	var localParts []string
	var domain string
	hasEmails := false

	for _, email := range emails {
		parts := strings.Split(email, "@")
		if len(parts) == 2 {
			hasEmails = true
			if domain == "" {
				domain = parts[1]
			}
			localParts = append(localParts, parts[0])
		} else {
			localParts = append(localParts, email)
		}
	}

	if len(localParts) == 0 {
		return strings.Join(emails, ", ")
	}

	commonPart := FindCommonPartWithoutSuffix(localParts)

	if hasEmails && domain != "" {
		return commonPart + "@" + domain
	}
	return commonPart
}

// FindCommonPartWithoutSuffix finds the common part of usernames by removing numeric suffixes
func FindCommonPartWithoutSuffix(names []string) string {
	if len(names) == 0 {
		return ""
	}
	if len(names) == 1 {
		return RemoveNumericSuffix(names[0])
	}

	var cleanedNames []string
	for _, name := range names {
		cleanedNames = append(cleanedNames, RemoveNumericSuffix(name))
	}

	commonPrefix := cleanedNames[0]
	for _, cleaned := range cleanedNames[1:] {
		commonPrefix = FindLongestCommonPrefix(commonPrefix, cleaned)
	}

	if len(commonPrefix) < constants.MinPrefixLengthForGrouping {
		return cleanedNames[0]
	}

	return commonPrefix
}

// RemoveNumericSuffix removes numeric suffixes like -2, -3, etc.
func RemoveNumericSuffix(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '-' {
			suffix := name[i+1:]
			if IsNumeric(suffix) {
				return name[:i]
			}
			break
		}
	}
	return name
}

// IsNumeric checks if a string contains only digits
func IsNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// FindLongestCommonPrefix finds the longest common prefix of two strings
func FindLongestCommonPrefix(s1, s2 string) string {
	minLen := len(s1)
	if len(s2) < minLen {
		minLen = len(s2)
	}

	for i := 0; i < minLen; i++ {
		if s1[i] != s2[i] {
			return s1[:i]
		}
	}
	return s1[:minLen]
}
