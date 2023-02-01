package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sse/commom"
	"sse/plugins"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"
	"github.com/spf13/viper"
)

// 设置中文字体
func init() {
	fontPaths := findfont.List()
	for _, path := range fontPaths {
		if strings.Contains(path, "simhei.ttf") {
			os.Setenv("FYNE_FONT", path)
			break
		}
	}
}

// 存储Hunter数据的结构体
type HunterJsonResult struct {
	Code int64 `json:"code"`
	Data struct {
		AccountType string `json:"account_type"`
		Arr         []struct {
			AsOrg        string `json:"as_org"`
			Banner       string `json:"banner"`
			BaseProtocol string `json:"base_protocol"`
			City         string `json:"city"`
			Company      string `json:"company"`
			Component    []struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"component"`
			Country        string `json:"country"`
			Domain         string `json:"domain"`
			IP             string `json:"ip"`
			IsRisk         string `json:"is_risk"`
			IsRiskProtocol string `json:"is_risk_protocol"`
			IsWeb          string `json:"is_web"`
			Isp            string `json:"isp"`
			Number         string `json:"number"`
			Os             string `json:"os"`
			Port           int64  `json:"port"`
			Protocol       string `json:"protocol"`
			Province       string `json:"province"`
			StatusCode     int64  `json:"status_code"`
			UpdatedAt      string `json:"updated_at"`
			URL            string `json:"url"`
			WebTitle       string `json:"web_title"`
		} `json:"arr"`
		ConsumeQuota string `json:"consume_quota"`
		RestQuota    string `json:"rest_quota"`
		SyntaxPrompt string `json:"syntax_prompt"`
		Time         int64  `json:"time"`
		Total        int64  `json:"total"`
	} `json:"data"`
	Message string `json:"message"`
}

type SelfInfoFOFA struct {
	Avatar     string `json:"avatar"`
	Email      string `json:"email"`
	Error      bool   `json:"error"`
	Fcoin      int64  `json:"fcoin"`
	FofaServer bool   `json:"fofa_server"`
	FofacliVer string `json:"fofacli_ver"`
	IsVerified bool   `json:"is_verified"`
	Isvip      bool   `json:"isvip"`
	Message    string `json:"message"`
	Username   string `json:"username"`
	VipLevel   int64  `json:"vip_level"`
}

type FOFAJsonResult struct {
	Error   bool       `json:"error"`
	Errmsg  string     `json:"errmsg"`
	Mode    string     `json:"mode"`
	Page    int64      `json:"page"`
	Query   string     `json:"query"`
	Results [][]string `json:"results"`
	Size    int64      `json:"size"`
}

// 设置窗口界面
func myApp() {
	a := app.New()
	w := a.NewWindow("Spatial Search Engine 1.2 by qiwent@idi")
	w.SetIcon(theme.FyneLogo())
	w.Resize(fyne.NewSize(800, 500))
	// 5. 配置API读取、存储
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	configYaml := viper.New()
	configYaml.AddConfigPath(path)     //设置读取的文件路径
	configYaml.SetConfigName("config") //设置读取的文件名
	configYaml.SetConfigType("yaml")   //设置文件的类型
	if err := configYaml.ReadInConfig(); err != nil {
		panic(err)
	}
	hunterApi := widget.NewEntry()
	if str, ok := configYaml.Get("hunter.api").(string); ok {
		hunterApi.Text = str
	} else {
		hunterApi.Text = ""
	}
	hunterKey := widget.NewEntry()
	if str, ok := configYaml.Get("hunter.key").(string); ok {
		hunterKey.Text = str
	} else {
		hunterKey.Text = ""
	}
	fofaApi := widget.NewEntry()
	if str, ok := configYaml.Get("fofa.api").(string); ok {
		fofaApi.Text = str
	} else {
		fofaApi.Text = ""
	}
	fofaEmail := widget.NewEntry()
	if str, ok := configYaml.Get("fofa.email").(string); ok {
		fofaEmail.Text = str
	} else {
		fofaEmail.Text = ""
	}
	fofaKey := widget.NewEntry()
	if str, ok := configYaml.Get("fofa.key").(string); ok {
		fofaKey.Text = str
	} else {
		fofaKey.Text = ""
	}
	config := widget.NewForm(
		widget.NewFormItem("hunter api:", hunterApi),
		widget.NewFormItem("hunter key:", hunterKey),
		widget.NewFormItem("fofa api:", fofaApi),
		widget.NewFormItem("fofa email:", fofaEmail),
		widget.NewFormItem("fofa key:", fofaKey),
	)
	config.OnSubmit = func() {
		f, err := os.OpenFile("./config.yaml", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		} else {
			io.WriteString(f, "hunter:\n")
			io.WriteString(f, " api: "+hunterApi.Text+"\n")
			io.WriteString(f, " key: "+hunterKey.Text+"\n")
			io.WriteString(f, "fofa:\n")
			io.WriteString(f, " api: "+fofaApi.Text+"\n")
			io.WriteString(f, " email: "+fofaEmail.Text+"\n")
			io.WriteString(f, " key: "+fofaKey.Text)
		}
		f.Close()
	}
	config.OnCancel = func() {
		hunterApi.Text = ""
		hunterKey.Text = ""
		fofaApi.Text = ""
		fofaEmail.Text = ""
		fofaKey.Text = ""
		hunterApi.Refresh()
		hunterKey.Refresh()
		fofaApi.Refresh()
		fofaEmail.Refresh()
		fofaKey.Refresh()
	}
	/*==========================================2.hunter查询语法====================================================*/
	ADSearchHunter := widget.NewAccordion()
	adSearchHunter := container.New(layout.NewFormLayout(),
		container.NewVBox(
			widget.NewLabel("连接符"),
			widget.NewLabel("="),
			widget.NewLabel("=="),
			widget.NewLabel("!="),
			widget.NewLabel("!=="),
			widget.NewLabel("&&、||"),
			widget.NewLabel("()"),
		),
		container.NewVBox(
			widget.NewLabel("查询含义"),
			widget.NewLabel("模糊查询，表示查询包含关键词的资产"),
			widget.NewLabel("精确查询，表示查询有且仅有关键词的资产"),
			widget.NewLabel("模糊剔除，表示剔除包含关键词的资产"),
			widget.NewLabel("精确剔除，表示剔除有且仅有关键词的资产"),
			widget.NewLabel("多种条件组合查询，&&同and，表示和；||同or，表示或"),
			widget.NewLabel("括号内表示查询优先级最高"),
		),
	)
	item := widget.NewAccordionItem("搜索技巧", adSearchHunter)
	ADSearchHunter.Append(item)
	dataHunter := [][]string{
		{"语法分类", "语法内容", "语法说明"},
		{"IP", "ip=\"1.1.1.1\"", "搜索IP为 ”1.1.1.1”的资产"},
		{"IP", "ip=\"220.181.111.1/24\"", "搜索起始IP为”220.181.111.1“的C段资产"},
		{"IP", "ip.port=\"80\"", "搜索开放端口为”80“的资产"},
		{"IP", "ip.country=\"CN\" 或 ip.country=\"中国\"", "搜索IP对应主机所在国为”中国“的资产"},
		{"IP", "ip.province=\"江苏\"", "搜索IP对应主机在江苏省的资产"},
		{"IP", "ip.city=\"北京\"", "搜索IP对应主机所在城市为”北京“市的资产"},
		{"IP", "ip.isp=\"电信\"", "搜索运营商为”中国电信”的资产"},
		{"IP", "ip.os=\"Windows\"", "搜索操作系统标记为”Windows“的资产"},
		{"IP", "app=\"Hikvision 海康威视 Firmware 5.0+\" && ip.ports=\"8000\"", "检索使用了Hikvision且ip开放8000端口的资产"},
		{"IP", "ip.port_count>\"2\"", "搜索开放端口大于2的IP（支持等于、大于、小于）"},
		{"IP", "ip.ports=\"80\" && ip.ports=\"443\"", "查询开放了80和443端口号的资产"},
		{"IP", "ip.tag=\"CDN\"", "查询包含IP标签\"CDN\"的资产"},
		{"domain域名", "is_domain=true", "搜索域名标记不为空的资产"},
		{"domain域名", "domain=\"qianxin.com\"", "搜索域名包含\"qianxin.com\"的网站"},
		{"domain域名", "domain.suffix=\"qianxin.com\"", "搜索主域为\"qianxin.com\"的网站"},
		{"header请求头", "header.server==\"Microsoft-IIS/10\"", "搜索server全名为“Microsoft-IIS/10”的服务器"},
		{"header请求头", "header.content_length=\"691\"", "搜索HTTP消息主体的大小为691的网站"},
		{"header请求头", "header.status_code=\"402\"", "搜索HTTP请求返回状态码为”402”的资产"},
		{"header请求头", "header=\"elastic\"", "搜索HTTP请求头中含有”elastic“的资产"},
		{"web网站信息", "is_web=true", "搜索web资产"},
		{"web网站信息", "web.title=\"北京\"", "从网站标题中搜索“北京”"},
		{"web网站信息", "web.body=\"网络空间测绘\"", "搜索网站正文包含”网络空间测绘“的资产"},
		{"web网站信息", "after=\"2021-01-01\" && before=\"2021-12-31\"", "搜索2021年的资产"},
		{"web网站信息", "web.similar=\"baidu.com:443\"", "查询与baidu.com:443网站的特征相似的资产"},
		{"web网站信息", "web.similar_icon==\"17262739310191283300\"", "查询网站icon与该icon相似的资产"},
		{"web网站信息", "web.icon=\"22eeab765346f14faf564a4709f98548\"", "查询网站icon与该icon相同的资产"},
		{"web网站信息", "web.similar_id=\"3322dfb483ea6fd250b29de488969b35\"", "查询与该网页相似的资产"},
		{"web网站信息", "web.tag=\"登录页面\"", "查询包含资产标签\"登录页面\"的资产"},
		{"icp备案信息", "icp.number=\"京ICP备16020626号-8\"", "搜索通过域名关联的ICP备案号为”京ICP备16020626号-8”的网站资产"},
		{"icp备案信息", "icp.web_name=\"奇安信\"", "搜索ICP备案网站名中含有“奇安信”的资产"},
		{"icp备案信息", "icp.name=\"奇安信\"", "搜索ICP备案单位名中含有“奇安信”的资产"},
		{"icp备案信息", "icp.type=\"企业\"", "搜索ICP备案主体为“企业”的资产"},
		{"protocol协议/端口响应", "protocol=\"http\"", "搜索协议为”http“的资产"},
		{"protocol协议/端口响应", "protocol.transport=\"udp\"", "搜索传输层协议为”udp“的资产"},
		{"protocol协议/端口响应", "protocol.banner=\"nginx\"", "查询端口响应中包含\"nginx\"的资产"},
		{"app组件信息", "app.name=\"小米 Router\"", "搜索标记为”小米 Router“的资产"},
		{"app组件信息", "app.type=\"开发与运维\"", "查询包含组件分类为\"开发与运维\"的资产"},
		{"app组件信息", "app.vendor=\"PHP\"", "查询包含组件厂商为\"PHP\"的资产"},
		{"app组件信息", "app.version=\"1.8.1\"", "查询包含组件版本为\"1.8.1\"的资产"},
		{"cert证书", "cert=\"baidu\"", "搜索证书中带有baidu的资产"},
		{"cert证书", "cert.subject=\"qianxin.com\"", "搜索证书使用者是qianxin.com的资产"},
		{"cert证书", "cert.subject_org=\"奇安信科技集团股份有限公司\"", "搜索证书使用者组织是奇安信科技集团股份有限公司的资产"},
		{"cert证书", "cert.issuer=\"Let's Encrypt Authority X3\"", "搜索证书颁发者是Let's Encrypt Authority X3的资产"},
		{"cert证书", "cert.issuer_org=\"Let's Encrypt\"", "搜索证书颁发者组织是Let's Encrypt的资产"},
		{"cert证书", "cert.sha-1=\"be7605a3b72b60fcaa6c58b6896b9e2e7442ec50\"", "搜索证书签名哈希算法sha1为be7605a3b72b60fcaa6c58b6896b9e2e7442ec50的资产"},
		{"cert证书", "cert.sha-256=\"4e529a65512029d77a28cbe694c7dad1e60f98b5cb89bf2aa329233acacc174e\"", "搜索证书签名哈希算法sha256为4e529a65512029d77a28cbe694c7dad1e60f98b5cb89bf2aa329233acacc174e的资产"},
		{"cert证书", "cert.sha-md5=\"aeedfb3c1c26b90d08537523bbb16bf1\"", "搜索证书签名哈希算法shamd5为aeedfb3c1c26b90d08537523bbb16bf1的资产"},
		{"cert证书", "cert.serial_number=\"35351242533515273557482149369\"", "搜索证书序列号是35351242533515273557482149369的资产"},
		{"cert证书", "cert.is_expired=true", "搜索证书已过期的资产"},
		{"cert证书", "cert.is_trust=true", "搜索证书可信的资产"},
		{"AS", "as.number=\"136800\"", "搜索asn为\"136800\"的资产"},
		{"AS", "as.name=\"CLOUDFLARENET\"", "搜索asn名称为\"CLOUDFLARENET\"的资产"},
		{"AS", "as.org=\"PDR\"", "搜索asn注册机构为\"PDR\"的资产"},
		{"tls-jarm", "tls-jarm.hash=\"21d19d00021d21d21c21d19d21d21da1a818a999858855445ec8a8fdd38eb5\"", "搜索tls-jarm哈希为21d19d00021d21d21c21d19d21d21da1a818a999858855445ec8a8fdd38eb5的资产"},
	}
	syntaxTableHunter := widget.NewTable(
		func() (int, int) {
			return len(dataHunter), len(dataHunter[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		}, nil)
	syntaxTableHunter.UpdateCell = func(id widget.TableCellID, o fyne.CanvasObject) {
		label := o.(*widget.Label)
		label.SetText(dataHunter[id.Row][id.Col])
		label.Wrapping = fyne.TextWrapBreak
	}
	for i := 0; i < len(dataHunter); i++ {
		syntaxTableHunter.SetRowHeight(i, 40)
	}
	syntaxTableHunter.SetColumnWidth(0, 100)
	syntaxTableHunter.SetColumnWidth(1, 300)
	syntaxTableHunter.SetColumnWidth(2, 350)
	hunterSyntax := container.NewBorder(ADSearchHunter, nil, nil, nil, syntaxTableHunter)
	/*==========================================4.fofa查询语法====================================================*/
	title := widget.NewLabel("")
	title.Text = "直接输入查询语句,将从标题,html内容,http头信息,url字段中搜索;\n如果查询表达式有多个与或关系,尽量在外面用（）包含起来；\n新增==完全匹配的符号,可以加快搜索速度,比如查找qq.com所有host,可以是domain==qq.com"
	title.Wrapping = fyne.TextWrapBreak
	// 创建表格存储语法
	dataFOFA := [][]string{
		{"例句", "用途说明", "注"},
		{"title=\"beijing\"", "从标题中搜索“北京”", "-"},
		{"header=\"elastic\"", "从http头中搜索“elastic”", "-"},
		{"body=\"网络空间测绘\"", "从html正文中搜索“网络空间测绘”", "-"},
		{"fid=\"sSXXGNUO2FefBTcCLIT/2Q==\"", "查找相同的网站指纹", "搜索网站类型资产"},
		{"domain=\"qq.com\"", "搜索根域名带有qq.com的网站。", "-"},
		{"icp=\"京ICP证030173号\"", "查找备案号为“京ICP证030173号”的网站", "搜索网站类型资产"},
		{"js_name=\"js/jquery.js\"", "查找网站正文中包含js/jquery.js的资产", "搜索网站类型资产"},
		{"js_md5=\"82ac3f14327a8b7ba49baa208d4eaa15\"", "查找js源码与之匹配的资产", "-"},
		{"cname=\"ap21.inst.siteforce.com\"", "查找cname为\"ap21.inst.siteforce.com\"的网站", "-"},
		{"cname_domain=\"siteforce.com\"", "查找cname包含“siteforce.com”的网站", "-"},
		{"cloud_name=\"Aliyundun\"", "通过云服务名称搜索资产", "-"},
		{"icon_hash=\"-247388890\"", "搜索使用此 icon 的资产", "仅限FOFA高级会员使用"},
		{"host=\".gov.cn\"", "从url中搜索”.gov.cn”", "搜索要用host作为名称"},
		{"port=\"6379\"", "查找对应“6379”端口的资产", "-"},
		{"ip=\"1.1.1.1\"", "从ip中搜索包含“1.1.1.1”的网站", "搜索要用ip作为名称"},
		{"ip=\"220.181.111.1/24\"", "查询IP为“220.181.111.1”的C网段资产", "-"},
		{"status_code=\"402\"", "查询服务器状态为“402”的资产", "查询网站类型数据"},
		{"protocol=\"quic\"", "查询quic协议资产", "搜索指定协议类型(在开启端口扫描的情况下有效)"},
		{"country=\"CN\"", "搜索指定国家(编码)的资产。", "-"},
		{"region=\"Xinjiang Uyghur Autonomous Region\"", "搜索指定行政区的资产。", "-"},
		{"city=\"hangzhou\"", "搜索指定城市的资产。", "-"},
		{"cert=\"baidu\"", "搜索证书(https或者imaps等)中带有baidu的资产。", "-"},
		{"cert.subject=\"Oracle Corporation\"", "搜索证书持有者是Oracle Corporation的资产", "-"},
		{"cert.issuer=\"DigiCert\"", "搜索证书颁发者为DigiCert Inc的资产", "-"},
		{"cert.is_valid=true", "验证证书是否有效，true有效，false无效", "仅限FOFA高级会员使用"},
		{"jarm=\"2ad...83e81\"", "搜索JARM指纹", "-"},
		{"banner=\"users\" && protocol=\"ftp\"", "搜索FTP协议中带有users文本的资产。", "-"},
		{"type=\"service\"", "搜索所有协议资产，支持subdomain和service两种", "搜索所有协议资产"},
		{"os=\"centos\"", "搜索CentOS资产。", "-"},
		{"server==\"Microsoft-IIS/10\"", "搜索IIS 10服务器。", "-"},
		{"server==\"Microsoft-IIS/10\"", "搜索IIS 10服务器。", "-"},
		{"app=\"Microsoft-Exchange\"", "搜索Microsoft-Exchange设备", "-"},
		{"after=\"2017\" && before=\"2017-10-01\"", "时间范围段搜索", "-"},
		{"asn=\"19551\"", "搜索指定asn的资产。", "-"},
		{"org=\"LLC Baxet\"", "搜索指定org(组织)的资产。", "-"},
		{"base_protocol=\"udp\"", "搜索指定udp协议的资产。", "-"},
		{"is_fraud=false", "排除仿冒/欺诈数据", "-"},
		{"is_honeypot=false", "排除蜜罐数据", "仅限FOFA高级会员使用"},
		{"is_ipv6=true", "搜索ipv6的资产", "搜索ipv6的资产,只接受true和false。"},
		{"is_domain=true", "搜索域名的资产", "搜索域名的资产,只接受true和false。"},
		{"is_cloud=true", "筛选使用了云服务的资产", "-"},
		{"port_size=\"6\"", "查询开放端口数量等于\"6\"的资产", "仅限FOFA会员使用"},
		{"port_size_gt=\"6\"", "查询开放端口数量大于\"6\"的资产", "仅限FOFA会员使用"},
		{"port_size_lt=\"6\"", "查询开放端口数量少于\"6\"的资产", "仅限FOFA会员使用"},
		{"ip_ports=\"80,161\"", "搜索同时开放80和161端口的ip", "搜索同时开放80和161端口的ip资产(以ip为单位的资产数据)"},
		{"ip_country=\"CN\"", "搜索中国的ip资产(以ip为单位的资产数据)。", "搜索中国的ip资产"},
		{"ip_region=\"Zhejiang\"", "搜索指定行政区的ip资产(以ip为单位的资产数据)。", "搜索指定行政区的资产"},
		{"ip_city=\"Hangzhou\"", "搜索指定城市的ip资产(以ip为单位的资产数据)。", "搜索指定城市的资产"},
		{"ip_after=\"2021-03-18\"", "搜索2021-03-18以后的ip资产(以ip为单位的资产数据)。", "搜索2021-03-18以后的ip资产"},
		{"ip_before=\"2019-09-09\"", "搜索2019-09-09以前的ip资产(以ip为单位的资产数据)。", "搜索2019-09-09以前的ip资产"},
	}
	syntaxTableFOFA := widget.NewTable(
		func() (int, int) {
			return len(dataFOFA), len(dataFOFA[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		}, nil)
	syntaxTableFOFA.UpdateCell = func(id widget.TableCellID, o fyne.CanvasObject) {
		label := o.(*widget.Label)
		label.SetText(dataFOFA[id.Row][id.Col])
		label.Wrapping = fyne.TextWrapBreak
	}
	for i := 0; i < len(dataFOFA); i++ {
		syntaxTableFOFA.SetRowHeight(i, 40)
	}
	syntaxTableFOFA.SetColumnWidth(0, 300)
	syntaxTableFOFA.SetColumnWidth(1, 400)
	syntaxTableFOFA.SetColumnWidth(2, 350)
	ADSearchFOFA := widget.NewAccordion()
	adSearchFOFA := container.New(layout.NewFormLayout(),
		container.NewVBox(
			widget.NewLabel("逻辑连接符"),
			widget.NewLabel("="),
			widget.NewLabel("=="),
			widget.NewLabel("&&"),
			widget.NewLabel("||"),
			widget.NewLabel("!="),
			widget.NewLabel("~="),
			widget.NewLabel("()"),
		),
		container.NewVBox(
			widget.NewLabel("具体含义"),
			widget.NewLabel("匹配，=\"\"时，可查询不存在字段或者值为空的情况"),
			widget.NewLabel("完全匹配，==\"\"时，可查询存在且值为空的情况"),
			widget.NewLabel("与"),
			widget.NewLabel("或者"),
			widget.NewLabel("不匹配，!=\"\"时，可查询值为空的情况"),
			widget.NewLabel("正则语法匹配专用（高级会员独有，不支持body）"),
			widget.NewLabel("确认查询优先级，括号内容优先级最高"),
		),
	)
	itemFOFA := widget.NewAccordionItem("高级搜索", adSearchFOFA)
	ADSearchFOFA.Append(itemFOFA)
	fofaSyntax := container.NewBorder(ADSearchFOFA, title, nil, nil, syntaxTableFOFA)
	/*==========================================1.hunter搜索查询====================================================*/
	search1 := widget.NewEntry()
	search1.PlaceHolder = "Search..."
	search1.ActionItem = widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		search1.Text = ""
		search1.Refresh()
	})
	searchTime := widget.NewSelect([]string{"最近一个月", "最近半年", "最近一年"}, nil)
	searchTime.SetSelected("最近一个月")
	hPageNum := widget.NewSelect([]string{"10条/页", "50条/页", "100条/页"}, nil)
	hPageNum.SetSelected("10条/页")
	assets := widget.NewSelect([]string{"全部资产", "web服务资产"}, nil)
	assets.SetSelected("web服务资产")
	deDuplication := "false"
	dataDeDuplication := widget.NewCheck("数据去重", func(b bool) {
		if b {
			deDuplication = "true"
		} else {
			deDuplication = "false"
		}
	})
	resultHunterData := [][]string{
		{"序号", "URL", "IP", "端口/服务", "域名", "应用/组件", "站点标题", "状态码", "ICP备案企业", "地理位置", "更新时间"},
	}
	resultHunterShow := widget.NewTable(
		func() (int, int) {
			return len(resultHunterData), len(resultHunterData[0])
		},
		func() fyne.CanvasObject {
			return widget.NewEntry()
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			entry := o.(*widget.Entry)
			entry.SetText(resultHunterData[id.Row][id.Col])
		},
	)
	resultHunterShow.SetColumnWidth(0, 50)
	resultHunterShow.SetColumnWidth(1, 250)
	resultHunterShow.SetColumnWidth(2, 150)
	resultHunterShow.SetColumnWidth(3, 100)
	resultHunterShow.SetColumnWidth(4, 200)
	resultHunterShow.SetColumnWidth(5, 200)
	resultHunterShow.SetColumnWidth(6, 200)
	resultHunterShow.SetColumnWidth(7, 60)
	resultHunterShow.SetColumnWidth(8, 200)
	resultHunterShow.SetColumnWidth(9, 150)
	resultHunterShow.SetColumnWidth(10, 150)
	for i := 0; i < 10; i++ {
		resultHunterShow.SetRowHeight(i, 40)
	}
	currentPageHunter := widget.NewEntry()
	currentPageHunter.Text = "1"
	hunterSurplus := widget.NewLabel("")
	hSearchDataSize := widget.NewLabel("")
	// 查询功能实现
	searchButtonHunter := widget.NewButtonWithIcon("查询", theme.SearchIcon(), func() {
		resultHunterData = resultHunterData[:1]
		currentPageHunter.Refresh()
		t := time.Now()                         // 获取当前时间
		beforeMonth := t.AddDate(0, -1, 0)      // 一个月前的日期
		beforeHalfyear := t.AddDate(0, 0, -179) // 半年前的日期
		beforeYear := t.AddDate(-1, 0, 0)       // 一年前的日期
		var selectTime, selectPage, selectAssets string
		switch searchTime.Selected {
		case "最近一个月":
			selectTime = beforeMonth.Format("2006-01-02")
		case "最近半年":
			selectTime = beforeHalfyear.Format("2006-01-02")
		case "最近一年":
			selectTime = beforeYear.Format("2006-01-02")
		}
		switch hPageNum.Selected {
		case "10条/页":
			selectPage = "10"
		case "50条/页":
			selectPage = "50"
		case "100条/页":
			selectPage = "100"
		}
		switch assets.Selected {
		case "全部资产":
			selectAssets = "3"
		case "web服务资产":
			selectAssets = "1"
		}
		addressHunter := hunterApi.Text + "/openApi/search?api-key=" + hunterKey.Text + "&search=" + commom.HunterBaseEncode(search1.Text) + "&page=" +
			currentPageHunter.Text + "&page_size=" + selectPage + "&is_web=" + selectAssets + "&port_filter=" + deDuplication + "&start_time=" + selectTime + "&end_time=" + t.Format("2006-01-02")
		r, err := http.Get(addressHunter)
		if err != nil {
			panic(err)
		}
		b, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		var hunterJR HunterJsonResult
		json.Unmarshal([]byte(string(b)), &hunterJR)
		asse := ""
		p, _ := strconv.Atoi(currentPageHunter.Text)
		size, _ := strconv.Atoi(selectPage)
		if len(hunterJR.Data.Arr) == 0 {
			dialog.ShowInformation("提示", "查询数据结果为空", w)
			hunterSurplus.Text = hunterJR.Data.RestQuota
			hunterSurplus.Refresh()
		} else {
			for i := 0; i < size; i++ {
				for _, v := range hunterJR.Data.Arr[i].Component {
					asse = asse + v.Name + v.Version + " "
				}
				resultHunterData = append(resultHunterData, []string{
					strconv.Itoa(10*(p-1) + i + 1), hunterJR.Data.Arr[i].URL, hunterJR.Data.Arr[i].IP, strconv.FormatInt(hunterJR.Data.Arr[i].Port, 10) + "/" + hunterJR.Data.Arr[i].Protocol,
					hunterJR.Data.Arr[i].Domain, asse, hunterJR.Data.Arr[i].WebTitle, strconv.FormatInt(hunterJR.Data.Arr[i].StatusCode, 10), hunterJR.Data.Arr[i].Company,
					hunterJR.Data.Arr[i].Country + "" + hunterJR.Data.Arr[i].Province + "" + hunterJR.Data.Arr[i].City, hunterJR.Data.Arr[i].UpdatedAt,
				})
				asse = ""
			}
			resultHunterShow.Refresh()
			hunterSurplus.Text = hunterJR.Data.RestQuota
			hunterSurplus.Refresh()
			hSearchDataSize.Text = "共" + strconv.FormatInt(hunterJR.Data.Total, 10) + "条资产，用时" + strconv.FormatInt(hunterJR.Data.Time, 10)
			hSearchDataSize.Refresh()
		}
	})

	headerHunter := container.NewBorder(nil, nil, nil, searchButtonHunter, search1)
	configItemHunter := container.NewHBox(
		searchTime,
		assets,
		dataDeDuplication,
		layout.NewSpacer(), container.NewHBox(
			hunterSurplus,
			widget.NewButtonWithIcon("提示", theme.InfoIcon(), func() {
				dialog.ShowInformation("提示", "API接口查询暂不支持资产标签分类功能", w)
			}),
			widget.NewButtonWithIcon("数据导出", theme.MailForwardIcon(), func() {
				if len(resultHunterData) > 1 {
					fileName := fmt.Sprintf("./result/Hunter/assets_%s.csv", strconv.FormatInt(time.Now().Unix(), 10))
					os.Create(fileName)
					file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, os.ModePerm)
					if err != nil {
						panic(err)
					}
					defer file.Close()
					// 写入UTF-8 BOM，防止中文乱码
					file.WriteString("\xEF\xBB\xBF")
					w := csv.NewWriter(file)
					w.Write(resultHunterData[0])
					// 写文件需要flush，不然缓存满了，后面的就写不进去了，只会写一部分
					for i := 1; i < len(resultHunterData); i++ {
						w.Write(resultHunterData[i])
					}
					w.Flush()
				}
			})),
	)
	adjustPageNumHunter := container.NewBorder(nil, nil, nil, container.NewHBox(
		hSearchDataSize,
		hPageNum,
		widget.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), func() {
			p, _ := strconv.Atoi(currentPageHunter.Text)
			if p > 1 {
				currentPageHunter.Text = strconv.Itoa(p - 1)
				currentPageHunter.Refresh()
			}
		}),
		currentPageHunter,
		widget.NewButtonWithIcon("", theme.MediaSkipNextIcon(), func() {
			p, _ := strconv.Atoi(currentPageHunter.Text)
			if p >= 1 {
				currentPageHunter.Text = strconv.Itoa(p + 1)
				currentPageHunter.Refresh()
			}
		}),
	))
	hunter := container.NewBorder(container.NewVBox(headerHunter, configItemHunter), adjustPageNumHunter, nil, nil, resultHunterShow)
	/*==========================================3.fofa搜索查询====================================================*/
	search2 := widget.NewEntry()
	search2.PlaceHolder = "Search..."
	search2.ActionItem = widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		search2.Text = ""
		search2.Refresh()
	})
	fPageNum := widget.NewSelect([]string{"100条/页", "500条/页", "1000条/页", "10000条/页"}, nil)
	fPageNum.SetSelected("100条/页")
	currentPageFOFA := widget.NewEntry()
	currentPageFOFA.Text = "1"
	resultFOFAData := [][]string{
		{"序号", "HOST", "标题", "IP", "端口", "域名", "协议", "Server指纹", "地理位置", "备案号"},
	}
	resultFOFAShow := widget.NewTable(
		func() (int, int) {
			return len(resultFOFAData), len(resultFOFAData[0])
		},
		func() fyne.CanvasObject {
			return widget.NewEntry()
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			entry := o.(*widget.Entry)
			entry.SetText(resultFOFAData[id.Row][id.Col])
		},
	)
	resultFOFAShow.SetColumnWidth(0, 50)
	resultFOFAShow.SetColumnWidth(1, 250)
	resultFOFAShow.SetColumnWidth(2, 200)
	resultFOFAShow.SetColumnWidth(3, 150)
	resultFOFAShow.SetColumnWidth(4, 60)
	resultFOFAShow.SetColumnWidth(5, 200)
	resultFOFAShow.SetColumnWidth(6, 70)
	resultFOFAShow.SetColumnWidth(7, 100)
	resultFOFAShow.SetColumnWidth(8, 150)
	resultFOFAShow.SetColumnWidth(9, 150)
	resultFOFAShow.SetColumnWidth(10, 100)
	for i := 0; i < len(resultFOFAData); i++ {
		resultFOFAShow.SetRowHeight(i, 40)
	}
	selfLevel := widget.NewLabel("")
	searchDataSize := widget.NewLabel("")
	searchButtonFOFA := widget.NewButtonWithIcon("查询", theme.SearchIcon(), func() {
		resultFOFAData = resultFOFAData[:1]
		// 请求个人用户数据
		selfInfo := fmt.Sprintf("%s/api/v1/info/my?email=%s&key=%s", fofaApi.Text, fofaEmail.Text, fofaKey.Text)
		r, err := http.Get(selfInfo)
		if err != nil {
			panic(err)
		}
		b, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		var selfIF SelfInfoFOFA
		json.Unmarshal([]byte(string(b)), &selfIF)
		vipLevel := selfIF.VipLevel
		switch vipLevel {
		case 0:
			selfLevel.Text = "普通用户"
		case 1:
			selfLevel.Text = "普通会员"
		case 2:
			selfLevel.Text = "高级会员"
		case 3:
			selfLevel.Text = "企业会员"
		}
		selfLevel.Text = "当前用户为:" + selfLevel.Text
		selfLevel.Refresh()
		// 请求查询数据
		selectPage := ""
		switch fPageNum.Selected {
		case "100条/页":
			selectPage = "100"
		case "500条/页":
			selectPage = "500"
		case "1000条/页":
			selectPage = "1000"
		case "10000条/页":
			selectPage = "10000"
		}
		addressFOFA := fofaApi.Text + "/api/v1/search/all?email=" + fofaEmail.Text + "&key=" + fofaKey.Text + "&qbase64=" +
			commom.FofaBaseEncode(search2.Text) + "&page=" + currentPageFOFA.Text + "&size=" + selectPage +
			"&fields=host,title,ip,domain,port,protocol,banner,country_name,region,city,icp"
		r1, err1 := http.Get(addressFOFA)
		if err1 != nil {
			panic(err1)
		}
		defer r1.Body.Close()
		b1, _ := io.ReadAll(r1.Body)
		var fofaJT FOFAJsonResult
		json.Unmarshal([]byte(string(b1)), &fofaJT)
		searchDataSize.Text = fmt.Sprintf("当前查询结果数量:%s条,目前已显示%s条", strconv.FormatInt(fofaJT.Size, 10), selectPage)
		searchDataSize.Refresh()
		p, _ := strconv.ParseInt(selectPage, 10, 64)
		j, _ := strconv.Atoi(currentPageFOFA.Text)
		if fofaJT.Error {
			dialog.ShowInformation("提示", fofaJT.Errmsg, w)
		} else {
			if fofaJT.Size > 0 {
				if fofaJT.Size >= p {
					for i := 0; i < int(p); i++ {
						resultFOFAData = append(resultFOFAData, []string{
							strconv.Itoa(10*(j-1) + i + 1), fofaJT.Results[i][0], fofaJT.Results[i][1], fofaJT.Results[i][2],
							fofaJT.Results[i][4], fofaJT.Results[i][3], fofaJT.Results[i][5], fofaJT.Results[i][6], fofaJT.Results[i][7] +
								" " + fofaJT.Results[i][8] + " " + fofaJT.Results[i][9], fofaJT.Results[i][10],
						})
					}
				} else {
					for i := 0; i < int(fofaJT.Size); i++ {
						resultFOFAData = append(resultFOFAData, []string{
							strconv.Itoa(10*(j-1) + i + 1), fofaJT.Results[i][0], fofaJT.Results[i][1], fofaJT.Results[i][2],
							fofaJT.Results[i][4], fofaJT.Results[i][3], fofaJT.Results[i][5], fofaJT.Results[i][6], fofaJT.Results[i][7] +
								" " + fofaJT.Results[i][8] + " " + fofaJT.Results[i][9], fofaJT.Results[i][10],
						})
					}
				}

			} else {
				dialog.ShowInformation("提示", "未查询到数据结果", w)
			}
		}
	})
	configItemFOFA := container.NewHBox(
		selfLevel,
		layout.NewSpacer(), container.NewHBox(
			widget.NewButtonWithIcon("提示", theme.InfoIcon(), func() {
				dialog.ShowInformation("提示", "本工具不提供支持企业会员特权专项如:\n蜜罐、其他数据排除，FID字段查询等\n\nFOFA暂不支持domain=\"gov.cn\"方式直接查询全部政府域名\n但支持domain=\"hangzhou.gov.cn\"查询子域名", w)
			}),
			widget.NewButtonWithIcon("数据导出", theme.MailForwardIcon(), func() {
				fileName := fmt.Sprintf("./result/FOFA/assets_%s.csv", strconv.FormatInt(time.Now().Unix(), 10))
				os.Create(fileName)
				file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, os.ModePerm)
				if err != nil {
					panic(err)
				}
				defer file.Close()
				// 写入UTF-8 BOM，防止中文乱码
				file.WriteString("\xEF\xBB\xBF")
				w := csv.NewWriter(file)
				w.Write(resultFOFAData[0])
				// 写文件需要flush，不然缓存满了，后面的就写不进去了，只会写一部分
				for i := 1; i < len(resultFOFAData); i++ {
					w.Write(resultFOFAData[i])
				}
				w.Flush()
			})),
	)
	headerFOFA := container.NewBorder(nil, nil, nil, searchButtonFOFA, search2)
	adjustPageNumFOFA := container.NewBorder(nil, nil, nil, container.NewHBox(
		searchDataSize,
		fPageNum,
		widget.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), func() {
			p, _ := strconv.Atoi(currentPageFOFA.Text)
			if p > 1 {
				currentPageFOFA.Text = strconv.Itoa(p - 1)
				currentPageFOFA.Refresh()
			}
		}),
		currentPageFOFA,
		widget.NewButtonWithIcon("", theme.MediaSkipNextIcon(), func() {
			p, _ := strconv.Atoi(currentPageFOFA.Text)
			if p >= 1 {
				currentPageFOFA.Text = strconv.Itoa(p + 1)
				currentPageFOFA.Refresh()
			}
		}),
	))
	fofa := container.NewBorder(container.NewVBox(headerFOFA, configItemFOFA), adjustPageNumFOFA, nil, nil, resultFOFAShow)
	/*==========================================6.子域名暴破====================================================*/
	domain := widget.NewEntry()
	domain.PlaceHolder = "请输入域名"
	domain.Validator = validation.NewRegexp(`[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+\.?`, "")
	filePath := widget.NewLabel("")
	filePath.Alignment = fyne.TextAlignCenter
	filePath.SetText("请选择子域名字典文件TXT")
	resultSubdomainBurst := widget.NewMultiLineEntry()
	subdomainSearchButton := widget.NewButtonWithIcon("解析", theme.SearchIcon(), func() {
		go func() {
			if domain.Text != "" {
				if isOk, _ := regexp.MatchString(`[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+\.?`, domain.Text); isOk {
					resultSubdomainBurst.SetText("域名解析结果为：" + plugins.IPResolution(domain.Text))
					resultSubdomainBurst.Refresh()
				} else {
					resultSubdomainBurst.SetText("域名错误")
					resultSubdomainBurst.Refresh()
				}
			}
		}()
	})
	importFile := widget.NewButtonWithIcon("导入", theme.DocumentIcon(), func() {
		dialog.ShowFileOpen(func(uc fyne.URIReadCloser, err error) {
			if uc != nil {
				filePath.SetText(uc.URI().Path())
			}
		}, w)
	})
	subdomainBurstButton := widget.NewButtonWithIcon("暴破", theme.SearchReplaceIcon(), func() {
		resultSubdomainBurst.Text = ""
		start := time.Now().Unix()
		go func() {
			resultSubdomainBurst.SetText(plugins.SubdomainBurst(filePath.Text, domain.Text))
			resultSubdomainBurst.Text += fmt.Sprintf("\n共用时:%ds", time.Now().Unix()-start)
		}()
		resultSubdomainBurst.Refresh()
	})
	searchSubdomain := container.NewBorder(nil, nil, nil, subdomainSearchButton, domain)
	imports := container.NewBorder(nil, nil, nil, container.NewHBox(importFile, subdomainBurstButton), filePath)
	headerSubdomain := container.NewBorder(searchSubdomain, imports, nil, nil)
	subdomain := container.NewBorder(headerSubdomain, nil, nil, nil, resultSubdomainBurst)
	/*==========================================7.端口扫描====================================================*/
	t1 := []string{"1433", "1521", "3306", "5432", "6379", "9200", "11211", "27017"}
	t2 := []string{"21", "22", "23", "25", "53", "80", "81", "110", "111", "123", "135", "139", "389", "443", "445", "465", "500", "515", "548", "623",
		"636", "873", "902", "1080", "1099", "1433", "1434", "1521", "1883", "2049", "2181", "2375", "2379", "3128", "3306", "3389", "4730", "5222", "5432",
		"5555", "5601", "5672", "5900", "5938", "5984", "6000", "6379", "7001", "7077", "8080", "8081", "8443", "8545", "8686", "9000", "9001", "9042", "9092",
		"9100", "9200", "9418", "9999", "11211", "27017", "37777", "50000", "50070", "61616"}
	t3 := []string{"1-65535"}
	host := widget.NewMultiLineEntry()
	host.PlaceHolder = "IP:192.168.1.1\n网段:192.168.1.1/24\n域名:www.baidu.com\n多个目标用换行或,分割"
	ports := widget.NewEntry()
	ports.PlaceHolder = "端口号"
	ports.ActionItem = widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("提示", "端口扫描功能不会进行主机探活,即np方式进行扫描\n不设置并发或者超时参数时\n并发默认100Goroutine\n超时默认为3000ms\n减少超时参数可以明显提升扫描速度", w)
	})
	thread := widget.NewEntry()
	thread.PlaceHolder = "并发"
	timeout := widget.NewEntry()
	timeout.PlaceHolder = "超时(ms)"
	rg := widget.NewRadioGroup([]string{"数据库", "企业", "全端口", "自定义"}, func(s string) {
		switch s {
		case "数据库":
			ports.Text = strings.Join(t1, ",")
			ports.Refresh()
		case "企业":
			ports.Text = strings.Join(t2, ",")
			ports.Refresh()
		case "全端口":
			ports.Text = strings.Join(t3, ",")
			ports.Refresh()
		case "自定义":
			ports.Text = ""
			ports.Refresh()
		}
	})
	rg.Horizontal = true
	resultPortScan := widget.NewMultiLineEntry()
	scan := widget.NewButtonWithIcon("扫描", theme.SearchIcon(), func() {
		resultPortScan.Text = ""
		start := time.Now().Unix()
		go func() {
			s := plugins.PortScan(ports.Text, host.Text, thread.Text, timeout.Text)
			resultPortScan.SetText(s)
			resultPortScan.Text += fmt.Sprintf("共用时:%ds", time.Now().Unix()-start)
		}()
		resultPortScan.Refresh()
	})
	headerPortScan := container.NewBorder(nil, container.NewBorder(nil, rg, nil, nil, ports), nil,
		container.NewBorder(container.New(layout.NewGridWrapLayout(fyne.NewSize(70, 35)), timeout, thread), scan, nil, nil), host)
	portscan := container.NewBorder(headerPortScan, nil, nil, nil, resultPortScan)
	/*====================================================总体布局=======================================================*/
	tabs := container.NewAppTabs(
		container.NewTabItem("Hunter", hunter),
		container.NewTabItem("H查询语法", hunterSyntax),
		container.NewTabItem("FOFA", fofa),
		container.NewTabItem("F查询语法", fofaSyntax),
		container.NewTabItem("配置", config),
		container.NewTabItem("子域名暴破", subdomain),
		container.NewTabItem("端口扫描", portscan),
	)
	tabs.SetTabLocation(container.TabLocationLeading)
	w.SetContent(tabs)
	w.ShowAndRun()
}

func main() {
	commom.CreateFile()
	myApp()
}
