package helpers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"xui-tg-admin/internal/constants"
	"xui-tg-admin/internal/models"
)

// SubscriptionSummary represents aggregated subscription data
type SubscriptionSummary struct {
	TotalUp      int64
	TotalDown    int64
	Enable       bool
	ExpiryTime   int64
	InboundNames []string
	Emails       []string
}

// FormatDetailedUsersReport formats a detailed users information report
func FormatDetailedUsersReport(inbounds []models.Inbound) string {
	subscriptionSummary := AggregateUserDataBySubID(inbounds)

	if len(subscriptionSummary) == 0 {
		return "No active subscriptions found."
	}

	var totalUp, totalDown int64
	activeCount := 0

	for _, data := range subscriptionSummary {
		totalUp += data.TotalUp
		totalDown += data.TotalDown
		if data.Enable {
			activeCount++
		}
	}

	var sb strings.Builder
	sb.WriteString("<b>📊 Detailed Subscription Information</b>\n\n")

	totalUpGB := float64(totalUp) / constants.BytesInGB
	totalDownGB := float64(totalDown) / constants.BytesInGB

	for _, data := range subscriptionSummary {
		upGB := float64(data.TotalUp) / constants.BytesInGB
		downGB := float64(data.TotalDown) / constants.BytesInGB

		statusText := "🔴"
		if data.Enable {
			statusText = "🟢"
		}

		groupedEmails := GroupSimilarEmails(data.Emails)
		sb.WriteString(fmt.Sprintf("%s <b>%s</b>\n", statusText, strings.Join(groupedEmails, ", ")))
		sb.WriteString(fmt.Sprintf("├ ⬆️ %.2f GB\n", upGB))
		sb.WriteString(fmt.Sprintf("└ ⬇️ %.2f GB\n\n", downGB))
	}

	sb.WriteString("<b>📈 Summary</b>\n")
	sb.WriteString(fmt.Sprintf("├ 👥 Total: %d subscriptions (%d active)\n", len(subscriptionSummary), activeCount))
	sb.WriteString(fmt.Sprintf("├ ⬆️ Total Upload: %.2f GB\n", totalUpGB))
	sb.WriteString(fmt.Sprintf("└ ⬇️ Total Download: %.2f GB\n\n", totalDownGB))

	return sb.String()
}

// AggregateUserDataBySubID aggregates user data by subscription ID from all inbounds
func AggregateUserDataBySubID(inbounds []models.Inbound) map[string]*SubscriptionSummary {
	subscriptionSummary := make(map[string]*SubscriptionSummary)
	emailToSubID := CreateEmailToSubIDMapping(inbounds)

	for _, inbound := range inbounds {
		for _, clientStat := range inbound.ClientStats {
			subID, exists := emailToSubID[clientStat.Email]
			if !exists {
				continue
			}

			if subscriptionSummary[subID] == nil {
				subscriptionSummary[subID] = &SubscriptionSummary{
					TotalUp:      0,
					TotalDown:    0,
					Enable:       clientStat.Enable,
					ExpiryTime:   clientStat.ExpiryTime,
					InboundNames: []string{},
					Emails:       []string{},
				}
			}

			summary := subscriptionSummary[subID]
			summary.TotalUp += clientStat.Up
			summary.TotalDown += clientStat.Down

			if clientStat.Enable {
				summary.Enable = true
			}

			if clientStat.ExpiryTime > summary.ExpiryTime {
				summary.ExpiryTime = clientStat.ExpiryTime
			}

			addUniqueString(&summary.InboundNames, inbound.Remark)
			addUniqueString(&summary.Emails, clientStat.Email)
		}
	}

	return subscriptionSummary
}

