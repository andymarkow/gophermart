//nolint:wrapcheck
package orders

import (
	"errors"
	"time"
	"unicode"

	"github.com/andymarkow/gophermart/internal/domain/users"
	"github.com/shopspring/decimal"
)

var (
	ErrOrderNumberEmpty         = errors.New("order number is empty")
	ErrOrderNumberFormatInvalid = errors.New("order number format is invalid")
)

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	number     string
	userLogin  string
	status     OrderStatus
	accrual    decimal.Decimal
	uploadedAt time.Time
}

func NewOrder(number string, userLogin string) (*Order, error) {
	if err := ValidateOrderNumber(number); err != nil {
		return nil, err
	}

	if err := validateUserLogin(userLogin); err != nil {
		return nil, err
	}

	return &Order{
		number:     number,
		userLogin:  userLogin,
		status:     OrderStatusNew,
		uploadedAt: time.Now(),
	}, nil
}

func (o *Order) Number() string {
	return o.number
}

func (o *Order) UserLogin() string {
	return o.userLogin
}

func (o *Order) Status() OrderStatus {
	return o.status
}

func (o *Order) Accrual() decimal.Decimal {
	return o.accrual
}

func (o *Order) UploadedAt() time.Time {
	return o.uploadedAt
}

func (o *Order) SetStatus(status OrderStatus) {
	o.status = status
}

func (o *Order) SetAccrual(accrual decimal.Decimal) {
	o.accrual = accrual
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
