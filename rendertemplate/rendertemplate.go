package rendertemplate

import (
	"fmt"
	"strconv"
	"strings"
)

func Render(template string, paramStrIn *string) string {
	if paramStrIn == nil || *paramStrIn == "" {
		return template
	}

	paramStr := *paramStrIn

	paramCount := strings.Count(paramStr, "=")

	if paramCount == 0 {
		return template
	}

	replacerArgs := make([]string, 0, paramCount*2)

	start := 0
	for i := 0; i < len(paramStr); i++ {
		if paramStr[i] == ';' || i == len(paramStr)-1 {
			end := i
			if i == len(paramStr)-1 && paramStr[i] != ';' {
				end = i + 1
			}

			equalPos := strings.Index(paramStr[start:end], "=")
			if equalPos != -1 {
				key := strings.TrimSpace(paramStr[start : start+equalPos])
				value := strings.TrimSpace(paramStr[start+equalPos+1 : end])

				// Форматирование значений ключей
				if strings.HasSuffix(key, "_AMT") {
					value = formatAmount(value)
				} else if key == "CUR" {
					value = formatCurrency(value)
				}

				replacerArgs = append(replacerArgs, "{{"+key+"}}", value)
			}
			start = i + 1
		}
	}

	replacer := strings.NewReplacer(replacerArgs...)
	return replacer.Replace(template)
}

// Функция для форматирования сумм
func formatAmount(amount string) string {
	amount = strings.ReplaceAll(amount, " ", "")  // Удаление всех пробелов
	amount = strings.ReplaceAll(amount, ",", ".") // Замена всех запятых на точки
	parsed, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return amount // Возвращаем оригинальное значение, если произошла ошибка
	}
	// Разделяем целую и дробную части
	parts := strings.Split(fmt.Sprintf("%.2f", parsed), ".")
	intPart := parts[0]
	fracPart := parts[1]

	// Форматируем целую часть с пробелами
	var result strings.Builder
	n := len(intPart)
	for i, digit := range intPart {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteString(" ")
		}
		result.WriteRune(digit)
	}

	// Добавляем дробную часть
	result.WriteString(".")
	result.WriteString(fracPart)

	return result.String()
}

// Функция для форматирования валют
func formatCurrency(currency string) string {
	currency = strings.ReplaceAll(currency, " ", "") // Удаление всех пробелов
	switch currency {
	case "USD":
		return "$"
	case "EUR":
		return "€"
	case "KZT":
		return "₸"
	case "RUB":
		return "₽"
	case "CNY", "JPY":
		return "¥"
	case "CHF":
		return "Fr"
	case "GBP":
		return "£"
	case "KGS":
		return "c"
	case "CAD":
		return "CA$"
	case "SEK":
		return "kr"
	case "AUD":
		return "A$"
	case "TRY":
		return "₺"
	case "TJS":
		return "SM"
	case "AED":
		return "Dh"
	case "UZS":
		return "сўм"
	default:
		return currency // Возвращаем оригинальное значение, если валюта не найдена в списке
	}
}
