package metric

import (
	"fmt"
)

func ExampleNewMetric() {
	// Создание метрики типа counter
	counter := NewMetric("requests", CounterType, int64(1))
	fmt.Printf("Counter: id: %s, type: %s, delta: %v\n",
		counter.ID, counter.MType, *counter.Delta)

	// Создание метрики типа gauge
	gauge := NewMetric("response_time", GaugeType, float64(100))
	fmt.Printf("Gauge: id: %s, type: %s, value: %g\n",
		gauge.ID, gauge.MType, *gauge.Value)

	// Output:
	// Counter: id: requests, type: counter, delta: 1
	// Gauge: id: response_time, type: gauge, value: 100
}

func ExampleMetric_GetHash() {
	// Создание метрики типа gauge
	gauge := NewMetric("response_time", GaugeType, float64(100))

	// Вычисление HMAC-хэша с использованием ключа "secret"
	hash := gauge.GetHash("secret")

	fmt.Printf("Hash: %s\n", hash)

	// Output:
	// Hash: ebe32462f14d296421c8e84415bdb651a548f354d6fc1327a9267d01d5a01bbb
}

func ExampleMetric_IsValidHash() {
	// Создание метрики типа gauge
	gaugeMetric := NewMetric("response_time", GaugeType, float64(100))

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
