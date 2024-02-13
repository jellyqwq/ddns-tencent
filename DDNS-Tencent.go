package main

import (
	"fmt"
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

// 定义一个type
type DDNSproperty struct {
	Domain     string
	SubDomain  string
	RecordType string
	RecordLine string
	Value      string
	RecordId   string
}

var (
	SecretId       string
	SecretKey      string
	CronExpression string = "*/5 * * * *"

	ipv4Property *DDNSproperty
	ipv6Property *DDNSproperty
)

func init() {
	// 先检查有无 SecretId和SecretKey, 没有则退出, 有则写入全局变量
	for _, value := range []string{"SecretId", "SecretKey"} {
		tempVar := os.Getenv(value)
		if tempVar == "" {
			log.Panicf("%s cannot be empty.", value)
		}
		if value == "SecretId" {
			SecretId = tempVar
		} else {
			SecretKey = tempVar
		}
	}

	// 是否需要覆盖定时执行
	tempVar := os.Getenv("CronExpression")
	if tempVar != "" {
		CronExpression = tempVar
	}

	// 检查是否有 Ipv<4|6>Domain 参数, 没有则退出
	for _, value := range []string{"Ipv4", "Ipv6"} {
		if value == "Ipv4" {
			ipv4Property = NewDDNSproperty()
			ipv4Property.Domain = os.Getenv(fmt.Sprintf("%sDomain", value))
			if ipv4Property.Domain == "" {
				log.Printf("Skip %s ", value)
				continue
			}
			ipv4Property.SubDomain = os.Getenv(fmt.Sprintf("%sSubDomain", value))
			if os.Getenv(fmt.Sprintf("%sRecordLine", value)) != "" {
				ipv4Property.RecordLine = os.Getenv(fmt.Sprintf("%sRecordLine", value))
			}
			ipv4Property.RecordType = "A"
			ipv4Property.RecordId = os.Getenv(fmt.Sprintf("%sRecordId", value))
		} else {
			ipv6Property = NewDDNSproperty()
			ipv6Property.Domain = os.Getenv(fmt.Sprintf("%sDomain", value))
			if ipv6Property.Domain == "" {
				log.Printf("Skip %s ", value)
				continue
			}
			ipv6Property.SubDomain = os.Getenv(fmt.Sprintf("%sSubDomain", value))
			if os.Getenv(fmt.Sprintf("%sRecordLine", value)) != "" {
				ipv6Property.RecordLine = os.Getenv(fmt.Sprintf("%sRecordLine", value))
			}
			ipv6Property.RecordType = "AAAA"
			ipv6Property.RecordId = os.Getenv(fmt.Sprintf("%sRecordId", value))
		}
	}

	if ipv4Property.Domain == "" && ipv6Property.Domain == "" {
		log.Panic("No service need to run.")
	}
}

func NewDDNSproperty() *DDNSproperty {
	return &DDNSproperty{
		RecordLine: "默认",
	}
}

func (property *DDNSproperty) ModifyRecord() {
	// 实例化一个认证对象，入参需要传入腾讯云账户 SecretId 和 SecretKey，此处还需注意密钥对的保密
	// 代码泄露可能会导致 SecretId 和 SecretKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考，建议采用更安全的方式来使用密钥，请参见：https://cloud.tencent.com/document/product/1278/85305
	// 密钥可前往官网控制台 https://console.cloud.tencent.com/cam/capi 进行获取
	credential := common.NewCredential(
		SecretId,
		SecretKey,
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := dnspod.NewClient(credential, "", cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := dnspod.NewModifyRecordRequest()

	request.Domain = common.StringPtr(property.Domain)
	request.SubDomain = common.StringPtr(property.SubDomain)
	request.RecordType = common.StringPtr(property.RecordType)
	request.RecordLine = common.StringPtr(property.RecordLine)
	request.Value = common.StringPtr(property.Value)
	ui64, err := strconv.ParseUint(property.RecordId, 10, 64)
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

	ipv4 := ""
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
				if IPN.IP.To4() != nil {
					ipv4 = IPN.IP.To4().String()
					if ipv4Property.Domain != "" && ipv6Property.Domain == "" {
						log.Println("IPV4 Address: ", ipv4)
						ipv4Property.Value = ipv4
						ipv4Property.ModifyRecord()
						return
					}
				} else if IPN.IP.To16() != nil {
					ipv6 = IPN.IP.To16().String()
					if ipv6Property.Domain != "" && ipv4Property.Domain == "" {
						log.Println("IPV6 Address: ", ipv6)
						ipv6Property.Value = ipv6
						ipv6Property.ModifyRecord()
						return
					}
				}
			}
		}

		// 两个都找到了提前结束寻找
		if ipv4 != "" && ipv6 != "" {
			log.Println("IPV4 Address: ", ipv4)
			ipv4Property.Value = ipv4
			ipv4Property.ModifyRecord()

			log.Println("IPV6 Address: ", ipv6)
			ipv6Property.Value = ipv6
			ipv6Property.ModifyRecord()

			return
		}
	}
}

func main() {
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
