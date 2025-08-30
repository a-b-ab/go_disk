package task

import (
	"os"
	"time"

	"go-cloud-disk/service/file"
	"go-cloud-disk/utils"
)

func DeleteLastDayFile() error {
	uploadDay := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	dst := utils.FastBuildString("./user/", uploadDay)
	err := os.Remove(dst)
	return err
}

// AutoCleanExpiredFiles 自动清理过期文件
func AutoCleanExpiredFiles() error {
	var service file.RecycleBinService
	return service.AutoCleanExpiredFiles()
}

// AutoCleanByCapacity 按容量自动清理
func AutoCleanByCapacity() error {
	var service file.RecycleBinService
	return service.AutoCleanByCapacity()
}
