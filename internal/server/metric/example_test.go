package metric

import (
	"fmt"
)

func ExampleNewGaugeMetric() {
	// Создание метрики типа gauge
	gauge := NewGaugeMetric("response_time", 123.321)
	fmt.Printf("Gauge: id: %s, type: %s, value: %g\n",
		gauge.ID, gauge.MType, *gauge.Value)

	// Output:
	// Gauge: id: response_time, type: gauge, value: 123.321
}
func ExampleNewCounterMetric() {
	// Создание метрики типа counter
	counter := NewCounterMetric("requests", 123)
	fmt.Printf("Counter: id: %s, type: %s, delta: %v\n",
		counter.ID, counter.MType, *counter.Delta)

	// Output:
	// Counter: id: requests, type: counter, delta: 123
}

func ExampleMetric_GetHash() {
	// Создание метрики типа gauge
	gauge := NewGaugeMetric("response_time", 100)

	// Вычисление HMAC-хэша с использованием ключа "secret"
	hash := gauge.GetHash("secret")

	fmt.Printf("Hash: %s\n", hash)

	// Output:
	// Hash: ebe32462f14d296421c8e84415bdb651a548f354d6fc1327a9267d01d5a01bbb
}

func ExampleMetric_IsValidHash() {
	// Создание метрики типа gauge
	gaugeMetric := NewGaugeMetric("response_time", 100)

	// Вычисление HMAC-хэша с использованием ключа "secret"
	gaugeMetric.Hash = gaugeMetric.GetHash("secret")

	// Проверка корректности HMAC-хэша
	isValid := gaugeMetric.IsValidHash("secret")
	fmt.Printf("IsValid: %t\n", isValid)

	// Изменение хэша метрики
	gaugeMetric.Hash = "invalid hash"

	// Проверка корректности HMAC-хэша
	isValid = gaugeMetric.IsValidHash("secret")
	fmt.Printf("IsValid: %t\n", isValid)

	// Output:
	// IsValid: true
	// IsValid: false
}
