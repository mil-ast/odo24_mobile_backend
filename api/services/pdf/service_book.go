package pdf_service

import (
	"bytes"
	_ "embed"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	car_services_service "github.com/mil-ast/odo24_mobile_backend/api/services/car_services"
	groups_service "github.com/mil-ast/odo24_mobile_backend/api/services/groups"
)

//go:embed fonts/DejaVuSans.ttf
var fontRegular []byte

//go:embed fonts/DejaVuSans-Bold.ttf
var fontBold []byte

const (
	fontFamily = "DejaVu"
	siteURL    = "https://odo24.ru"
	bookTitle  = "Сервисная книжка"
)

// Геометрия страницы A4 (мм).
const (
	marginX = 15.0
	marginY = 15.0

	tableWidth = 180.0
	colDate    = 40.0
	colOdo     = 65.0
	colPrice   = 75.0

	headerHeight  = 9.0
	cellHeight    = 8.0
	commentHeight = 7.0
)

// BuildServiceBook собирает PDF сервисной книжки:
// титульная страница, затем по каждой группе — таблица её сервисов
// и страница с пустыми строками для ручного заполнения.
func BuildServiceBook(groups []groups_service.GroupModel, services []car_services_service.CarServiceRow) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(marginX, marginY, marginX)
	pdf.SetAutoPageBreak(true, marginY)
	pdf.SetCellMargin(2)
	pdf.AddUTF8FontFromBytes(fontFamily, "", fontRegular)
	pdf.AddUTF8FontFromBytes(fontFamily, "B", fontBold)

	drawTitlePage(pdf)

	servicesByGroup := make(map[uint64][]car_services_service.CarServiceRow)
	for _, service := range services {
		servicesByGroup[service.GroupID] = append(servicesByGroup[service.GroupID], service)
	}

	groups = sortedGroups(groups)

	for _, group := range groups {
		groupServices := servicesByGroup[group.GroupID]

		pdf.AddPage()
		drawGroupHeading(pdf, group.Name)

		if len(groupServices) == 0 {
			pdf.SetFont(fontFamily, "", 10)
			pdf.SetTextColor(120, 120, 120)
			pdf.CellFormat(tableWidth, cellHeight, "Записей нет", "", 1, "L", false, 0, "")
			pdf.SetTextColor(0, 0, 0)
			continue
		}

		drawTableHeader(pdf)
		for _, service := range groupServices {
			drawServiceRow(pdf, service)
		}

		drawBlankPage(pdf, group.Name)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func sortedGroups(groups []groups_service.GroupModel) []groups_service.GroupModel {
	result := make([]groups_service.GroupModel, len(groups))
	copy(result, groups)

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Sort == result[j].Sort {
			return result[i].GroupID < result[j].GroupID
		}
		return result[i].Sort < result[j].Sort
	})

	return result
}

