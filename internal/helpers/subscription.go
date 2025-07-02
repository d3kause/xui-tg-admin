package helpers

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"xui-tg-admin/internal/constants"
	"xui-tg-admin/internal/models"
)

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

	// Convert to slice for sorting
	var users []*UserTrafficSummary
	for _, summary := range userSummary {
		users = append(users, summary)
	}

	// Sort users by total traffic (descending)
	sort.Slice(users, func(i, j int) bool {
		return (users[i].TotalUp + users[i].TotalDown) > (users[j].TotalUp + users[j].TotalDown)
	})

	// Calculate totals
	var grandTotalUp, grandTotalDown int64
	activeCount := 0

	// Create formatted user lines
	var userLines []string

	for _, summary := range users {
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

		// Extract clean username (remove everything after @ or _)
		displayName := extractCleanUsername(summary.BaseUsername)

		// Convert traffic to GB with 2 decimal places
		downGB := float64(summary.TotalDown) / constants.BytesInGB
		upGB := float64(summary.TotalUp) / constants.BytesInGB

		// Add expiry info if set
		expiryInfo := ""
		if summary.ExpiryTime > 0 {
			expiryDate := time.Unix(summary.ExpiryTime/1000, 0)
			expiryInfo = fmt.Sprintf(" (until %s)", expiryDate.Format("02.01.06"))
		}

		userLine := UserLineData{
			StatusIcon:  statusIcon,
			DisplayName: displayName,
			DownGB:      downGB,
			UpGB:        upGB,
			ExpiryInfo:  expiryInfo,
		}
		userLines = append(userLines, formatUserLineAligned(userLine, 16))
	}

	// Add users section with proper alignment
	sb.WriteString("<pre>")
	for _, line := range userLines {
		sb.WriteString(line + "\n")
	}
	sb.WriteString("</pre>\n")

	// Add summary section
	sb.WriteString("\n<b>📈 Summary</b>\n")
	grandTotalDownGB := float64(grandTotalDown) / constants.BytesInGB
	grandTotalUpGB := float64(grandTotalUp) / constants.BytesInGB

	sb.WriteString("<pre>")
	sb.WriteString(fmt.Sprintf("%-15s %8.2f GB ⬇ %7.2f GB ⬆\n", "Total", grandTotalDownGB, grandTotalUpGB))
	sb.WriteString("─────────────────────────────\n")

	// Add per-inbound breakdown
	inboundTotals := make(map[string]*InboundTrafficStats)
	for _, summary := range users {
		for inboundName, stats := range summary.InboundStats {
			if inboundTotals[inboundName] == nil {
				inboundTotals[inboundName] = &InboundTrafficStats{Down: 0, Up: 0}
			}
			inboundTotals[inboundName].Down += stats.Down
			inboundTotals[inboundName].Up += stats.Up
		}
	}

	// Sort inbound names for consistent output
	var inboundNames []string
	for name := range inboundTotals {
		inboundNames = append(inboundNames, name)
	}
	sort.Strings(inboundNames)

	for _, inboundName := range inboundNames {
		stats := inboundTotals[inboundName]
		downGB := float64(stats.Down) / constants.BytesInGB
		upGB := float64(stats.Up) / constants.BytesInGB
		// Limit inbound name to 15 chars for alignment
		displayName := inboundName
		if len(displayName) > 15 {
			displayName = displayName[:12] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-15s %8.2f GB ⬇%7.2f GB ⬆\n", displayName+":", downGB, upGB))
	}
	sb.WriteString("</pre>")

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

// UserLineData represents data for formatting a user line
type UserLineData struct {
	StatusIcon  string
	DisplayName string
	DownGB      float64
	UpGB        float64
	ExpiryInfo  string
}

// formatUserLineAligned formats a user line with proper alignment
func formatUserLineAligned(data UserLineData, nameWidth int) string {
	displayName := data.DisplayName
	if len(displayName) > nameWidth {
		displayName = displayName[:nameWidth-3] + "..."
	}

	return fmt.Sprintf("%s %-*s %8.2f GB ⬇ %7.2f GB ⬆%s",
		data.StatusIcon,
		nameWidth,
		displayName,
		data.DownGB,
		data.UpGB,
		data.ExpiryInfo)
}
