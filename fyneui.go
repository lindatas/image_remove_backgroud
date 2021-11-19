package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	appHeight  float32
	appWidht  float32
	cHeight float32
	cWidht float32
)


func init_size()  {
	//app windows size
	appHeight  = 650
	appWidht = 900

	//bottom container size
	cHeight = 580
	cWidht = 900
}

type ImageResource struct {
	FileName    string
	FileContent []byte
}

type removebg_ui struct {
	parent_window	fyne.Window
	main_container 	*fyne.Container
	left_img 		*canvas.Image
	right_img		*canvas.Image
	left_button     *widget.Button
	right_button    *widget.Button
	//resource_img    *ImageResource
	result_img		*ImageResource
	color_label     *widget.Label
	freecall_label  *widget.Label
	api_key 		string
}

func (r *removebg_ui)loadUI(window fyne.Window)  {

	var img_width float32 = 250/3*2
	var img_hight float32 = 350/3*2
	r.parent_window = window
	r.left_img = canvas.NewImageFromImage(create_img())
	r.left_img.Resize(fyne.NewSize(img_width,img_hight))
	r.right_img = canvas.NewImageFromImage(create_img())
	r.right_img.Resize(fyne.NewSize(img_width,img_hight))

	r.left_button = r.OpenFil_button("打开图片",window)
	right_button := r.SaveFile_button("保存图片")
	removebg_button := r.Removebg_button("去除背景")

	r.left_img.Move(fyne.NewPos(cWidht/16,cHeight/8))
	r.right_img.Move(fyne.NewPos(cWidht/16*9,cHeight/8))

	lbutton_Y := r.left_img.Position().Y + r.left_img.Size().Height + 25
	rbutton_Y := r.right_img.Position().Y+ r.right_img.Size().Height + 150
	lbutton_X := r.left_img.Position().X + r.left_img.Size().Width/2 -50    //因为container 会撑大
	rbutton_X := r.right_img.Position().X + r.right_img.Size().Width/2 -50

	//button 不能直接放置无布局container中，不然无法点击
	left_button_con := container.NewVBox(r.left_button,removebg_button)
	right_button_con := container.NewVBox(right_button)

	left_button_con.Move(fyne.NewPos(lbutton_X,lbutton_Y))
	right_button_con.Move(fyne.NewPos(rbutton_X,rbutton_Y))

	//color select
	bg_colors := []string{"红色","蓝色","白色","无颜色"}

	var change_color func(string)
	change_color = r.change_color_ui
	radios := widget.NewRadioGroup(bg_colors, func(s string) {
		change_color(s)
	})
	radios.Horizontal = true
	//radios.Resize(fyne.NewSize(100,5))
	current_color_label := widget.NewLabel("当前RGB颜色：")
	current_num_label := widget.NewLabel("")
	current_color_con := container.NewGridWithColumns(2,current_color_label,current_num_label)
	//current_color_con.Resize(fyne.NewSize(100,5))
	r.color_label = current_num_label

	color_input := widget.NewEntry()
	color_input.SetPlaceHolder("#")
	color_input.OnChanged = func(s string) { r.change_color_bynum_ui(color_input.Text) }
	color_input_label := widget.NewLabel("请输入RGB颜色：")
	color_input_con := container.NewGridWithColumns(2,color_input_label,color_input)
	//color_input_con.Resize(fyne.NewSize(100,5))
	color_select_con :=container.NewVBox(radios,current_color_con,color_input_con)
	color_select_con.Move(fyne.NewPos(r.right_img.Position().X-70,rbutton_Y-150))

	r.freecall_label = widget.NewLabel("")
	setting_button := widget.NewButton("", func() {

	})
	setting_button.SetIcon(theme.SettingsIcon())
	setting_button_con := container.NewHBox(r.freecall_label,setting_button)
	setting_button_con.Move(fyne.NewPos(0,cHeight-5))


	r.main_container = container.NewWithoutLayout(r.left_img,left_button_con,r.right_img,right_button_con,color_select_con,setting_button_con)


	r.update_freecall_label()
}


