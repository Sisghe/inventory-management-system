package utils

import (
	"fmt"
	"time"
)

// HumanizeDurationMinutes converte una durata in una stringa leggibile (email-friendly)
func HumanizeDurationMinutes(d time.Duration) string {
	min := int(d.Round(time.Minute) / time.Minute)
	if min <= 1 {
		return "1 minuto"
	}
	return fmt.Sprintf("%d minuti", min)
}
