package xlsx

import (
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

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
