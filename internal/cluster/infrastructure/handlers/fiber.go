package handlers

import (
	modelResponse "clusterMonitor/internal/application/domain/models"
	"clusterMonitor/internal/cluster/domain/ports"
	clientHttp "clusterMonitor/pkg/client/domain/ports"
	logrus "clusterMonitor/pkg/logging/domain/ports"
	schedulerPort "clusterMonitor/pkg/scheduler/domain/ports"
	"clusterMonitor/pkg/server"
	"github.com/gofiber/fiber/v2"
)

type ClusterHdl struct {
	networkInfo server.NetworkInfo
	logger      logrus.LoggerApplication
	scheduler   schedulerPort.Operation
	http        clientHttp.HttpMethod
}

func (hdl *ClusterHdl) Salutfa(ctx *fiber.Ctx) error {
	//TODO implement me
	panic("implement me")
}

func (hdl *ClusterHdl) SalutaDfiPiu(ctx *fiber.Ctx) error {
	//body := ctx.Body()

	//Log the raw body for debugging
	//log.Println("Raw Body:", string(body))

	// Parse JSON body into the MyRequest struct

	//get, err := hdl.http.Get()
	//if err != nil {
	//	return err
	//}
	//hdl.logger.Info(get.String())

	hdl.logger.Info("im here")

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

	return ctx.Status(fiber.StatusOK).JSON(nil)
}

//func (hdl *ClusterHdl) Check(ctx *fiber.Ctx) error {
//	//body := ctx.Body()
//
//	//Log the raw body for debugging
//	//log.Println("Raw Body:", string(body))
//
//	// Parse JSON body into the MyRequest struct
//
//	hdl.logger.Info("im here")
//
//	var requestBody modelResponse.RequestInfo
//	if err := ctx.BodyParser(&requestBody); err != nil {
//		//log.Println("Error parsing body:", err)
//		// Return a 400 Bad Request response if parsing fails
//		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
//			"error": "Invalid request body",
//		})
//	}
//
//	// Log the parsed request body
//	//log.Printf("Parsed Request: %+v", requestBody)
//
//	return ctx.Status(fiber.StatusOK).JSON(nil)
//}

var _ ports.HelloHandlers = (*ClusterHdl)(nil)

func NewClusterHandler(info server.NetworkInfo, logger logrus.LoggerApplication, httpCLient clientHttp.HttpMethod, schedulerHandler schedulerPort.Operation) *ClusterHdl {
	return &ClusterHdl{networkInfo: info, logger: logger, http: httpCLient, scheduler: schedulerHandler}
}
