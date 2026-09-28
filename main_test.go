package main

import (
	"errors"
	"testing"
)

// MockExporter - заглушка для тестирования делегирования.
type MockExporter struct {
	ReturnErr error
	Called    bool
	LastData  ReportData
}

func (m *MockExporter) Export(data ReportData) error {
	m.Called = true
	m.LastData = data
	return m.ReturnErr
}

func TestSummaryReport_Delegation(t *testing.T) {
	mock := &MockExporter{}
	report := &SummaryReport{}
	report.SetExporter(mock)

	data := ReportData{
		Title: "Test",
		Rows:  []map[string]string{{"a": "1"}, {"b": "2"}},
	}

	result, err := report.Generate(data)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if !mock.Called {
		t.Fatal("экспортер не был вызван")
	}
	if len(mock.LastData.Rows) != 1 {
		t.Errorf("summary должен передавать 1 строку, получено %d", len(mock.LastData.Rows))
	}
	if result == "" {
		t.Error("ожидается непустой результат")
	}
}

func TestDetailedReport_Delegation(t *testing.T) {
	mock := &MockExporter{}
	report := &DetailedReport{}
	report.SetExporter(mock)

	data := ReportData{
		Title: "Test",
		Rows:  []map[string]string{{"a": "1"}, {"b": "2"}},
	}

	result, err := report.Generate(data)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if !mock.Called {
		t.Fatal("экспортер не был вызван")
	}
	if len(mock.LastData.Rows) != 2 {
		t.Errorf("detailed должен передавать все 2 строки, получено %d", len(mock.LastData.Rows))
	}
	if result == "" {
		t.Error("ожидается непустой результат")
	}
}

func TestReport_ExporterError(t *testing.T) {
	mock := &MockExporter{ReturnErr: ErrExportFailed}
	report := &SummaryReport{}
	report.SetExporter(mock)

	_, err := report.Generate(ReportData{Title: "T", Rows: []map[string]string{{"a": "1"}}})
	if !errors.Is(err, ErrExportFailed) {
		t.Errorf("ожидалась ErrExportFailed, получено %v", err)
	}
}

func TestAdapter_Success(t *testing.T) {
	adapter := NewLegacyXMLAdapter(&LegacyXMLExporter{})
	err := adapter.Export(ReportData{
		Title: "Test",
		Rows:  []map[string]string{{"key": "value"}},
	})
	if err != nil {
		t.Errorf("ожидался nil, получено %v", err)
	}
}

func TestAdapter_FormatError(t *testing.T) {
	adapter := NewLegacyXMLAdapter(&LegacyXMLExporter{})
	err := adapter.Export(ReportData{Title: "", Rows: nil})
	if !errors.Is(err, ErrFormatInvalid) {
		t.Errorf("ожидалась ErrFormatInvalid, получено %v", err)
	}
}

func TestAdapter_WriteError(t *testing.T) {
	adapter := NewLegacyXMLAdapter(&LegacyXMLExporter{})
	err := adapter.Export(ReportData{
		Title: "Test",
		Rows:  []map[string]string{{"simulate_error": "yes"}},
	})
	if !errors.Is(err, ErrExportFailed) {
		t.Errorf("ожидалась ErrExportFailed, получено %v", err)
	}
}
