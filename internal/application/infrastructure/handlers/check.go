package handlers

import (
	modelResponse "clusterMonitor/internal/application/domain/models"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
)

// Register godoc
// @Summary Register a new user
// @Description Register a new user
// @Accept json
// @Produce json
// @Tags Application
// @Param body body modelResponse.RequestInfo true "User registration information"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /api/v1/application/check [post]
func (hdl *ApplicationHdl) Check(ctx *fiber.Ctx) error {
	//body := ctx.Body()

	//Log the raw body for debugging
	//log.Println("Raw Body:", string(body))

	// Parse JSON body into the MyRequest struct
	var requestBody modelResponse.RequestInfo
	if err := ctx.BodyParser(&requestBody); err != nil {
		//log.Println("Error parsing body:", err)
		// Return a 400 Bad Request response if parsing fails
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Log the parsed request body
	//log.Printf("Parsed Request: %+v", requestBody)

	_, _ = hdl.service.CheckRequest(&requestBody)
	response, err := hdl.service.ExecRequest(&requestBody)

	if err != nil {
		//log.Println("Error parsing body:", err)
		// Return a 400 Bad Request response if parsing fails

		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(response.Body(), &jsonResponse); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse JSON response",
		})
	}

	// Restituisci il JSON decodificato
	return ctx.Status(fiber.StatusOK).JSON(jsonResponse)
}