func drawTitlePage(pdf *fpdf.Fpdf) {
	pdf.AddPage()

	pdf.SetFont(fontFamily, "B", 34)
	pdf.SetXY(marginX, 90)
	pdf.CellFormat(tableWidth, 20, bookTitle, "", 1, "C", false, 0, "")

	pdf.SetDrawColor(160, 160, 160)
	pdf.Line(marginX+60, 125, marginX+tableWidth-60, 125)

	pdf.SetFont(fontFamily, "", 12)
	pdf.SetTextColor(90, 90, 90)
	_, pageHeight := pdf.GetPageSize()
	pdf.SetXY(marginX, pageHeight-25)
	pdf.CellFormat(tableWidth, 10, siteURL, "", 1, "C", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(0, 0, 0)
}

func drawGroupHeading(pdf *fpdf.Fpdf, name string) {
	pdf.SetFont(fontFamily, "B", 14)
	pdf.MultiCell(tableWidth, 8, name, "", "L", false)

	y := pdf.GetY() + 1
	pdf.SetDrawColor(160, 160, 160)
	pdf.Line(marginX, y, marginX+tableWidth, y)
	pdf.SetDrawColor(0, 0, 0)

	pdf.SetY(y + 3)
	pdf.SetFont(fontFamily, "", 10)
}

func drawTableHeader(pdf *fpdf.Fpdf) {
	pdf.SetFont(fontFamily, "B", 10)
	pdf.SetFillColor(235, 235, 235)

	pdf.CellFormat(colDate, headerHeight, "Дата", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colOdo, headerHeight, "Пробег, км", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colPrice, headerHeight, "Цена, руб.", "1", 1, "C", true, 0, "")

	pdf.SetFillColor(255, 255, 255)
	pdf.SetFont(fontFamily, "", 10)
}

func drawServiceRow(pdf *fpdf.Fpdf, service car_services_service.CarServiceRow) {
	comment := ""
	if service.Description != nil {
		comment = strings.TrimSpace(*service.Description)
	}

	rowHeight := cellHeight + commentRowHeight(pdf, comment)

	// Строка не помещается — переносим на новую страницу и повторяем шапку таблицы.
	if pdf.GetY()+rowHeight > maxY(pdf) {
		pdf.AddPage()
		drawTableHeader(pdf)
	}

	x := marginX
	y := pdf.GetY()

	pdf.SetXY(x, y)
	pdf.CellFormat(colDate, cellHeight, formatDate(service.Dt), "1", 0, "C", false, 0, "")
	pdf.CellFormat(colOdo, cellHeight, formatNumber(service.Odo), "1", 0, "C", false, 0, "")
	pdf.CellFormat(colPrice, cellHeight, formatNumber(service.Price), "1", 0, "C", false, 0, "")

	drawCommentRow(pdf, x, y+cellHeight, comment)
}

func drawBlankPage(pdf *fpdf.Fpdf, groupName string) {
	pdf.AddPage()
	drawGroupHeading(pdf, groupName)
	drawTableHeader(pdf)

	for pdf.GetY()+cellHeight+commentHeight <= maxY(pdf) {
		y := pdf.GetY()
		pdf.SetXY(marginX, y)
		pdf.CellFormat(colDate, cellHeight, "", "1", 0, "C", false, 0, "")
		pdf.CellFormat(colOdo, cellHeight, "", "1", 0, "C", false, 0, "")
		pdf.CellFormat(colPrice, cellHeight, "", "1", 0, "C", false, 0, "")
		drawCommentRow(pdf, marginX, y+cellHeight, "")
	}
}

func drawCommentRow(pdf *fpdf.Fpdf, x, y float64, comment string) {
	pdf.SetFont(fontFamily, "", 9)
	pdf.SetXY(x, y)

	if comment == "" {
		pdf.CellFormat(tableWidth, commentHeight, "", "LRB", 1, "L", false, 0, "")
	} else {
		pdf.MultiCell(tableWidth, commentHeight, comment, "LRB", "L", false)
	}

	pdf.SetFont(fontFamily, "", 10)
}

func commentRowHeight(pdf *fpdf.Fpdf, comment string) float64 {
	if comment == "" {
		return commentHeight
	}

	lines := pdf.SplitLines([]byte(comment), tableWidth-2*pdf.GetCellMargin())
	return math.Max(commentHeight, float64(len(lines))*commentHeight)
}

func maxY(pdf *fpdf.Fpdf) float64 {
	_, pageHeight := pdf.GetPageSize()
	return pageHeight - marginY
}

func formatDate(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	layouts := []string{
		"2006-01-02",
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"02.01.2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t.Format("02.01.2006")
		}
	}

	return value
}

func formatNumber(value *uint32) string {
	if value == nil {
		return ""
	}

	digits := strconv.FormatUint(uint64(*value), 10)

	var b strings.Builder
	for i, digit := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(digit)
	}

	return b.String()
}
