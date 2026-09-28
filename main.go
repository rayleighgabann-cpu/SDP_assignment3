package main

import (
	"errors"
	"fmt"
	"os"
)

// ==========================================
// 1. БАЗОВЫЕ ТИПЫ И КОНТРАКТЫ (Bridge)
// ==========================================

// ReportData - данные для отчета.
type ReportData struct {
	Title string
	Rows  []map[string]string
}

// Ошибки контракта Implementor.
var (
	ErrExportFailed  = errors.New("export failed")
	ErrFormatInvalid = errors.New("invalid format")
)

// Exporter - интерфейс Implementor (низкоуровневая иерархия).
type Exporter interface {
	Export(data ReportData) error
}

// ==========================================
// 2. НАТИВНЫЕ РЕАЛИЗАЦИИ (Implementors)
// ==========================================

type CSVExporter struct{}

func (c *CSVExporter) Export(data ReportData) error {
	if len(data.Rows) == 0 {
		return ErrFormatInvalid
	}
	fmt.Printf("[CSV] Успешно экспортирован отчет '%s' (%d строк)\n", data.Title, len(data.Rows))
	return nil
}

type JSONExporter struct{}

func (j *JSONExporter) Export(data ReportData) error {
	if len(data.Rows) == 0 {
		return ErrFormatInvalid
	}
	fmt.Printf("[JSON] Успешно экспортирован отчет '%s' (%d строк)\n", data.Title, len(data.Rows))
	return nil
}

// ==========================================
// 3. НЕСОВМЕСТИМЫЙ LEGACY-КЛАСС (НЕ МЕНЯТЬ!)
// ==========================================

// LegacyXMLExporter - проприетарная библиотека.
// НЕСОВМЕСТИМОСТЬ:
//  1. Другое имя метода: WriteXML вместо Export
//  2. Другие параметры: (map[string]string, int) вместо (ReportData)
//  3. Другой механизм ошибок: возвращает int-коды вместо Go error
type LegacyXMLExporter struct{}

func (l *LegacyXMLExporter) WriteXML(data map[string]string, flags int) int {
	if len(data) == 0 {
		return -2 // код ошибки формата
	}
	if _, ok := data["simulate_error"]; ok {
		return -1 // код ошибки записи
	}
	fmt.Printf("[LegacyXML] Успешно записан XML (flags=%d)\n", flags)
	return 0 // успех
}

// ==========================================
// 4. АДАПТЕР (Adapter)
// ==========================================

type LegacyXMLAdapter struct {
	legacy *LegacyXMLExporter
}

func NewLegacyXMLAdapter(legacy *LegacyXMLExporter) *LegacyXMLAdapter {
	return &LegacyXMLAdapter{legacy: legacy}
}

// Export адаптирует вызов к legacy-методу.
func (a *LegacyXMLAdapter) Export(data ReportData) error {
	// 1. Трансляция параметров: ReportData -> map[string]string
	flat := make(map[string]string)
	if data.Title != "" {
		flat["title"] = data.Title
	}
	for i, row := range data.Rows {
		for k, v := range row {
			// Сохраняем оригинальный ключ (legacy ищет именно его)
			flat[k] = v
			// И ключ с префиксом строки (для уникальности)
			flat[fmt.Sprintf("row%d_%s", i, k)] = v
		}
	}

	// 2. Вызов несовместимого метода
	status := a.legacy.WriteXML(flat, 1)

	// 3. Трансляция ошибок: int -> Go error
	switch status {
	case 0:
		return nil
	case -1:
		return ErrExportFailed
	case -2:
		return ErrFormatInvalid
	default:
		return ErrExportFailed
	}
}

// ==========================================
// 5. АБСТРАКЦИИ (Bridge - высокоуровневая иерархия)
// ==========================================

// Report - базовая абстракция. Зависит ТОЛЬКО от интерфейса Exporter.
type Report struct {
	exporter Exporter
}

func (r *Report) SetExporter(e Exporter) {
	r.exporter = e
}

func (r *Report) Generate(data ReportData) (string, error) {
	err := r.exporter.Export(data)
	if err != nil {
		return "", err
	}
	return "Отчет сгенерирован: " + data.Title, nil
}

// SummaryReport - уточненная абстракция 1.
type SummaryReport struct {
	Report
}

func (s *SummaryReport) Generate(data ReportData) (string, error) {
	summaryData := ReportData{Title: data.Title + " (Summary)"}
	if len(data.Rows) > 0 {
		summaryData.Rows = data.Rows[:1]
	}

	result, err := s.Report.Generate(summaryData)
	if err != nil {
		return "", err
	}
	return "SUMMARY: " + result, nil
}

// DetailedReport - уточненная абстракция 2.
type DetailedReport struct {
	Report
}

func (d *DetailedReport) Generate(data ReportData) (string, error) {
	result, err := d.Report.Generate(data)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("DETAILED: %s [%d строк]", result, len(data.Rows)), nil
}

// ==========================================
// 6. МОДУЛЬ СЛОЖНОСТИ: Dynamic Implementor Selection
// ==========================================

// CreateExporter динамически выбирает реализацию в рантайме.
func CreateExporter(format string) Exporter {
	switch format {
	case "csv":
		return &CSVExporter{}
	case "json":
		return &JSONExporter{}
	case "xml":
		return NewLegacyXMLAdapter(&LegacyXMLExporter{})
	default:
		fmt.Printf("Неизвестный формат '%s', используется csv\n", format)
		return &CSVExporter{}
	}
}

// ==========================================
// 7. КЛИЕНТСКИЙ КОД (main)
// ==========================================

func main() {
	format := "csv"
	if len(os.Args) > 1 {
		format = os.Args[1]
	}

	exporter := CreateExporter(format)

	data := ReportData{
		Title: "Продажи за месяц",
		Rows: []map[string]string{
			{"product": "Телефон", "sales": "10"},
			{"product": "Ноутбук", "sales": "5"},
		},
	}

	summary := &SummaryReport{}
	summary.SetExporter(exporter)
	res, err := summary.Generate(data)
	if err != nil {
		fmt.Println("Ошибка Summary:", err)
	} else {
		fmt.Println(res)
	}

	detailed := &DetailedReport{}
	detailed.SetExporter(exporter)
	res, err = detailed.Generate(data)
	if err != nil {
		fmt.Println("Ошибка Detailed:", err)
	} else {
		fmt.Println(res)
	}
}
