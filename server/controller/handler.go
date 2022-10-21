package controller

import (
	"log"
	"net"
	"net/http"
	"server/entity"
	"server/libs"
	"server/listener"
	"server/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (ct *Controller) getLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func (ct *Controller) postAuth(c *gin.Context) {
	var user libs.UserInfo
	err := c.ShouldBind(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 2001,
			"msg":  "无效的参数",
		})
		return
	}
	if user.Username == ct.Options.UserInfo.Username && user.Password == ct.Options.UserInfo.Password {
		// 生成Token
		tokenString, err := libs.GenToken(user.Username)
		c.SetCookie("Authorization", tokenString, 3600, "/", "*", false, true)
		if err != nil {
			log.Println(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 2000,
			"msg":  "success",
			"data": gin.H{"token": tokenString},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 2002,
		"msg":  "鉴权失败",
	})

}

func (ct *Controller) setDevice(c *gin.Context) {
	var body libs.ReceiveClient
	var entityDevice entity.Device
	if err := c.BindJSON(&body); err != nil {
		log.Println(err)
	}
	log.Println(body)

	//客户端获取到的Laddr ip 只有一个
	if len(body.IPs) == 1 {
		entityDevice.ClientIP = body.IPs[0]
		//判断这个Ip是否在资产列表中
		if utils.InSlice(ct.Options.NodeIPs, entityDevice.ClientIP) {
			entityDevice.NodeIP = entityDevice.ClientIP
		}
	}

	entityDevice.Hostname = body.Hostname
	entityDevice.MacAddress = body.MacAddress
	// body.RemoteIP = c.

	ct.DeviceService.Insert(entityDevice)

}

func (ct *Controller) getDevices(c *gin.Context) {
	devices := ct.DeviceService.FindAll()
	log.Println(devices)
	c.HTML(http.StatusOK, "devices.html", gin.H{
		"Devices": devices,
	})

}

func (ct *Controller) getResources(c *gin.Context) {
	resources := ct.Options.Resources
	var results []entity.Resources
	for _, resource := range resources {
		for _, ip := range resource.IP {
			var result entity.Resources
			result.ClusterName = resource.Name
			result.NodeIP = ip
			t := ct.DeviceService.GetUpdateTime(ip)
			result.DaemonUpdate = t.Format("2006-01-02 15:04:05")

			t2 := ct.FalcoService.GetUpdateTime(ip)
			result.FalcoUpdate = t2.Format("2006-01-02 15:04:05")

			results = append(results, result)
		}

	}

	log.Println(results)
	c.HTML(http.StatusOK, "resources.html", gin.H{
		"Resources": results,
	})
}

func (ct *Controller) resourcesUpdate(c *gin.Context) {
	//初始化资产
	resources := ct.Options.Resources
	var entyResource entity.Resource
	for _, resource := range resources {
		for _, ip := range resource.IP {
			entyResource.ClusterName = resource.Name
			entyResource.NodeIP = ip
			ct.ResourceService.Insert(entyResource)
		}

	}
}

func (ct *Controller) setFalco(c *gin.Context) {
	var body entity.ParseHttpFalco
	var entityFalco entity.Falco
	if err := c.BindJSON(&body); err != nil {
		log.Println(err)
	}

	//客户端获取到的Laddr ip 只有一个
	if len(body.IPs) == 1 {
		//判断这个Ip是否在资产列表中
		if utils.InSlice(ct.Options.NodeIPs, body.IPs[0]) {
			entityFalco.NodeIP = body.IPs[0]
		}
	}

	entityFalco.Falco = body.Falco

	ct.FalcoService.Insert(entityFalco)

}

func (ct *Controller) restartFalco(c *gin.Context) {
	ip := c.Query("nodeip")
	command := "falco restart"
	var conn net.Conn
	if utils.InSlice(ct.Options.NodeIPs, ip) {
		if ct.Listenner.Connlist[ip] != nil {
			conn = ct.Listenner.Connlist[ip]
			listener.RunComand(command, conn)
		}
	}
	time.Sleep(time.Second * 2)
	if conn != nil {
		listener.ReadMessage(conn)
	}

	// var body entity.ParseHttpFalco
	// var entityFalco entity.Falco
	// if err := c.BindJSON(&body); err != nil {
	// 	log.Println(err)
	// }

	// //客户端获取到的Laddr ip 只有一个
	// if len(body.IPs) == 1 {
	// 	//判断这个Ip是否在资产列表中
	// 	if utils.InSlice(ct.Options.NodeIPs, body.IPs[0]) {
	// 		entityFalco.NodeIP = body.IPs[0]
	// 	}
	// }

	// entityFalco.Falco = body.Falco

	// ct.FalcoService.Insert(entityFalco)

}

func (ct *Controller) getGenerate(c *gin.Context) {
	c.HTML(http.StatusOK, "generate.html", gin.H{
		"Address":  ct.Options.ServerAddress,
		"Port":     strings.ReplaceAll(ct.Options.ServerPort, ":", ""),
		"Filename": ct.Options.ClientName,
	})
}

func (ct *Controller) generateClient(c *gin.Context) {

	log.Println(c.PostForm("filename"))
	var clientGenerate entity.GenerateClient
	if err := c.ShouldBind(&clientGenerate); err != nil {
		log.Println(err)
	}
	log.Println(clientGenerate)

}

// func (ct *Controller) getDevicesHandler(c *gin.Context) {
// 	devices, err := ct.DeviceService.FindAll()
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	c.HTML(http.StatusOK, "devices.html", gin.H{
// 		"Devices": devices,
// 	})
// }
