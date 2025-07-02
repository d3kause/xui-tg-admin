package models

import (
	"fmt"
	"sort"
	"time"
)

// SortType представляет тип сортировки пользователей
type SortType int

const (
	SortByCreationOrder SortType = iota // По порядку создания (ID)
	SortByExpiryDate                    // По дате истечения
	SortByTrafficTotal                  // По общему трафику
	SortByStatus                        // По статусу (активные первые)
	SortByName                          // По имени (алфавитный)
)

// MemberInfo содержит расширенную информацию о пользователе для сортировки и фильтрации
type MemberInfo struct {
	BaseUsername string   // Базовое имя пользователя (без постфикса)
	FullEmails   []string // Все email'ы пользователя во всех inbound'ах
	ID           int      // ID для сортировки по порядку создания
	Enable       bool     // Активен ли пользователь
	ExpiryTime   int64    // Время истечения (миллисекунды)
	TotalUp      int64    // Общий загруженный трафик
	TotalDown    int64    // Общий скачанный трафик
	TotalTraffic int64    // Общий трафик (Up + Down)
	IsExpired    bool     // Истек ли срок действия
}

// GetSortName возвращает читаемое название типа сортировки
func (st SortType) GetSortName() string {
	switch st {
	case SortByCreationOrder:
		return "📅 По дате добавления"
	case SortByExpiryDate:
		return "⏰ По дате истечения"
	case SortByTrafficTotal:
		return "📊 По общему трафику"
	case SortByStatus:
		return "🔄 По статусу"
	case SortByName:
		return "🔤 По имени"
	default:
		return "📅 По дате добавления"
	}
}

// IsExpiredMember проверяет, истек ли срок действия пользователя
func (m *MemberInfo) IsExpiredMember() bool {
	if m.ExpiryTime == 0 {
		return false // Бессрочный
	}
	return time.Now().UnixMilli() > m.ExpiryTime
}

// GetExpiryStatus возвращает статус истечения в читаемом виде
func (m *MemberInfo) GetExpiryStatus() string {
	if m.ExpiryTime == 0 {
		return "∞ Бессрочный"
	}

	if m.IsExpiredMember() {
		return "❌ Истек"
	}

	expiryDate := time.Unix(m.ExpiryTime/1000, 0)
	daysLeft := int(time.Until(expiryDate).Hours() / 24)

	if daysLeft <= 0 {
		return "⚠️ Истекает сегодня"
	} else if daysLeft <= 7 {
		return fmt.Sprintf("⚠️ %d дн.", daysLeft)
	}

	return fmt.Sprintf("✅ %d дн.", daysLeft)
}

// SortMembers сортирует список пользователей по указанному типу
func SortMembers(members []MemberInfo, sortType SortType) {
	sort.Slice(members, func(i, j int) bool {
		switch sortType {
		case SortByCreationOrder:
			return members[i].ID < members[j].ID
		case SortByExpiryDate:
			// Бессрочные в конец, остальные по возрастанию даты истечения
			if members[i].ExpiryTime == 0 && members[j].ExpiryTime == 0 {
				return members[i].BaseUsername < members[j].BaseUsername
			}
			if members[i].ExpiryTime == 0 {
				return false
			}
			if members[j].ExpiryTime == 0 {
				return true
			}
			return members[i].ExpiryTime < members[j].ExpiryTime
		case SortByTrafficTotal:
			return members[i].TotalTraffic > members[j].TotalTraffic // По убыванию
		case SortByStatus:
			// Активные первые, потом неактивные
			if members[i].Enable != members[j].Enable {
				return members[i].Enable
			}
			return members[i].BaseUsername < members[j].BaseUsername
		case SortByName:
			return members[i].BaseUsername < members[j].BaseUsername
		default:
			return members[i].ID < members[j].ID
		}
	})
}
