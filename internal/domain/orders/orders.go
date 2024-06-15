//nolint:wrapcheck
package orders

import (
	"errors"
	"time"
	"unicode"

	"github.com/andymarkow/gophermart/internal/domain/users"
)

var (
	ErrOrderNumberEmpty         = errors.New("order number is empty")
	ErrOrderNumberFormatInvalid = errors.New("order number format is invalid")
)

type OrderStatus string

var (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	Number     string
	UserLogin  string
	Status     OrderStatus
	Accrual    int
	UploadedAt time.Time
}

func NewOrder(number string, userLogin string) (*Order, error) {
	if err := ValidateOrderNumber(number); err != nil {
		return nil, err
	}

	if err := validateUserLogin(userLogin); err != nil {
		return nil, err
	}

	return &Order{
		Number:     number,
		UserLogin:  userLogin,
		UploadedAt: time.Now(),
		Status:     OrderStatusNew,
	}, nil
}

func ValidateOrderNumber(number string) error {
	if number == "" {
		return ErrOrderNumberEmpty
	}

	if !validateByLuhn(number) {
		return ErrOrderNumberFormatInvalid
	}

	return nil
}

func validateUserLogin(userLogin string) error {
	return users.ValidateLogin(userLogin)
}

// validateByLuhn checks number is valid or not based on Luhn algorithm.
func validateByLuhn(number string) bool {
	var sum int
	double := false

	for i := len(number) - 1; i >= 0; i-- {
		n := number[i]

		if !unicode.IsDigit(rune(n)) {
			return false // invalid character
		}

		digit := int(n - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit

		double = !double
	}

	return sum%10 == 0
}

// CalculateLuhn return the check number
// func CalculateLuhn(number int) int {
// 	checkNumber := checksum(number)

// 	if checkNumber == 0 {
// 		return 0
// 	}
// 	return 10 - checkNumber
// }

// // validateByLuhn checks if the number is valid or not based on Luhn algorithm
// func validateByLuhn(number int) bool {
// 	return (number%10+checksum(number/10))%10 == 0
// }

// func checksum(number int) int {
// 	var luhn int

// 	for i := 0; number > 0; i++ {
// 		cur := number % 10

// 		if i%2 == 0 { // even
// 			cur = cur * 2
// 			if cur > 9 {
// 				cur = cur%10 + cur/10
// 			}
// 		}

// 		luhn += cur
// 		number = number / 10
// 	}
// 	return luhn % 10
// }
