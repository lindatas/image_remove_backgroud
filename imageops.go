package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"golang.org/x/image/draw"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

//removebg 实现
func RemoveBG_api(api_key string,file_content []byte) ([]byte,error) {

	imagefile,err := os.Open("static/3.jpg")
	if err != nil {
		panic(err)
	}
	defer imagefile.Close()
	read := bufio.NewReader(imagefile)
	all,err := ioutil.ReadAll(read)
	if err != nil {
		panic(err)
	}

	return RemoveBG_from_base64_image(api_key,all)
}

//根据api key 获取当前free_call的数量
func RemoveBG_Freecall(api_key string) (result int,err error)  {
	//返回json结构
	type account struct {
		Data struct {
			Attributes struct {
				Credits struct {
					Subscription int `json:"subscription"`
					Payg         int `json:"payg"`
					Enterprise   int `json:"enterprise"`
					Total        int `json:"total"`
				} `json:"credits"`
				Api struct {
					FreeCalls int    `json:"free_calls"`
					Sizes     string `json:"sizes"`
				} `json:"api"`
			} `json:"attributes"`
		} `json:"data"`
	}
	var account_info account

	api_url := "https://api.remove.bg/v1.0/account"

	client := &http.Client{}
	request,err := http.NewRequest("GET",api_url,nil)
	if err != nil{
		return 0,err
	}
	request.Header.Set("Accept","*/*")
	request.Header.Set("X-API-Key",api_key)

	rsp,err := client.Do(request)
	if err != nil{
		return 0,err
	}
	defer rsp.Body.Close()
	body, err2 := ioutil.ReadAll(rsp.Body)
	if err2 != nil {
		return 0,err
	}

	err3 := json.Unmarshal(body,&account_info)
	if err != nil {
		return 0,err3
	}
	return account_info.Data.Attributes.Api.FreeCalls,nil
}

func RemoveBG_from_base64_image(api_key string,image_content []byte) ([]byte,error) {
	//api_url := "https://api.remove.bg/v1.0/removebg"
	api_url := "http://127.0.0.1:8889"


	encode_content := "image_file_b64=" + base64.StdEncoding.EncodeToString(image_content) + "&size=regular"
	body := bytes.NewReader([]byte(encode_content))
	req,err := http.NewRequest("POST",api_url,body)
	if err != nil {
		return nil,err
	}
	req.Header.Set("Accept","*/*")
	req.Header.Set("Content-Type","application/x-www-form-urlencoded")
	req.Header.Set("X-API-Key",api_key)
	req.Header.Set("X-Connection-Key","keep-alive")

	//tr := &http.Transport{
	//	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//}
	//proxy_url,_ := url.Parse("http://127.0.0.1:8080")
	//tr.Proxy = http.ProxyURL(proxy_url)
	client := &http.Client{
		//Transport: tr,
	}

	rsp,err := client.Do(req)
	if err != nil {
		return nil,err
	}

	resopse_body,err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil,err
	}
	if rsp.StatusCode != 200 {
		return nil,errors.New(string(resopse_body))
	}

	return resopse_body,nil

}

func get_imagefile_config(file_name string,file_content []byte) (int,int,string,error){
	//allow_img_type := []string{"png","jpg","jpeg"}
	file_ext := path.Ext(file_name)

	if file_ext != ".png" && file_ext != ".jpg" && file_ext != ".jpeg"{
		return 0,0,"",errors.New("请选择正确图片类型")
	}
	c,f_type,err := image.DecodeConfig(bytes.NewReader(file_content))
	if err != nil {
		return 0,0,"",err
	}
	return c.Width,c.Height,f_type,nil
}


func img_change_color(file_content []byte,bg_color color.RGBA) ([]byte,error){

	img,_,err := image.Decode(bytes.NewReader(file_content))
	if err != nil {
		return nil,err
	}

	bound := img.Bounds()
	dx := bound.Dx()
	dy := bound.Dy()
	newIm := image.NewRGBA(image.Rect(0,0,dx,dy))
	newIm2 := image.NewRGBA(image.Rect(0,0,dx,dy))
	for i := 0; i < dx; i++ {
		for j := 0; j < dy; j++ {
			newIm2.Set(i,j,bg_color)
		}
	}

	//使用循环遍历(0,0,0)的像素点，并进行替换，人物会存在黑边。因为存在(r,g,b) r,g,b均为个位数的像素点。所以这里使用贴图的方式
	draw.Draw(newIm,newIm.Bounds().Add(image.Pt(0,0)),newIm2,newIm2.Bounds().Min,draw.Src)

	//最重要的是后面这个参数draw.Over，如果使用Src得到的就是黑色背景的图。因为(0,0,0)为空白颜色，转换后显示为黑色
	draw.DrawMask(newIm,newIm.Bounds(),img,img.Bounds().Min,img,image.Point{0,0},draw.Over)

	buff := new(bytes.Buffer)
	//if file_type == "jpg" || file_type == "jpeg" {
	//	jpeg.Encode(buff,newIm,&jpeg.Options{Quality: 100})
	//
	//}else if file_type == "png" {
	//	//png.Encode(buff,newIm)
	//
	//}else{
	//	return []byte{},errors.New("变更底色时，遇到文件类型错误")
	//}
	jpeg.Encode(buff,newIm,&jpeg.Options{Quality: 100})

	return buff.Bytes(),nil
}

func load_api_key() (api_key string,err error)  {
	file_name := "config"
	file,err := os.Open(file_name)
	defer file.Close()
	if err != nil {
		return "",err
	}
	s,err := ioutil.ReadAll(file)
	if err != nil {
		return "",err
	}
	if len(s) < 10 {
		return "",errors.New("api key 长度不正确")
	}
	return string(s),nil
}
