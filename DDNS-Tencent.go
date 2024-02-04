package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

var (
	SecretId  string
	SecretKey string

	Domain     string
	SubDomain  string
	RecordType string
	RecordLine string
	// Value          string
	RecordId string

	CronExpression string
)

func init() {
	for _, value := range []string{"SecretId", "SecretKey", "Domain", "SubDomain", "RecordType", "RecordLine", "RecordId", "CronExpression"} {
		if os.Getenv(value) == "" {
			log.Panicf("%s is empty.", value)
		}
	}
}

func ModifyRecord(ipv6_addr string) {
	// 实例化一个认证对象，入参需要传入腾讯云账户 SecretId 和 SecretKey，此处还需注意密钥对的保密
	// 代码泄露可能会导致 SecretId 和 SecretKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考，建议采用更安全的方式来使用密钥，请参见：https://cloud.tencent.com/document/product/1278/85305
	// 密钥可前往官网控制台 https://console.cloud.tencent.com/cam/capi 进行获取
	credential := common.NewCredential(
		// "SecretId",
		os.Getenv("SecretId"),
		// "SecretKey",
		os.Getenv("SecretKey"),
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := dnspod.NewClient(credential, "", cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := dnspod.NewModifyRecordRequest()

	request.Domain = common.StringPtr(os.Getenv("Domain"))
	request.SubDomain = common.StringPtr(os.Getenv("SubDomain"))
	request.RecordType = common.StringPtr(os.Getenv("RecordType"))
	request.RecordLine = common.StringPtr(os.Getenv("RecordLine"))
	request.Value = common.StringPtr(ipv6_addr)
	ui64, err := strconv.ParseUint(os.Getenv("RecordId"), 10, 64)
	if err != nil {
		log.Panicln(err)
	}
	request.RecordId = common.Uint64Ptr(ui64)

	// 返回的resp是一个ModifyRecordResponse的实例，与请求对象对应
	response, err := client.ModifyRecord(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Printf("An API error has returned: %s\n", err)
		return
	}
	if err != nil {
		log.Panic(err)
	}
	// 输出json格式的字符串回包
	log.Printf("%s\n", response.ToJsonString())
}

func task() {
	ifaces, err := net.Interfaces()

	if err != nil {
		log.Panic(err)
	}

	ipv6 := ""

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			log.Panic(err)
		}
		for _, addr := range addrs {
			IPN := addr.(*net.IPNet)

			// fmt.Printf("IP Address: %v %v %v\n", addr, IPN.IP.IsPrivate(), IPN.IP.IsGlobalUnicast())
			if !IPN.IP.IsPrivate() && IPN.IP.IsGlobalUnicast() {
				// 如果是公网ip(非私有地址) 且是广播地址的网络
				log.Println("IPV6 Address: ", IPN.IP.To16().String())
				ipv6 = IPN.IP.To16().String()
			}
		}
		if ipv6 != "" {
			break
		}
	}

	ModifyRecord(ipv6)
}

func main() {
	// task()

	// 定时
	c := cron.New()
	_, err := c.AddFunc(os.Getenv("CronExpression"), func() {
		task()
	})

	if err != nil {
		log.Println("Add schedule task error.")
		log.Panic(err)
	}

	c.Start()

	for {
		time.Sleep(time.Second * 5)
	}
}
