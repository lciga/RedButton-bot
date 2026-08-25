package scoring

import "math"

// Функция расчёта стоимости таска по формуле CTFd.
// Возвращает минимальную стоимость после полного распада.
func Calculate(initial, minimum, decay int, solveCount int64) int {
	if decay <= 0 {
		return initial
	}

	initialValue := float64(initial)
	minimumValue := float64(minimum)
	decayValue := float64(decay)
	solves := float64(solveCount)
	value := int(math.Ceil(
		((minimumValue-initialValue)/(decayValue*decayValue))*(solves*solves) + initialValue,
	))
	if value < minimum {
		return minimum
	}

	return value
}