// CreateEmailToSubIDMapping creates a mapping of email to subscription ID by parsing inbound settings
func CreateEmailToSubIDMapping(inbounds []models.Inbound) map[string]string {
	emailToSubID := make(map[string]string)

	for _, inbound := range inbounds {
		if inbound.Settings == "" {
			continue
		}

		var settings models.InboundSettings
		if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
			continue
		}

		for _, client := range settings.Clients {
			if client.SubID != "" {
				emailToSubID[client.Email] = client.SubID
			}
		}
	}

	return emailToSubID
}

// FormatSubscriptionInfo formats subscription information for a single user
func FormatSubscriptionInfo(baseUsername string, durationStr string, expiryTime int64, createdEmails []string, commonSubId string, addErrors []string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Client added successfully!\n\nBase username: %s\n", baseUsername))

	if expiryTime == 0 {
		sb.WriteString("Duration: ∞ (infinite)\n")
	} else {
		sb.WriteString(fmt.Sprintf("Duration: %s days\nExpiry: %s\n",
			durationStr,
			time.Unix(expiryTime/1000, 0).Format(constants.DateFormat)))
	}

	sb.WriteString("Traffic limit: Unlimited\n")
	sb.WriteString("\nCreated accounts:\n")
	for _, email := range createdEmails {
		sb.WriteString(fmt.Sprintf("\n- %s", email))
	}

	if len(createdEmails) > 0 {
		subURL := fmt.Sprintf("https://iris.xele.one:2096/sub/%s?name=%s", commonSubId, commonSubId)
		sb.WriteString(fmt.Sprintf("\n\nLink to connect: %s", subURL))
	}

	if len(addErrors) > 0 {
		sb.WriteString(fmt.Sprintf("\n\nWarning: Failed to add to some inbounds:\n%s\n", strings.Join(addErrors, "\n")))
	}

	return sb.String()
}

// addUniqueString adds a string to a slice if it's not already present
func addUniqueString(slice *[]string, item string) {
	for _, existing := range *slice {
		if existing == item {
			return
		}
	}
	*slice = append(*slice, item)
}