func (r *removebg_ui) OpenFil_button(name string,window fyne.Window) *widget.Button {
	error_flag := false
	var file_content []byte
	var file_name string
	//参数r不能传3层，所以需要再申请变量
	image_container := r.left_img
	var f func(string,[]byte)
	f = r.image_selected

	return widget.NewButton(name, func() {
		//此文件打开操作是异步的，不会造成等待。相当于开启了一个协程
		dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				error_flag = true
				return
			}
			if r == nil {
				return
			}
			file_name = r.URI().Name()
			bytes1, err := ioutil.ReadAll(r)

			if err != nil {
				error_flag = true
				log.Fatalln(err)
			}
			file_content = bytes1

			//img,_,_ = image.Decode(bytes.NewBuffer(file_content))

			_,_,file_type,err := get_imagefile_config(file_name,file_content)
			if err != nil{
				dialog.ShowError(err,window)
				return
			}

			if file_type != "png" && file_type != "jpg" && file_type != "jpeg"{
				dialog.ShowError(errors.New("图片内容格式错误"),window)
				return
			}

			image_container.Resource = fyne.NewStaticResource(file_name,file_content)
			image_container.Refresh()
			defer f(file_name,file_content)
		}, window)

	})
}

func (r *removebg_ui) SaveFile_button (name string) *widget.Button {
	return widget.NewButton(name, func() {
		if r.right_img.Resource == nil || len(r.right_img.Resource.Content()) <100 {
			return
		}
		dialog.ShowFileSave(func(closer fyne.URIWriteCloser, err error) {
			_,err = closer.Write(r.right_img.Resource.Content())
			if err != nil {
				dialog.ShowError(err,r.parent_window)
			}
			_ = closer.Close()
		},r.parent_window)
	})
}

func (r *removebg_ui) Removebg_button (name string) *widget.Button {
	return widget.NewButton(name, func() {
		if r.right_img.Resource == nil || len(r.right_img.Resource.Content()) <100 {
			return
		}
		result,err := RemoveBG_from_base64_image(r.api_key,r.left_img.Resource.Content())
		if err != nil {
			fmt.Println(err)
			dialog.ShowError(err,r.parent_window)
			return
		}
		r.result_img = &ImageResource{
			r.left_img.Resource.Name(),
			result,
		}
		r.ShowResultFile(result)
		r.update_freecall_label()
	})
}

//完成图片选择后，先不去背景，直接显示在右边image里
func (r *removebg_ui) image_selected(file_name string,file_content []byte)  {
	r.result_img = &ImageResource{
		file_name,
		file_content,
	}
	r.ShowResultFile(file_content)
}

func (r *removebg_ui) ShowResultFile(file_content []byte) {
	r.right_img.Resource = fyne.NewStaticResource("test",file_content)
	r.right_img.Refresh()
}

func (r *removebg_ui) change_color_ui(s string)  {
	if r.result_img == nil || len(r.result_img.FileContent) == 0 {
		dialog.ShowError(errors.New("当前无法选择颜色"),r.parent_window)
		return
	}

	var bg_color color.RGBA
	switch s {
	case "红色":
		bg_color = color.RGBA{250,0,0,0}
		r.color_label.SetText("#FA0000")
	case "蓝色":
		bg_color = color.RGBA{0,238,238,0}
		r.color_label.SetText("#00EEEE")
	case "白色":
		bg_color = color.RGBA{255,255,255,0}
		r.color_label.SetText("#FFFFFF")
	default:
		r.ShowResultFile(r.result_img.FileContent)
		r.color_label.SetText("#FFFFFF")
		return
	}


	changed_file_content,err := img_change_color(r.result_img.FileContent,bg_color)
	if err != nil {
		dialog.ShowError(err,r.parent_window)
		return
	}
	r.ShowResultFile(changed_file_content)
}

func (r *removebg_ui) change_color_bynum_ui(s string) {
	if len(s) != 7 || s[0:1] != "#"{
		return
	}
	color_str := "0123456789abcdef"
	s = strings.ToLower(s)
	for i:=1; i<7 ;i++{
		if !strings.Contains(color_str,s[i:i+1]){
			return
		}
	}
	r_str,g_str,b_str := s[1:3],s[3:5],s[5:]
	r_u64,_ := strconv.ParseUint(r_str,16,64)
	g_u64,_ := strconv.ParseUint(g_str,16,64)
	b_u64,_ := strconv.ParseUint(b_str,16,64)
	r_uint8,g_uint8,b_uint8 := uint8(r_u64),uint8(g_u64),uint8(b_u64)

	bg_color := color.RGBA{r_uint8,g_uint8,b_uint8,0}
	changed_file_content,err := img_change_color(r.result_img.FileContent,bg_color)
	if err != nil {
		dialog.ShowError(err,r.parent_window)
		return
	}
	r.ShowResultFile(changed_file_content)
	r.color_label.SetText(s)

}

