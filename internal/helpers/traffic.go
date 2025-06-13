package helpers

import (
	"fmt"
	"strings"
	"xui-tg-admin/internal/constants"
	"xui-tg-admin/internal/models"
)

// FormatNetworkUsageReport formats a beautiful network usage report
func FormatNetworkUsageReport(inbounds []models.Inbound) string {
	var sb strings.Builder
	sb.WriteString("<b>Network Usage Report:</b>\n")
	sb.WriteString("<pre>\n")
	sb.WriteString("Email             | ↓ (GB) | ↑ (GB)\n")
	sb.WriteString("------------------|--------|--------\n")

	var totalUploadGB int64 = 0
	var totalDownloadGB int64 = 0

	for _, inbound := range inbounds {
		if len(inbound.ClientStats) == 0 {
			continue
		}

		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("Inbound: %s\n", inbound.Remark))

		inboundDownloadTotal, inboundUploadTotal := CalculateInboundTraffic(inbound.ClientStats)
		totalDownloadGB += inboundDownloadTotal
		totalUploadGB += inboundUploadTotal

		for _, client := range inbound.ClientStats {
			sb.WriteString(FormatTableLine(client.Email, client.Down, client.Up))
		}

		sb.WriteString("-----------\n")
		sb.WriteString(FormatTableLine("Total:", inboundDownloadTotal*constants.BytesInGB, inboundUploadTotal*constants.BytesInGB))
	}

	sb.WriteString("\n")
	sb.WriteString(FormatTableLine("Grand Total:", totalDownloadGB*constants.BytesInGB, totalUploadGB*constants.BytesInGB))
	sb.WriteString("</pre>")

	return sb.String()
}

// CalculateInboundTraffic calculates total traffic for an inbound (in GB)
func CalculateInboundTraffic(clientStats []models.ClientStat) (downloadGB int64, uploadGB int64) {
	for _, client := range clientStats {
		downloadGB += client.Down / constants.BytesInGB
		uploadGB += client.Up / constants.BytesInGB
	}
	return
}

// FormatTableLine formats a single line of the traffic table
func FormatTableLine(email string, downBytes int64, upBytes int64) string {
	downGB := float64(downBytes) / constants.BytesInGB
	upGB := float64(upBytes) / constants.BytesInGB

	displayEmail := email
	if len(email) > constants.MaxEmailDisplayLength {
		displayEmail = email[:constants.MaxEmailSuffixLength] + "..."
	}

	return fmt.Sprintf("%-17s | %6.2f | %6.2f\n", displayEmail, downGB, upGB)
}
