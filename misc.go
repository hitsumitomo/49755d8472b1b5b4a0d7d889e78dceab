package main

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ------------------------------------------------------------------------------------- IBAN validation
// validateIBAN checks if the given IBAN is valid. // DE89370400440532013000
func validateIBAN(iban string) bool {
	// 1. IBAN length must be between 15 and 34 characters
	if len(iban) < 15 || len(iban) > 34 {
		return false
	}

	// 2. Normalize IBAN
	iban = strings.ReplaceAll(iban, " ", "")
	iban = strings.ToUpper(iban)

	// 3. Check if the first two characters are letters
	if !isAlpha(iban[:2]) {
		return false
	}

	// 4. Move the first 4 characters to the end
	iban = iban[4:] + iban[:4]

	// 5. Change letters to numbers (A = 10, B = 11, ..., Z = 35)
	ibanWithDigits := ""
	for _, ch := range iban {
		if unicode.IsLetter(ch) {
			ibanWithDigits += fmt.Sprintf("%d", ch - 'A' + 10)

		} else {
			ibanWithDigits += string(ch)
		}
	}

	// 6. Check by Modulo 97
	return mod97(ibanWithDigits) == 1
}

func isAlpha(s string) bool {
	matched, _ := regexp.MatchString("^[A-Z]+$", s)
	return matched
}

// mod97 calculates the modulo 97 of the given IBAN.
func mod97(iban string) (result int) {
	// cut the IBAN into blocks of 9 and divide each by 97
	for i := 0; i < len(iban); i++ {
		result = (result * 10 + int(iban[i] - '0')) % 97
	}
	return result
}
// ------------------------------------------------------------------------------------- /IBAN validation

func validateNumber(account string) bool {
	return true
}

func validateName(name string) bool {
	return true
}

func validateAddress(address string) bool {
	return true
}

func validateAmount(amount float64, limits ...float64) bool {
	if len(limits) == 1 {
		if amount < limits[0] {
			return false
		}
	} else if len(limits) == 2 {
		if amount < limits[0] || amount > limits[1] {
			return false
		}
	}
	return true
}

func validateAccountType(accountType string) bool {
	return accountType == typeSending || accountType == typeReceiving
}

//go:embed index.html
var indexHTML string