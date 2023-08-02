package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func main() {
	router := gin.Default()

	router.GET("/test", DownloadExcel)

	router.Run(":9999")

	loop()
	// number := common.StringToNumber[int]("1")
	// a := int64(1)
	// strings := common.NumberToString(a)

	// ed, err := etcd.NewEtcd([]string{"127.0.0.1:2379"}, "", "")
	// if err != nil {
	// 	panic(err)
	// }
	// // data, err := ed.GetService("test", clientv3.WithPrefix())
	// // if err != nil {
	// // 	panic(err)
	// // }
	// reg, err := ed.NewServiceRegister(300)

	// if err != nil {
	// 	panic(err)
	// }

	// ed.ServiceRegister = reg

	// ed.Put("test999", "1000")

	// ed.PutService("test111", "10000")

	// fmt.Println("reg", reg)
	// time.Sleep(20 * time.Second)

}
func DownloadExcel(c *gin.Context) {
	xlsx := excelize.NewFile()
	xlsx.SetCellValue("Sheet1", "A1", "abc")
	xlsx.SetCellValue("Sheet1", "A2", "efg")
	//_ = xlsx.SaveAs("./aaa.xlsx")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+"test.xlsx")
	c.Header("Content-Transfer-Encoding", "binary")
	_ = xlsx.Write(c.Writer)
}
func loop() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT) //其中 SIGKILL = kill -9 <pid> 可能无法截获
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
			continue
		default:
			return
		}
	}
}
