package initialize

import (
	"SmartGraphite-server/global"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func MongoInit() {
	//数据储存在101服务器上
	if global.MongoClient101 == nil {
		global.MongoClient101 = getMongoClient("mongodb://admin:sdl%4020230228@101.42.233.226:27017")
	}
	//设备到后台管理去查
	graphiteManager := global.MongoClient101.Database("graphiteManager")
	{
		global.DeviceTypeColl = graphiteManager.Collection("deviceType")
		global.DeviceColl = graphiteManager.Collection("device")
		global.ElectricityUnitColl = graphiteManager.Collection("ElectricityUnit")
		//测试使用的数据库
		global.TestColl = graphiteManager.Collection("Test")
	}
	smartGraphite := global.MongoClient101.Database("smartGraphite")
	{
		//echarts
		global.EchartsColl = smartGraphite.Collection("Echarts")
		//设备解析动态配置
		//ba
		global.BATransducer = smartGraphite.Collection("BATransducer")
		global.BATransducerTenData = smartGraphite.Collection("BATransducerTenData")
		global.BATransducerThirtyData = smartGraphite.Collection("BATransducerThirtyData")
		global.BATransducerHourData = smartGraphite.Collection("BATransducerHourData")
		//bb
		global.BBTransducer = smartGraphite.Collection("BBTransducer")
		global.BBTransducerTenData = smartGraphite.Collection("BBTransducerTenData")
		global.BBTransducerThirtyData = smartGraphite.Collection("BBTransducerThirtyData")
		global.BBTransducerHourData = smartGraphite.Collection("BBTransducerHourData")
		//bc
		global.BCTransducer = smartGraphite.Collection("BCTransducer")
		global.BCTransducerTenData = smartGraphite.Collection("BCTransducerTenData")
		global.BCTransducerThirtyData = smartGraphite.Collection("BCTransducerThirtyData")
		global.BCTransducerHourData = smartGraphite.Collection("BCTransducerHourData")
		//变频器电流
		//D1
		global.D1Transducer = smartGraphite.Collection("D1Transducer")
		global.D1TransducerTenData = smartGraphite.Collection("D1TransducerTenData")
		global.D1TransducerThirtyData = smartGraphite.Collection("D1TransducerThirtyData")
		global.D1TransducerHourData = smartGraphite.Collection("D1TransducerHourData")
		//D2
		global.D2Transducer = smartGraphite.Collection("D2Transducer")
		global.D2TransducerTenData = smartGraphite.Collection("D2TransducerTenData")
		global.D2TransducerThirtyData = smartGraphite.Collection("D2TransducerThirtyData")
		global.D2TransducerHourData = smartGraphite.Collection("D2TransducerHourData")
		//D3
		global.D3Transducer = smartGraphite.Collection("D3Transducer")
		global.D3TransducerTenData = smartGraphite.Collection("D3TransducerTenData")
		global.D3TransducerThirtyData = smartGraphite.Collection("D3TransducerThirtyData")
		global.D3TransducerHourData = smartGraphite.Collection("D3TransducerHourData")
		//焙烧临时设备
		global.RoastTemp = smartGraphite.Collection("RoastTemp")
		global.E2RoastTemp = smartGraphite.Collection("E2RoastTemp")
		global.E2RoastTempMin = smartGraphite.Collection("E2RoastTempMin")
		global.E2RoastTempTM = smartGraphite.Collection("E2RoastTempTM")
		global.E2RoastTempReport = smartGraphite.Collection("E2RoastTempReport")
		global.E2RoastTempReportNew = smartGraphite.Collection("E2RoastTempReportNew")
		global.E2extrusionReport = smartGraphite.Collection("E2extrusionReport")

		//小压机设备
		global.E5RoastTemp = smartGraphite.Collection("E5RoastTemp")
		global.E5RoastTempMin = smartGraphite.Collection("E5RoastTempMin")
		global.E5RoastTempReport = smartGraphite.Collection("E5RoastTempReport")
		global.E5extrusionReport = smartGraphite.Collection("E5extrusionReport")

		//沥青温度
		global.E4RoastTemp = smartGraphite.Collection("E4RoastTemp")
		global.E4RoastTempMin = smartGraphite.Collection("E4RoastTempMin")
		//以上为动态解析的设备

		//焙烧脱硝
		global.RoastDenitrification = smartGraphite.Collection("RoastDenitrification")
		global.RoastDenitrificationHis = smartGraphite.Collection("RoastDenitrificationHis")
		global.RoastDenitrificationTen = smartGraphite.Collection("RoastDenitrificationTen")
		global.RoastDenitrificationThirty = smartGraphite.Collection("RoastDenitrificationThirty")
		global.RoastDenitrificationHour = smartGraphite.Collection("RoastDenitrificationHour")
		//煅烧脱硝
		global.CalDenitrification = smartGraphite.Collection("CalDenitrification")
		global.CalDenitrificationHis = smartGraphite.Collection("CalDenitrificationHis")
		global.CalDenitrificationTen = smartGraphite.Collection("CalDenitrificationTen")
		global.CalDenitrificationThirty = smartGraphite.Collection("CalDenitrificationThirty")
		global.CalDenitrificationHour = smartGraphite.Collection("CalDenitrificationHour")

		//三相电表
		global.ThreeElectricityDataColl = smartGraphite.Collection("ThreeElectricityData")
		global.ThreeElectricityHisDataColl = smartGraphite.Collection("ThreeElectricityHisData")
		global.ThreeElectricityTenMinDataColl = smartGraphite.Collection("ThreeElectricityTenMinData")
		global.ThreeElectricityThirtyMinDataColl = smartGraphite.Collection("ThreeElectricityThirtyMinData")
		global.ThreeElectricityHourDataColl = smartGraphite.Collection("ThreeElectricityHourData")
		global.ThreeElectricityECharsColl = smartGraphite.Collection("ThreeElectricityECharts")

		//成型plc
		global.FormPlcDataColl = smartGraphite.Collection("FormPlcData")
		global.FormPlcHisDataColl = smartGraphite.Collection("FormPlcHisData")
		global.FormPlcTenMinDataColl = smartGraphite.Collection("FormPlcTenMinData")
		global.FormPlcThirtyMinDataColl = smartGraphite.Collection("FormPlcThirtyMinData")
		global.FormPlcHourDataColl = smartGraphite.Collection("FormPlcHourData")
		//压型配料报表
		global.FormChargerReportColl = smartGraphite.Collection("FormChargerReport")
		global.FormChargerReportForwardColl = smartGraphite.Collection("FormChargerReportForward")
		global.FormChargerReportEndTime = smartGraphite.Collection("FormChargerReportEndTime")
		global.FormPlcEchartsColl = smartGraphite.Collection("FormPlcEcharts")
		//焙烧
		global.RoastDataColl = smartGraphite.Collection("RoastData")
		global.RoastHisDataColl = smartGraphite.Collection("RoastHisData")
		global.RoastTenMinDataColl = smartGraphite.Collection("RoastTenMinData")
		global.RoastThirtyDataColl = smartGraphite.Collection("RoastThirtyData")
		global.RoastHourDataColl = smartGraphite.Collection("RoastHourData")
		global.RoastEcahrtsColl = smartGraphite.Collection("RoastEcharts")

		//焙烧
		global.CRoastHisDataColl = smartGraphite.Collection("CRoastHisData")
		global.CRoastTenMinDataColl = smartGraphite.Collection("CRoastTenMinData")
		global.CRoastThirtyDataColl = smartGraphite.Collection("CRoastThirtyData")
		global.CRoastHourDataColl = smartGraphite.Collection("CRoastHourData")
		//浸渍
		global.DipDataColl = smartGraphite.Collection("DippingData")
		global.DipDataHisColl = smartGraphite.Collection("DippingHisData")
		global.DipTenMinDataColl = smartGraphite.Collection("DippingTenMinData")
		global.DipThirtyMinDataColl = smartGraphite.Collection("DippingThirtyMinData")
		global.DipHourDataColl = smartGraphite.Collection("DippingHourData")
		//隧道窑
		global.DTUDataColl = smartGraphite.Collection("DTUData")
		global.DTUHisDataColl = smartGraphite.Collection("DTUHisData")
		global.DTUTenMinDataColl = smartGraphite.Collection("DTUTenMinData")
		global.DTUThirtyMinDataColl = smartGraphite.Collection("DTUThirtyMinData")
		global.DTUHourDataColl = smartGraphite.Collection("DTUHourData")
		global.TunnelDataColl = smartGraphite.Collection("TunnelData")
		global.TunnelHisDataColl = smartGraphite.Collection("TunnelHisData")
		global.TunnelTenMinDataColl = smartGraphite.Collection("TunnelTenMinData")
		global.TunnelThirtyMinDataColl = smartGraphite.Collection("TunnelThirtyMinData")
		global.TunnelHourDataColl = smartGraphite.Collection("TunnelHourData")
		//石墨化
		global.GraphitingDataColl = smartGraphite.Collection("GraphitingData")
		global.GraphitingHisDataColl = smartGraphite.Collection("GraphitingHisData")
		global.GraphitinTenDataColl = smartGraphite.Collection("GraphitingTenData")
		global.GraphitingFiveDataColl = smartGraphite.Collection("GraphitingFiveData") //新增石墨化5分钟数据表
		global.GraphitinFifteenDataColl = smartGraphite.Collection("GraphitingFifteenData")
		global.GraphitingThirtyDataColl = smartGraphite.Collection("GraphitingThirtyData")
		global.GraphitingHourDataColl = smartGraphite.Collection("GraphitingHourData")
		global.GraPowerTimeColl = smartGraphite.Collection("GraPowerTime")                     //石墨化电量时刻数据
		global.GraElectricTimeColl = smartGraphite.Collection("GraElectricTime")               //石墨化有功功率时刻数据
		global.GraReportFormDataColl = smartGraphite.Collection("GraReportFormData")           //报表数据
		global.GraphitePlcReportColl = smartGraphite.Collection("GraphitePlcReport")           //石墨化PLC报表数据
		global.GraphitingdisplacementColl = smartGraphite.Collection("Graphitingdisplacement") //石墨化位移对照表
		//坩埚
		global.CrucibleDataColl = smartGraphite.Collection("CrucibleData")
		global.CrucibleHisDataColl = smartGraphite.Collection("CrucibleHisData")
		global.CrucibleTenMinDataColl = smartGraphite.Collection("CrucibleTenMinData")
		global.CrucibleThirtyDataColl = smartGraphite.Collection("CrucibleThirtyMinData")
		global.CrucibleHourDataColl = smartGraphite.Collection("CrucibleHourData")
		//吸料天车
		global.WestAirCarDataColl = smartGraphite.Collection("WestAirCarData") //西跨
		global.WestAirCarHisDataColl = smartGraphite.Collection("WestAirCarHisData")
		global.EastAirCarDataColl = smartGraphite.Collection("EastAirCarData") //东跨
		global.EastAirCarHisDataColl = smartGraphite.Collection("EastAirCarHisData")
		global.AirCarRunTimeColl = smartGraphite.Collection("AirCarRunTime") //运行时间
		//焙烧湿电
		global.RoastWetElectricDataColl = smartGraphite.Collection("RoastWetElectricData")
		global.RoastWetElectricHisDataColl = smartGraphite.Collection("RoastWetElectricHisData")
		global.RoastWetElectricTenDataColl = smartGraphite.Collection("RoastWetElectricTenData")
		global.RoastWetElectricThirtyDataColl = smartGraphite.Collection("RoastWetElectricThirtyData")
		global.RoastWetElectricHourDataColl = smartGraphite.Collection("RoastWetElectricHourData")
		//隧道窑湿电
		global.TunWetElectricDataColl = smartGraphite.Collection("TunnelWetElecData")
		global.TunWetElectricHisDataColl = smartGraphite.Collection("TunnelWetElecHisData")
		global.TunWetElectricTenDataColl = smartGraphite.Collection("TunnelWetElecTenData")
		global.TunWetElectricThirtyDataColl = smartGraphite.Collection("TunnelWetElecThirtyData")
		global.TunWetElectricHourDataColl = smartGraphite.Collection("TunnelWetElecHourData")
		//石墨化湿电
		global.GraWetElectricDataColl = smartGraphite.Collection("GraWetElecData")
		global.GraWetElectricHisDataColl = smartGraphite.Collection("GraWetElecHisData")
		global.GraWetElectricTenDataColl = smartGraphite.Collection("GraWetElecTenData")
		global.GraWetElectricThirtyDataColl = smartGraphite.Collection("GraWetElecThirtyData")
		global.GraWetElectricHourDataColl = smartGraphite.Collection("GraWetElecHourData")
		global.GraphiteOriginList = smartGraphite.Collection("GraphiteOriginListColl")
	}
	if global.MongoClient106 == nil {
		global.MongoClient106 = getMongoClient("mongodb://admin:sdl%40zzh20230228@106.52.170.16:27017")
	}
	smartGraphiteHB := global.MongoClient106.Database("smartGraphiteHB")
	{
		global.DataMinuteHisColl = smartGraphiteHB.Collection("dataMinuteHis")
	}
}
func getMongoClient(uri string) *mongo.Client {
	clientOptions := options.Client().ApplyURI(uri)

	MongoClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println(err)
	}
	if err = MongoClient.Ping(context.TODO(), nil); err != nil {
		log.Println(err)
	}
	return MongoClient
}
