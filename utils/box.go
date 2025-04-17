package utils

import (
	"SmartGraphite-server/global"
	"SmartGraphite-server/model"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func FindToken() {
	timeUnix := time.Now().UnixNano() / 1e6
	uid := "00bdd125113744de83b690a7a896b69b"
	sid := "b3bd0acff064472db2944173a8470640"
	random := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000)) //生成0-1000随机字符串
	timestamp := fmt.Sprintf("%v", timeUnix)
	ctx := md5.New() //md5加密
	ctx.Write([]byte(uid + sid + random + timestamp))
	signature := strings.ToUpper(hex.EncodeToString(ctx.Sum(nil))) //签名转换成字符串和大写32位
	//http请求调用初始化token接口
	URL := "http://sukon-cloud.com/api/v1/token/initToken"
	urlValues := url.Values{}
	urlValues.Add("uid", uid)
	urlValues.Add("sid", sid)
	urlValues.Add("random", random)
	urlValues.Add("timestamp", timestamp)
	urlValues.Add("signature", signature)
	res, err := http.PostForm(URL, urlValues)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	var data model.Token
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println(err)
	}
	global.Token = data.Data.Token
	hour := int(math.Floor(float64(data.Data.Expire / 3600)))
	if hour <= 0 {
		time.Sleep(time.Second * 10)
		FindToken()
		fmt.Println("token时效等于0,重新获取token")
		return
	} else {
		if hour == 24 {
			hour = hour - 1
		}
		t := strconv.Itoa(hour)
		global.Spec = "0 0 */" + t + " * * *"
		fmt.Println("获取token成功")
	}
	return
}