// FormatCompactTrafficReport formats a compact and beautiful traffic report for X-Ray users
func FormatCompactTrafficReport(inbounds []models.Inbound, onlineUsers []string) string {
	if len(inbounds) == 0 {
		return "📭 <b>No Users Found</b>\n\nThere are no users in the system yet."
	}

	// Create a set of online users for quick lookup
	onlineSet := make(map[string]bool)
	for _, user := range onlineUsers {
		// Extract base username from online user email
		baseUser := ExtractBaseUsername(user)
		onlineSet[baseUser] = true
	}

	// Aggregate user data by base username
	userSummary := make(map[string]*UserTrafficSummary)

	for _, inbound := range inbounds {
		for _, clientStat := range inbound.ClientStats {
			baseUsername := ExtractBaseUsername(clientStat.Email)

			if userSummary[baseUsername] == nil {
				userSummary[baseUsername] = &UserTrafficSummary{
					BaseUsername: baseUsername,
					TotalUp:      0,
					TotalDown:    0,
					Enable:       clientStat.Enable,
					ExpiryTime:   clientStat.ExpiryTime,
					InboundStats: make(map[string]*InboundTrafficStats),
				}
			}

			summary := userSummary[baseUsername]
			summary.TotalUp += clientStat.Up
			summary.TotalDown += clientStat.Down

			// Keep enabled status if any client is enabled
			if clientStat.Enable {
				summary.Enable = true
			}

			// Use the latest expiry time
			if clientStat.ExpiryTime > summary.ExpiryTime {
				summary.ExpiryTime = clientStat.ExpiryTime
			}

			// Track stats per inbound
			if summary.InboundStats[inbound.Remark] == nil {
				summary.InboundStats[inbound.Remark] = &InboundTrafficStats{
					Down: 0,
					Up:   0,
				}
			}
			summary.InboundStats[inbound.Remark].Down += clientStat.Down
			summary.InboundStats[inbound.Remark].Up += clientStat.Up
		}
	}

	if len(userSummary) == 0 {
		return "📭 <b>No Active Users</b>\n\nNo user traffic data available."
	}

	var sb strings.Builder
	sb.WriteString("<b>📊 Traffic Usage Report</b>\n\n")

	// Calculate totals
	var grandTotalUp, grandTotalDown int64
	activeCount := 0

	// Format each user
	for _, summary := range userSummary {
		grandTotalUp += summary.TotalUp
		grandTotalDown += summary.TotalDown

		if summary.Enable {
			activeCount++
		}

		// Determine online status
		statusIcon := "🔴"
		if onlineSet[summary.BaseUsername] {
			statusIcon = "🟢"
		}

		// Convert traffic to GB with 2 decimal places
		downGB := float64(summary.TotalDown) / constants.BytesInGB
		upGB := float64(summary.TotalUp) / constants.BytesInGB

		// Extract clean username (remove everything after @ or _)
		displayName := extractCleanUsername(summary.BaseUsername)

		// Format user line
		userLine := fmt.Sprintf("%s <b>%s</b> — ⬇ %.2f GB ⬆ %.2f GB", statusIcon, displayName, downGB, upGB)

		// Add expiry and limit info in parentheses if needed
		var infoItems []string

		// Add expiry date if set
		if summary.ExpiryTime > 0 {
			expiryDate := time.Unix(summary.ExpiryTime/1000, 0)
			infoItems = append(infoItems, fmt.Sprintf("until %s", expiryDate.Format("02.01.06")))
		}

		// Note: TrafficLimitGb is always 0 (unlimited) in this system, so we don't show it
		// If it was available, we would add: infoItems = append(infoItems, fmt.Sprintf("limit %d GB", limitGB))

		if len(infoItems) > 0 {
			userLine += fmt.Sprintf(" (%s)", strings.Join(infoItems, ", "))
		}

		sb.WriteString(userLine + "\n")
	}

	// Add summary section
	sb.WriteString("\n<b>📈 Summary</b>\n")
	grandTotalDownGB := float64(grandTotalDown) / constants.BytesInGB
	grandTotalUpGB := float64(grandTotalUp) / constants.BytesInGB

	sb.WriteString(fmt.Sprintf("Total ⬇ %.2f GB ⬆ %.2f GB\n", grandTotalDownGB, grandTotalUpGB))

	// Add per-inbound breakdown
	inboundTotals := make(map[string]*InboundTrafficStats)
	for _, summary := range userSummary {
		for inboundName, stats := range summary.InboundStats {
			if inboundTotals[inboundName] == nil {
				inboundTotals[inboundName] = &InboundTrafficStats{Down: 0, Up: 0}
			}
			inboundTotals[inboundName].Down += stats.Down
			inboundTotals[inboundName].Up += stats.Up
		}
	}

	for inboundName, stats := range inboundTotals {
		downGB := float64(stats.Down) / constants.BytesInGB
		upGB := float64(stats.Up) / constants.BytesInGB
		sb.WriteString(fmt.Sprintf("Inbound %s: ⬇ %.2f GB ⬆ %.2f GB\n", inboundName, downGB, upGB))
	}

	return sb.String()
}

// UserTrafficSummary represents aggregated traffic data for a user
type UserTrafficSummary struct {
	BaseUsername string
	TotalUp      int64
	TotalDown    int64
	Enable       bool
	ExpiryTime   int64
	InboundStats map[string]*InboundTrafficStats
}

// InboundTrafficStats represents traffic stats for a specific inbound
type InboundTrafficStats struct {
	Down int64
	Up   int64
}

// extractCleanUsername removes everything after @ or _ to get clean display name
func extractCleanUsername(username string) string {
	// Find @ symbol
	if atIndex := strings.Index(username, "@"); atIndex != -1 {
		return username[:atIndex]
	}

	// Find _ symbol
	if underIndex := strings.Index(username, "_"); underIndex != -1 {
		return username[:underIndex]
	}

	return username
}
