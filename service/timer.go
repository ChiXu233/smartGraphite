package service

import (
	"fmt"
	"github.com/robfig/cron"
	"runtime"
	"time"
)

func OneMin() {
	E4RoastDataFilter()                 //e4分钟值生成
	E5RoastDataFilter()                 //e5分钟值生成
	E2RoastDataFilter()                 //e2
	WriteDenitrificationDataOperation() //焙烧，煅烧脱硝每分钟写入实时值
}

func FiveMin() {
	EchartsOperation("石墨化5分钟推送", "be67c2b8216e49e8981a95663413f115", "5") //石墨化位移变化量计算,5分钟间隔
}

func TenMin() {
	DataOperation(time.Minute * 10)            //Dtu
	DipDataOperation(time.Minute * 10)         //浸渍
	TunnelDataOperation(time.Minute * 10)      //隧道窑
	GraphitingDataOperation(time.Minute * 10)  //石墨化
	RoastingDataOperation(time.Minute * 10)    //焙烧湿电
	TunWetDataOperation(time.Minute * 10)      //隧道窑湿电
	GraWetDataOperation(time.Minute * 10)      //石墨化湿电
	RoastDataOperation(time.Minute * 10)       //焙烧温度
	FormPlcDataOperation(time.Minute * 10)     //成型PLC
	CruciblePlcDataOperation(time.Minute * 10) //坩埚PLC
	//FormChargerReportOperationForward()       //成型配料报表向前计算
	ThreeElectricityDataOperation(time.Minute * 10)     //三相电表
	DeviceDataOperation("642401d201972e9942398321", 10) //ba
	DeviceDataOperation("6424049f01972e9942398339", 10) //bb
	DeviceDataOperation("6423f6ba01972e994239829f", 10) //bc
	DeviceDataOperation("6426410bbda900f9bafd1f50", 10) //d1
	DeviceDataOperation("64264129bda900f9bafd1f54", 10) //d2
	DeviceDataOperation("642640c9bda900f9bafd1f49", 10) //d3
	DenitrificationDataOperation(10 * time.Minute)      //脱硝
	E2RoastGetTM()                                      //e2
}
func ThirtyMin() {
	DataOperation(time.Minute * 30)                     //Dtu
	DipDataOperation(time.Minute * 30)                  //浸渍
	TunnelDataOperation(time.Minute * 30)               //隧道窑
	GraphitingDataOperation(time.Minute * 30)           //石墨化
	RoastingDataOperation(time.Minute * 30)             //焙烧湿电
	TunWetDataOperation(time.Minute * 30)               //隧道窑湿电
	GraWetDataOperation(time.Minute * 30)               //石墨化湿电
	RoastDataOperation(time.Minute * 30)                //焙烧温度
	FormPlcDataOperation(time.Minute * 30)              //成型PLC
	CruciblePlcDataOperation(time.Minute * 30)          //坩埚PLC
	ThreeElectricityDataOperation(time.Minute * 30)     //三相电表
	DeviceDataOperation("642401d201972e9942398321", 30) //ba
	DeviceDataOperation("6424049f01972e9942398339", 30) //bb
	DeviceDataOperation("6423f6ba01972e994239829f", 30) //bc
	DeviceDataOperation("6426410bbda900f9bafd1f50", 30) //d1
	DeviceDataOperation("64264129bda900f9bafd1f54", 30) //d2
	DeviceDataOperation("642640c9bda900f9bafd1f49", 30) //d3
	DenitrificationDataOperation(30 * time.Minute)      //脱硝

}
func OneHour() {
	DataOperation(time.Hour)                            //Dtu
	DipDataOperation(time.Hour)                         //浸渍
	TunnelDataOperation(time.Hour)                      //隧道窑
	GraphitingDataOperation(time.Hour)                  //石墨化
	RoastingDataOperation(time.Hour)                    //焙烧湿电
	TunWetDataOperation(time.Hour)                      //隧道窑湿电
	GraWetDataOperation(time.Hour)                      //石墨化湿电
	AirCarDataOperation(time.Hour)                      //吸料天车运行时间数据
	RoastDataOperation(time.Hour)                       //焙烧温度
	FormPlcDataOperation(time.Hour)                     //成型PLC
	FormChargerReportOperation(time.Hour)               //压型配料报表
	FormPlcTrend()                                      //成型PLC图表预计算
	RoastEchartsTrend()                                 //焙烧温度图表
	CruciblePlcDataOperation(time.Hour)                 //坩埚PLC
	ThreeElectricityDataOperation(time.Hour)            //三相电表
	ThreeElectricityECharsOperation()                   //三相电表EChars
	DeviceDataOperation("642401d201972e9942398321", 60) //ba
	DeviceDataOperation("6424049f01972e9942398339", 60) //bb
	DeviceDataOperation("6423f6ba01972e994239829f", 60) //bc
	DeviceDataOperation("6426410bbda900f9bafd1f50", 60) //d1
	DeviceDataOperation("64264129bda900f9bafd1f54", 60) //d2
	DeviceDataOperation("642640c9bda900f9bafd1f49", 60) //d3
	//echarts
	DeviceEchartsOperation("642401d201972e9942398321", "平均值", 1, 23*60, 60)
	DeviceEchartsOperation("6424049f01972e9942398339", "平均值", 1, 23*60, 60)
	DeviceEchartsOperation("6423f6ba01972e994239829f", "平均值", 1, 23*60, 60)
	DeviceEchartsOperation("6426410bbda900f9bafd1f50", "平均值", 1, 23*60, 60)
	DeviceEchartsOperation("64264129bda900f9bafd1f54", "平均值", 1, 23*60, 60)
	DeviceEchartsOperation("642640c9bda900f9bafd1f49", "平均值", 1, 23*60, 60)
	DenitrificationDataOperation(time.Hour) //脱硝
	//E2RoastTempDelete()                     //e2 数据定时删除，保留一天的原始数据
	//E2RoastTempMinDelete()                  //e2 数据定时删除，保留一周的分钟数据
}
func FifteenMin() {
	GraphitingDataOperation(time.Minute * 15)
	EchartsOperation("石墨化15分钟推送", "be67c2b8216e49e8981a95663413f115", "15") //石墨化位移变化量计算,5分钟间隔
}

func CornTimer() {
	if runtime.GOOS != "linux" {
		return
	}
	fmt.Println("定时器开始工作")
	c := cron.New() //新建一个定时任务对象
	//1分钟一次
	_ = c.AddFunc("0 */1 * * * *", OneMin)
	//5分钟02秒一次
	_ = c.AddFunc("2/59 */5 * * * *", FiveMin)
	//10分钟一次
	_ = c.AddFunc("0 */10 * * * *", TenMin)
	//15分钟存一次最新 目前只有石墨化用
	_ = c.AddFunc("0 */15 * * * *", FifteenMin)
	//30分钟一次
	_ = c.AddFunc("0 */30 * * * *", ThirtyMin)
	//1小时一次
	_ = c.AddFunc("0 0 * * * *", OneHour)
	c.Start() //开始

	select {} //阻塞住,保持程序运行
}
