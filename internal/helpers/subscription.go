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
