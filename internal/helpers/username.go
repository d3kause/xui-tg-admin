package helpers

import (
	"fmt"
	"xui-tg-admin/internal/constants"
)

// ExtractBaseUsername извлекает базовое имя пользователя без постфикса номера инбаунда
// Например: "qwe-qwe-qwe-1" -> "qwe-qwe-qwe", "user123-2" -> "user123", "user123" -> "user123"
func ExtractBaseUsername(email string) string {
	// Ищем с конца строки последний дефис, за которым идут только цифры
	for i := len(email) - 1; i >= 0; i-- {
		if email[i] == constants.UsernameSeparator[0] {
			// Проверяем, что после дефиса идут только цифры
			suffix := email[i+1:]
			if len(suffix) > 0 && IsNumeric(suffix) {
				return email[:i]
			}
			// Если после дефиса не только цифры, продолжаем поиск
		}
	}
	// Если не нашли дефис с цифрами, возвращаем всю строку
	return email
}

// IsEmailMatchingBaseUsername проверяет, соответствует ли email базовому имени пользователя
// Например: IsEmailMatchingBaseUsername("qwe-qwe-qwe-1", "qwe-qwe-qwe") -> true
//
//	IsEmailMatchingBaseUsername("user123-2", "user123") -> true
//	IsEmailMatchingBaseUsername("user123", "user123") -> true
//	IsEmailMatchingBaseUsername("user456-1", "user123") -> false
func IsEmailMatchingBaseUsername(email, baseUsername string) bool {
	// Сначала извлекаем базовое имя из email
	extractedBase := ExtractBaseUsername(email)
	return extractedBase == baseUsername
}

// FormatEmailWithInboundNumber форматирует email с номером инбаунда
// Например: FormatEmailWithInboundNumber("qwe-qwe-qwe", 1) -> "qwe-qwe-qwe-1"
func FormatEmailWithInboundNumber(baseUsername string, inboundNumber int) string {
	return fmt.Sprintf("%s%s%d", baseUsername, constants.UsernameSeparator, inboundNumber)
}
