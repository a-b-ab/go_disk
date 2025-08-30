package task

import (
	"time"

	"go-cloud-disk/utils/logger"

	"github.com/robfig/cron/v3"
)

// Cron 全局定时任务实例
var Cron *cron.Cron

// jobFunc 定时任务函数类型
type jobFunc func() error

// Run 运行任务并打印任务执行结果
func Run(jobName string, job jobFunc) {
	// 计算任务执行时间
	from := time.Now().UnixNano()
	err := job()
	to := time.Now().UnixNano()
	if err != nil {
		logger.Log().Error("%s 执行失败: %dms\n 错误:%v", jobName, (to-from)/int64(time.Millisecond), err)
	} else {
		logger.Log().Info("%s 执行成功: %dms\n", jobName, (to-from)/int64(time.Millisecond))
	}
}

// CronJob 启动定时任务
func CronJob() {
	if Cron == nil {
		Cron = cron.New()
	}

	// 每天凌晨0点重置日排行榜
	if _, err := Cron.AddFunc("@daily", func() { Run("重置日排行榜", RestartDailyRank) }); err != nil {
		logger.Log().Error("设置重置日排行榜任务失败", err)
	}
	// 每天凌晨1点删除昨日文件
	if _, err := Cron.AddFunc("0 1 * * *", func() { Run("删除昨日文件", DeleteLastDayFile) }); err != nil {
		logger.Log().Error("设置删除昨日文件任务失败", err)
	}
	// 每天凌晨2点自动清理过期文件
	if _, err := Cron.AddFunc("0 2 * * *", func() { Run("自动清理过期文件", AutoCleanExpiredFiles) }); err != nil {
		logger.Log().Error("设置自动清理过期文件任务失败", err)
	}
	// 每小时清理超容量文件
	if _, err := Cron.AddFunc("@hourly", func() { Run("按容量自动清理", AutoCleanByCapacity) }); err != nil {
		logger.Log().Error("设置按容量自动清理任务失败", err)
	}

	Cron.Start()
}
