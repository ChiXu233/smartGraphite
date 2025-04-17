package global

import "go.mongodb.org/mongo-driver/mongo"

var (
	MongoClient106                    *mongo.Client
	MongoClient101                    *mongo.Client
	DeviceTypeColl                    *mongo.Collection
	DeviceColl                        *mongo.Collection
	DipDataColl                       *mongo.Collection //浸渍表
	DipDataHisColl                    *mongo.Collection
	DipTenMinDataColl                 *mongo.Collection
	DipThirtyMinDataColl              *mongo.Collection
	DipHourDataColl                   *mongo.Collection
	DTUDataColl                       *mongo.Collection //隧道窑DTU
	DTUHisDataColl                    *mongo.Collection
	DTUTenMinDataColl                 *mongo.Collection
	DTUThirtyMinDataColl              *mongo.Collection
	DTUHourDataColl                   *mongo.Collection
	TunnelDataColl                    *mongo.Collection //隧道窑
	TunnelHisDataColl                 *mongo.Collection
	TunnelTenMinDataColl              *mongo.Collection
	TunnelThirtyMinDataColl           *mongo.Collection
	TunnelHourDataColl                *mongo.Collection
	GraphitingDataColl                *mongo.Collection //石墨化
	GraphitingdisplacementColl        *mongo.Collection //石墨化位移对照表
	GraphitingHisDataColl             *mongo.Collection
	GraphitingFiveDataColl            *mongo.Collection //石墨5分钟数据计算
	GraphitinTenDataColl              *mongo.Collection
	GraphitinFifteenDataColl          *mongo.Collection
	GraphitingThirtyDataColl          *mongo.Collection
	GraphitingHourDataColl            *mongo.Collection
	GraPowerTimeColl                  *mongo.Collection
	GraElectricTimeColl               *mongo.Collection
	GraReportFormDataColl             *mongo.Collection //石墨化报表(根据判断有功功率生成)
	GraphitePlcReportColl             *mongo.Collection //石墨化生产工艺报表(根据有功电量时刻值生成)
	CrucibleDataColl                  *mongo.Collection //坩埚
	CrucibleHisDataColl               *mongo.Collection
	CrucibleTenMinDataColl            *mongo.Collection //坩埚10分钟数据
	CrucibleThirtyDataColl            *mongo.Collection //坩埚30分钟数据
	CrucibleHourDataColl              *mongo.Collection //坩埚小时数据
	WestAirCarDataColl                *mongo.Collection //西跨吸料天车
	WestAirCarHisDataColl             *mongo.Collection
	EastAirCarDataColl                *mongo.Collection //东跨吸料天车
	EastAirCarHisDataColl             *mongo.Collection
	AirCarRunTimeColl                 *mongo.Collection //吸料天车运行时间表
	TunWetElectricDataColl            *mongo.Collection //隧道窑湿电
	TunWetElectricHisDataColl         *mongo.Collection
	TunWetElectricTenDataColl         *mongo.Collection
	TunWetElectricThirtyDataColl      *mongo.Collection
	TunWetElectricHourDataColl        *mongo.Collection
	RoastWetElectricDataColl          *mongo.Collection //焙烧湿电
	RoastWetElectricHisDataColl       *mongo.Collection
	RoastWetElectricTenDataColl       *mongo.Collection
	RoastWetElectricThirtyDataColl    *mongo.Collection
	RoastWetElectricHourDataColl      *mongo.Collection
	GraWetElectricDataColl            *mongo.Collection //石墨化湿电
	GraWetElectricHisDataColl         *mongo.Collection
	GraWetElectricTenDataColl         *mongo.Collection
	GraWetElectricThirtyDataColl      *mongo.Collection
	GraWetElectricHourDataColl        *mongo.Collection
	RoastDataColl                     *mongo.Collection //焙烧
	RoastHisDataColl                  *mongo.Collection //历史数据
	RoastTenMinDataColl               *mongo.Collection //10分钟焙烧温度数据
	RoastThirtyDataColl               *mongo.Collection //30分钟焙烧温度数据
	RoastHourDataColl                 *mongo.Collection //小时焙烧温度数据
	FormPlcDataColl                   *mongo.Collection //成型plc
	FormPlcHisDataColl                *mongo.Collection
	FormPlcTenMinDataColl             *mongo.Collection //10分钟成型plc数据
	FormPlcThirtyMinDataColl          *mongo.Collection //30分钟成型plc数据
	FormPlcHourDataColl               *mongo.Collection //小时成型plc数据
	FormChargerReportColl             *mongo.Collection //压型配料报表
	FormChargerReportForwardColl      *mongo.Collection //压型配料向前表
	FormChargerReportEndTime          *mongo.Collection //压型配料报表每次计算的时间,用来查找之前的数据
	FormPlcEchartsColl                *mongo.Collection //成型PLC图表
	RoastEcahrtsColl                  *mongo.Collection //焙烧图表
	ThreeElectricityDataColl          *mongo.Collection //三相电表实时
	ThreeElectricityHisDataColl       *mongo.Collection //三相电表历史数据
	ThreeElectricityTenMinDataColl    *mongo.Collection //三相电表10分钟数据
	ThreeElectricityThirtyMinDataColl *mongo.Collection //三相电表30分钟数据
	ThreeElectricityHourDataColl      *mongo.Collection //三相电表小时数据
	ThreeElectricityECharsColl        *mongo.Collection
	ElectricityUnitColl               *mongo.Collection //记录三相电表的单位转换
	EchartsColl                       *mongo.Collection //echarts
	TestColl                          *mongo.Collection //测试使用的数据表
	BBTransducer                      *mongo.Collection //2号焙烧变频器
	BBTransducerTenData               *mongo.Collection
	BBTransducerThirtyData            *mongo.Collection
	BBTransducerHourData              *mongo.Collection
	BATransducer                      *mongo.Collection //1号焙烧变频器
	BATransducerTenData               *mongo.Collection
	BATransducerThirtyData            *mongo.Collection
	BATransducerHourData              *mongo.Collection
	BCTransducer                      *mongo.Collection //3号焙烧变频器
	BCTransducerTenData               *mongo.Collection
	BCTransducerThirtyData            *mongo.Collection
	BCTransducerHourData              *mongo.Collection
	D3Transducer                      *mongo.Collection //4G变频器电流,名开始起错了，不想改了，电表,D1,D3都是
	D3TransducerTenData               *mongo.Collection
	D3TransducerThirtyData            *mongo.Collection
	D3TransducerHourData              *mongo.Collection
	D2Transducer                      *mongo.Collection //4G变频器电流
	D2TransducerTenData               *mongo.Collection
	D2TransducerThirtyData            *mongo.Collection
	D2TransducerHourData              *mongo.Collection
	D1Transducer                      *mongo.Collection //4G变频器电流
	D1TransducerTenData               *mongo.Collection
	D1TransducerThirtyData            *mongo.Collection
	D1TransducerHourData              *mongo.Collection
	RoastDenitrification              *mongo.Collection //焙烧脱硝
	RoastDenitrificationHis           *mongo.Collection
	RoastDenitrificationTen           *mongo.Collection
	RoastDenitrificationThirty        *mongo.Collection
	RoastDenitrificationHour          *mongo.Collection
	CalDenitrification                *mongo.Collection //煅烧脱硝
	CalDenitrificationHis             *mongo.Collection
	CalDenitrificationTen             *mongo.Collection
	CalDenitrificationThirty          *mongo.Collection
	CalDenitrificationHour            *mongo.Collection
	RoastTemp                         *mongo.Collection //焙烧临时设备
	E2RoastTemp                       *mongo.Collection
	E2RoastTempMin                    *mongo.Collection //此设备特殊
	E2RoastTempTM                     *mongo.Collection
	E2RoastTempReport                 *mongo.Collection
	E2RoastTempReportNew              *mongo.Collection
	E2extrusionReport                 *mongo.Collection //挤压报表
	E5RoastTemp                       *mongo.Collection //小压机设备
	E5RoastTempMin                    *mongo.Collection
	E5RoastTempReport                 *mongo.Collection //小成型晾料报表生成
	E5extrusionReport                 *mongo.Collection
	E4RoastTemp                       *mongo.Collection //沥青温度
	E4RoastTempMin                    *mongo.Collection

	CRoastHisDataColl    *mongo.Collection //焙烧历史数据
	CRoastTenMinDataColl *mongo.Collection //焙烧10分钟焙烧温度数据
	CRoastThirtyDataColl *mongo.Collection //焙烧30分钟焙烧温度数据
	CRoastHourDataColl   *mongo.Collection //焙烧小时焙烧温度数据

	//焙烧CEMS读取数据
	DataMinuteHisColl *mongo.Collection

	//石墨化原料配料表
	GraphiteOriginList *mongo.Collection //石墨化配料表
)

// plc用
var Token string
var Spec string

//
////plc用
//var Token2 string
//var Spec2 string