func (r *removebg_ui) update_freecall_label() {
	if r.freecall_label == nil || r.freecall_label.Text == ""{
		api_key,err := load_api_key()
		if err != nil {
			dialog.ShowError(err,r.parent_window)
			return
		}
		r.api_key = api_key
	}
	count,err := RemoveBG_Freecall(r.api_key)
	if err !=nil {
		r.freecall_label.SetText(err.Error())
	}else{
		r.freecall_label.SetText(fmt.Sprintf("removebg剩余免费次数：%d",count))
	}
}

func initFont()  {
	fontPaths := findfont.List()
	for _,path := range fontPaths {
		if strings.Contains(path,"msyh.ttc"){
			os.Setenv("FYNE_FONT",path)
			break
		}
	}
}

func clearFont()  {
	os.Unsetenv("FYNE_FONT")
}


func create_img() image.Image{
	width := 250
	height := 350

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Colors are defined by Red, Green, Blue, Alpha uint8 values.
	cyan := color.RGBA{190, 190, 190, 0xff}

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			switch {
			case y==x*7/5:
				img.Set(x,y,color.Black)
			case 350-y==x*7/5:
				img.Set(x,y,color.Black)
			default:
				img.Set(x, y, cyan)
			}
		}
	}
	return img
}


func OCR() *fyne.Container{
	left_image := canvas.NewImageFromFile("")
	left_image.SetMinSize(fyne.NewSize(300,200))
	left_button := widget.NewButton("文字识别", func() {

	})
	left_container := container.NewVBox(left_image,left_button)

	right_container := container.NewVBox(left_image,left_button)
	line1 := canvas.NewLine(color.Black)

	return container.NewHBox(left_container,line1,right_container)
}


func ui_layout(){
	initFont()
	init_size()

	myapp := app.New()
	w := myapp.NewWindow("test")
	ui1 := removebg_ui{}

	//myapp.Settings().SetTheme(&FontTheme{})
	w.Resize(fyne.NewSize(appWidht,appHeight))
	main_container := container.NewVBox()
	top_container := container.New(layout.NewHBoxLayout())
	bottom_container := container.New(layout.NewHBoxLayout())
	ui1.loadUI(w)
	remove_container := ui1.main_container
	ORC_container := OCR()


	main_button1 := widget.NewButton("文件操作", func() {
		lens := len(main_container.Objects)
		if lens >1 {
			main_container.Remove(main_container.Objects[lens-1])
		}
		main_container.Add(bottom_container)
	})
	main_button2 := widget.NewButton("图片换背景", func() {
		lens := len(main_container.Objects)
		if lens >1 {
			main_container.Remove(main_container.Objects[lens-1])
		}
		main_container.Add(remove_container)
	})
	main_button3 := widget.NewButton("OCR文字识别", func() {
		lens := len(main_container.Objects)
		if lens >1 {
			main_container.Remove(main_container.Objects[lens-1])
		}
		main_container.Add(ORC_container)
	})

	line1 := canvas.NewLine(color.Black)
	line1.StrokeWidth = 1.1

	top_container.Add(main_button1)
	top_container.Add(main_button2)
	top_container.Add(main_button3)


	bottom_container.Resize(fyne.NewSize(cWidht,cHeight))
	//main_container := container.NewVBox(top_container,line1,bottom_container)
	main_container.Add(top_container)
	main_container.Add(line1)


	ract1 := canvas.NewRectangle(color.Black)
	ract1.SetMinSize(fyne.NewSize(cWidht,cHeight))
	bottom_container.Add(ract1)

	w.SetContent(main_container)

	w.ShowAndRun()
	clearFont()
}

func guistart()  {
	ui_layout()
}
