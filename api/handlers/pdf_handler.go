package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	car_services_service "github.com/mil-ast/odo24_mobile_backend/api/services/car_services"
	groups_service "github.com/mil-ast/odo24_mobile_backend/api/services/groups"
	pdf_service "github.com/mil-ast/odo24_mobile_backend/api/services/pdf"
	"github.com/mil-ast/odo24_mobile_backend/api/utils"
)

type PdfController struct {
	groupsService      *groups_service.GroupsService
	carServicesService *car_services_service.CarServicesService
}

func NewPdfController(groupsSrv *groups_service.GroupsService, carServicesSrv *car_services_service.CarServicesService) *PdfController {
	return &PdfController{
		groupsService:      groupsSrv,
		carServicesService: carServicesSrv,
	}
}

// GetServiceBookPDF отдаёт PDF сервисной книжки авто.
func (ctrl *PdfController) GetServiceBookPDF(c *gin.Context) {
	userID := c.MustGet("userID").(uint64)
	carID := c.MustGet("carID").(uint64)

	groups, err := ctrl.groupsService.GetGroupsByUser(userID)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "ServiceBookPDFGroupsError", "Не удалось получить группы", err)
		return
	}

	services, err := ctrl.carServicesService.GetAllServicesByCar(carID)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "ServiceBookPDFServicesError", "Не удалось получить записи", err)
		return
	}

	pdfBytes, err := pdf_service.BuildServiceBook(groups, services)
	if err != nil {
		utils.BindServiceErrorWithAbort(c, "ServiceBookPDFBuildError", "Не удалось сформировать PDF", err)
		return
	}

	c.Header("Content-Disposition", `attachment; filename="service_book.pdf"`)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
