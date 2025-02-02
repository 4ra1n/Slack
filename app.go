package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"slack-wails/core"
	alibaba "slack-wails/core/druid"
	"slack-wails/core/info"
	"slack-wails/core/jsfind"
	"slack-wails/core/portscan"
	"slack-wails/core/space"
	"slack-wails/core/webscan"
	"slack-wails/lib/clients"
	"slack-wails/lib/gologger"
	"slack-wails/lib/util"
	"strconv"
	"strings"
	"sync"
	"time"
)

// App struct
type App struct {
	ctx              context.Context
	workflowFile     string
	webfingerFile    string
	activefingerFile string
	cdnFile          string
	qqwryFile        string
	avFile           string
	bruteFile        string
	defaultPath      string
}

// NewApp creates a new App application struct
func NewApp() *App {
	home := util.HomeDir()
	return &App{
		workflowFile:     home + "/slack/config/workflow.yaml",
		webfingerFile:    home + "/slack/config/webfinger.yaml",
		activefingerFile: home + "/slack/config/dir.yaml",
		cdnFile:          home + "/slack/config/cdn.yaml",
		qqwryFile:        home + "/slack/config/qqwry.dat",
		avFile:           home + "/slack/config/antivirues.yaml",
		bruteFile:        home + "/slack/portburte",
		defaultPath:      home + "/slack/",
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Callgologger(level, msg string) {
	switch level {
	case "info":
		gologger.Info(a.ctx, msg)
	case "warning":
		gologger.Warning(a.ctx, msg)
	case "error":
		gologger.Error(a.ctx, msg)
	default:
		gologger.Debug(a.ctx, msg)
	}
}

func (a *App) IsRoot() bool {
	return os.Getuid() == 0
}

type Response struct {
	Error  bool
	Proto  string
	Header []map[string]string
	Body   string
}

func (a *App) GoFetch(method, target, body string, headers []map[string]string, timeout int, proxy clients.Proxy) *Response {
	if _, err := url.Parse(target); err != nil {
		return &Response{
			Error:  true,
			Proto:  "",
			Header: nil,
			Body:   "",
		}
	}
	hhhhheaders := http.Header{}
	for _, head := range headers {
		for k, v := range head {
			hhhhheaders.Set(k, v)
		}
	}
	resp, b, err := clients.NewRequest(method, target, hhhhheaders, bytes.NewReader([]byte(body)), 10, true, clients.JudgeClient(proxy))
	if err != nil {
		return &Response{
			Error:  true,
			Proto:  "",
			Header: nil,
			Body:   "",
		}
	}
	var headerArray []map[string]string
	for key, values := range resp.Header {
		// 对于每个键，创建一个新的 map 并添加键值对
		headerMap := make(map[string]string)
		headerMap["key"] = key
		headerMap["value"] = strings.Join(values, " ")

		// 将 map 添加到切片中
		headerArray = append(headerArray, headerMap)
	}
	return &Response{
		Error:  false,
		Proto:  resp.Proto,
		Header: headerArray,
		Body:   string(b),
	}
}

// fscan
func (a *App) Fscan2Txt(content string) string {
	var result string
	result += "[NetInfo]\n"
	for _, netinfo := range core.NetInfoReg.FindAllString(content, -1) {
		result += netinfo + "\n"
	}
	result += "\n"
	lines := strings.Split(content, "\n")
	for name, reg := range core.FscanRegs {
		result += core.MatchLine(name, reg, lines)
	}
	return result
}

// thinkdict
func (a *App) ThinkDict(userNameCN, userNameEN, companyName, companyDomain, birthday, jobNumber, connectWord string, weakList []string) []string {
	return core.GenerateDict(userNameCN, userNameEN, companyName, companyDomain, birthday, jobNumber, connectWord, weakList)
}

func (a *App) System(content string, mode int) [][]string {
	if mode == 0 { // tasklist
		datas, _ := core.AntivirusIdentify(content, a.avFile)
		return datas
	} else { // systeminfo
		return core.Patch(content)
	}
}

var onec sync.Once

func (a *App) CyberChefLocalServer() {
	onec.Do(func() {
		go func() {
			// 定义要服务的目录
			dir := util.HomeDir() + "/slack/CyberChef/"
			// 创建文件服务器
			fs := http.FileServer(http.Dir(dir))

			// 设置路由，将所有请求重定向到文件服务器
			http.Handle("/", fs)
			// 指定一个随机端口, 启动HTTP服务器
			err := http.ListenAndServe(fmt.Sprintf(":%d", 8731), nil)
			if err != nil {
				return
			}
		}()
	})
}

func (a *App) ExtractIP(text string) string {
	return core.Analysis(text)
}

func (a *App) DomainInfo(text string) string {
	s, s2, s3 := core.SeoChinaz(text)
	s4, s5, s6, s7, s8, s9 := core.WhoisChinaz(text)
	return fmt.Sprintf(`---备案查询---
公司名称: %v
	
备案号: %v
	
IP: %v
	
---whois查询---
更新时间: %v

创建时间: %v

过期时间: %v

注册商服务器: %v

DNS: %v

状态: %v`, s, s2, s3, s4, s5, s6, s7, s8, s9)
}

func (a *App) CheckCdn(domain string) string {
	var result string
	ips, cnames, err := core.Resolution(domain, 10)
	if err == nil {
		if len(cnames) != 0 {
			for name, cdns := range core.ReadCDNFile(a.cdnFile) {
				for _, cdn := range cdns {
					for _, cname := range cnames {
						if strings.Contains(cname, cdn) { // 识别到cdn
							return fmt.Sprintf("域名: %v 识别到CDN域名，CNAME: %v CDN名称: %v 解析到IP为: %v\n", domain, cname, name, strings.Join(ips, " | "))
						} else if strings.Contains(cname, "cdn") {
							return fmt.Sprintf("域名: %v CNAME中含有关键字: cdn，该域名可能使用了CDN技术 CNAME: %v 解析到IP为: %v \n", domain, cname, strings.Join(ips, " | "))
						}
					}
				}
			}
		}
		result = fmt.Sprintf("域名: %v 解析到IP为: %v CNAME: %v\n", domain, strings.Join(ips, ","), strings.Join(cnames, ","))
	} else {
		result = fmt.Sprintf("域名: %v 解析失败,%v\n", domain, err)
	}
	return result
}

// 初始化IP记录器
func (a *App) InitIPResolved() {
	core.IPResolved = make(map[string]int)
}

// subodomain
func (a *App) LoadSubDict(configPath string) []string {
	return util.LoadSubdomainDict(util.HomeDir()+configPath, "/dicc.txt")
}

func (a *App) Subdomain(subdomain string, timeout int) []string {
	sr := core.BurstSubdomain(subdomain, timeout, a.qqwryFile, a.cdnFile)
	return []string{sr.Subdomain, strings.Join(sr.Cname, " | "), strings.Join(sr.Ips, " | "), sr.Notes}
}

func (a *App) InitTycHeader(token string) {
	info.InitHEAD(token)
}

type TycAssetResult struct {
	Asset  []info.CompanyInfo
	Prompt string
}

// mode 0 = 查询子公司 . mode 1 = 查询公众号
func (*App) SubsidiariesAndDomains(companyName string, ratio int) TycAssetResult {
	var isFuzz string
	companyId, fuzzName := info.GetCompanyID(companyName) // 获得到一个模糊匹配后，关联度最高的名称
	if companyName != fuzzName {                          // 如果传进来的名称与模糊匹配的不相同
		isFuzz = fmt.Sprintf("天眼查模糊匹配名称为%v ——> %v,已替换原有名称进行查.", companyName, fuzzName)
	}
	asset, logs := info.SearchSubsidiary(fuzzName, companyId, ratio)
	if isFuzz == "" {
		logs = isFuzz + logs
	}
	return TycAssetResult{
		Asset:  asset,
		Prompt: logs,
	}

}

type WechatAssetResult struct {
	Asset  [][]string
	Prompt string
}

func (a *App) WechatOfficial(companyName string) WechatAssetResult {
	var isFuzz string
	companyId, fuzzName := info.GetCompanyID(companyName) // 获得到一个模糊匹配后，关联度最高的名称
	if companyName != fuzzName {                          // 如果传进来的名称与模糊匹配的不相同
		isFuzz = fmt.Sprintf("天眼查模糊匹配名称为%v ——> %v,已替换原有名称进行查.", companyName, fuzzName)
	}
	asset, logs := info.WeChatOfficialAccounts(fuzzName, companyId)
	if isFuzz == "" {
		logs = isFuzz + logs
	}
	return WechatAssetResult{
		Asset:  asset,
		Prompt: logs,
	}
}

type HunterSearch struct {
	Total string
	Info  string
}

// mode = 0 campanyName, mode = 1 domain or ip
func (a *App) AssetHunter(mode int, target, api string) HunterSearch {
	if mode == 0 {
		str := fmt.Sprintf("icp.name=\"%v\"", target)
		t, i := space.SearchTotal(api, str)
		return HunterSearch{
			Total: fmt.Sprint(t),
			Info:  i,
		}
	} else {
		var str string
		// 处理网站域名是IP的情况
		if util.RegIP.MatchString(target) {
			str = fmt.Sprintf("ip=\"%v\"", target)
		} else {
			str = fmt.Sprintf("domain.suffix=\"%v\"", target)
		}
		t, i := space.SearchTotal(api, str)
		return HunterSearch{
			Total: fmt.Sprint(t),
			Info:  i,
		}
	}
}

// dirsearch

func (a *App) LoadDirsearchDict(configPath string, newExts []string) []string {
	return util.LoadDirsearchDict(util.HomeDir()+configPath, "/dicc.txt", "%EXT%", newExts)
}

type PathData struct {
	Status   int    // 状态码
	Location string // server信息
	Length   int    // 主体内容
}

func (a *App) PathRequest(method, url string, timeout int, bodyExclude string, redirect bool, customHeader string) PathData {
	var pd PathData // 将响应头和响应头的数据存储到结构体中
	client := clients.NotFollowClient()
	if redirect {
		client = clients.DefaultClient()
	}
	var header = http.Header{}
	if customHeader != "" {
		for _, single := range strings.Split(customHeader, "\n") {
			temp := strings.Split(single, ":")
			header.Set(temp[0], temp[1])
		}
	}
	resp, body, err := clients.NewRequest(method, url, header, nil, timeout, true, client)
	if err != nil {
		return pd
	}
	if bodyExclude != "" && bytes.Contains(body, []byte(bodyExclude)) {
		pd.Status = 1 // status 1 表示被body排除在外，不计入ERROR请求中
	} else {
		pd.Status = resp.StatusCode
	}
	pd.Length = len(body)
	pd.Location = resp.Header.Get("Location")
	return pd
}

// portscan

func (a *App) IPParse(ipList []string) []string {
	return util.ParseIPs(ipList)
}

func (a *App) PortParse(text string) []int {
	return util.ParsePort(text)
}

func (a *App) HostAlive(targets []string, Ping bool) []string {
	return portscan.CheckLive(a.ctx, targets, Ping)
}

func (a *App) NewTcpScanner(ips []string, ports []int, thread, timeout int) {
	portscan.ExitFunc = false
	portscan.TcpScan(a.ctx, ips, ports, thread, timeout)
}

func (a *App) NewCorrespondsScan(specialTargets []string, timeout int) {
	var addrs []portscan.Address
	for _, target := range specialTargets {
		temp := strings.Split(target, ":")
		port, err := strconv.Atoi(temp[1]) // 如果后缀端口有误则继续
		if err != nil {
			continue
		}
		addrs = append(addrs, portscan.Address{
			IP:   temp[0],
			Port: port,
		})
	}
	portscan.ExitFunc = false
	portscan.CorrespondsScan(a.ctx, addrs, timeout)
}

func (a *App) NewSynScanner(ips []string, ports []int) {
	portscan.ExitFunc = false
	portscan.SynScan(a.ctx, ips, util.IntArrayToUint16Array(ports))
}

func (a *App) StopPortScan() {
	portscan.ExitFunc = true
}

// 端口暴破
func (a *App) PortBrute(host string, usernames, passwords []string) {
	portscan.PortBrute(a.ctx, host, usernames, passwords)
}

// fofa

func (a *App) FofaTips(query string) *space.TipsResult {
	var ff space.FofaConfig
	var ts space.TipsResult
	ff.AppId = "9e9fb94330d97833acfbc041ee1a76793f1bc691"
	ff.PrivateKey = `MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQC/TGN5+4FMXo7H3jRmostQUUEO1NwH10B8ONaDJnYDnkr5V0ZzUvkuola7JGSFgYVOUjgrmFGITG+Ne7AgR53Weiunlwp15MsnCa8/IWBoSHs7DX1O72xNHmEfFOGNPyJ4CsHaQ0B2nxeijs7wqKGYGa1snW6ZG/ZfEb6abYHI9kWVN1ZEVTfygI+QYqWuX9HM4kpFgy/XSzUxYE9jqhiRGI5f8SwBRVp7rMpGo1HZDgfMlXyA5gw++qRq7yHA3yLqvTPSOQMYJElJb12NaTcHKLdHahJ1nQihL73UwW0q9Zh2c0fZRuGWe7U/7Bt64gV2na7tlA62A9fSa1Dbrd7lAgMBAAECggEAPrsbB95MyTFc2vfn8RxDVcQ/dFCjEsMod1PgLEPJgWhAJ8HR7XFxGzTLAjVt7UXK5CMcHlelrO97yUadPAigHrwTYrKqEH0FjXikiiw0xB24o2XKCL+EoUlsCdg8GqhwcjL83Mke84c6Jel0vQBfdVQ+RZbetMCxqv1TpqpwW+iswlDY0+OKNxcDSnUyVkBko4M7bCqJ19DjzuHHLRmSuJhWLjX2PzdrVwIrRChxeJRR5AzrNE2BC/ssKasWjZfgkTOW6MS96q+wMLgwFGCQraU0f4AW5HA4Svg8iWT2uukcDg7VXXc/eEmkfmDGzmgsszUJZYb1hYsvjgbMP1ObwQKBgQDw1K0xfICYctiZ3aHS7mOk0Zt6B/3rP2z9GcJVs0eYiqH+lteLNy+Yx4tHtrQEuz16IKmM1/2Ghv8kIlOazpKaonk3JEwm1mCEXpgm4JI7UxPGQj/pFTCavKBBOIXxHJVSUSg0nKFkJVaoJiNy0CKwQNoFGdROk2fSYu8ReB/WlQKBgQDLWQR3RioaH/Phz8PT1ytAytH+W9M4P4tEx/2Uf5KRJxPQbN00hPnK6xxHAqycTpKkLkbJIkVWEKcIGxCqr6iGyte3xr30bt49MxIAYrdC0LtBLeWIOa88GTqYmIusqJEBmiy+A+DudM/xW4XRkgrOR1ZsagzI3FUVlei9DwFjEQKBgG8JH3EZfhDLoqIOVXXzA24SViTFWoUEETQAlGD+75udD2NaGLbPEtrV5ZmC2yzzRzzvojyVuQY1Z505VmKhq2YwUsLhsVqWrJlbI7uI/uLrQsq98Ml+Q5KUNS7c6KRqEU6KrIbVUHPj4zhTnTRqUhQBUoPXjNNNkyilBKSBReyhAoGAd3xGCIPdB17RIlW/3sFnM/o5bDmuojWMcw0ErvZLPCl3Fhhx3oNod9iw0/T5UhtFRV2/0D3n+gts6nFk2LbA0vtryBvq0C85PUK+CCX5QzR9Y25Bmksy8aBtcu7n27ttAUEDm1+SEuvmqA68Ugl7efwnBytFed0lzbo5eKXRjdECgYAk6pg3YIPi86zoId2dC/KfsgJzjWKVr8fj1+OyInvRFQPVoPydi6iw6ePBsbr55Z6TItnVFUTDd5EX5ow4QU1orrEqNcYyG5aPcD3FXD0Vq6/xrYoFTjZWZx23gdHJoE8JBCwigSt0KFmPyDsN3FaF66Iqg3iBt8rhbUA8Jy6FQA==`
	b, err := ff.GetTips(query)
	if err != nil {
		return &ts
	}
	json.Unmarshal(b, &ts)
	return &ts
}

func (a *App) FofaSearch(query, pageSzie, pageNum, address, email, key string, fraud, cert bool) *space.FofaSearchResult {
	return space.FofaApiSearch(query, pageSzie, pageNum, address, email, key, fraud, cert)
}

func (a *App) Sock5Connect(ip string, port, timeout int, username, password string) bool {
	client, err := clients.SelectProxy(&clients.Proxy{
		Enabled:  true,
		Mode:     "SOCK5",
		Address:  ip,
		Port:     port,
		Username: username,
		Password: password,
	}, clients.DefaultClient())
	if err != nil {
		return false
	}
	_, _, err = clients.NewRequest("GET", "http://www.baidu.com/", nil, nil, timeout, true, client)
	return err == nil
}

func (a *App) IconHash(target string) string {
	_, b, err := clients.NewSimpleGetRequest(target, clients.DefaultClient())
	if err != nil {
		return ""
	}
	return util.Mmh3Hash32(util.Base64Encode(b))
}

// infoscan
type AliveTarget struct {
	Status      bool
	ProtocolURL string
}

func (a *App) CheckTarget(host string, proxy clients.Proxy) *AliveTarget {
	protocolURL, err := clients.IsWeb(host, clients.JudgeClient(proxy))
	if err != nil {
		return &AliveTarget{
			Status:      false,
			ProtocolURL: host,
		}
	}
	return &AliveTarget{
		Status:      true,
		ProtocolURL: protocolURL,
	}
}

type InfoResult struct {
	URL          string
	StatusCode   int
	Length       int
	Title        string
	Fingerprints []string
	IsWAF        bool
	WAF          string
}

func (a *App) InitBurteDict() bool {
	if _, err := os.Stat(a.bruteFile); err != nil {
		os.MkdirAll(a.bruteFile+"/username", 0777)
		os.Mkdir(a.bruteFile+"/password", 0777)
		return true
	}
	return false
}

// 仅在执行时调用一次
func (a *App) InitRule() bool {
	return webscan.InitAll(a.webfingerFile, a.activefingerFile, a.workflowFile)
}

// webscan

func (a *App) FingerLength() int {
	return len(webscan.FingerprintDB)
}

func (a *App) FingerScan(target []string, proxy clients.Proxy) {
	webscan.NewFingerScan(a.ctx, target, proxy)
}

func (a *App) ActiveFingerScan(target []string, proxy clients.Proxy) {
	webscan.NewActiveFingerScan(a.ctx, target, proxy)
}

func (a *App) IsHighRisk(fingerprint string) bool {
	for name, wfe := range webscan.WorkFlowDB {
		if fingerprint == name {
			return len(wfe.PocsName) > 0
		}
	}
	return false
}

// 在漏扫开始时，需要轮询结果
// mode = 0  表示扫指纹漏洞， mode = 1 表示扫全漏洞
func (a *App) NucleiScanner(mode int, target string, fingerprints []string, nucleiPath string, interactsh bool, keywords, severity string) {
	nc := webscan.NewNucleiCaller(nucleiPath, interactsh, severity)
	nc.ReportDirStat()
	if mode == 0 {
		pe := nc.TargetBindFingerPocs(target, fingerprints)
		nc.CallerFP(a.ctx, pe)
	} else {
		keys := []string{}
		if keywords != "" {
			keys = strings.Split(keywords, ",")
		}
		nc.CallerAP(a.ctx, target, keys)
	}
}

func (a *App) NucleiEnabled(nucleiPath string) bool {
	return webscan.NewNucleiCaller(nucleiPath, false, "").Enabled()
}

func (a *App) WebPocLength() int {
	return len(webscan.ALLPoc())
}

// hunter

func (a *App) HunterTips(query string) *space.HunterTipsResult {
	return space.SearchHunterTips(query)
}

func (a *App) HunterSearch(api, query, pageSize, pageNum, times, asset string, deduplication bool) *space.HunterResult {
	hr := space.HunterApiSearch(api, query, pageSize, pageNum, times, asset, deduplication)
	time.Sleep(time.Second * 2)
	return hr
}

func (a *App) JSFind(target, customPrefix string) (fs *jsfind.FindSomething) {
	jsLinks := jsfind.ExtractJS(target)
	var wg sync.WaitGroup
	limiter := make(chan bool, 100)
	wg.Add(1)
	limiter <- true
	go func() {
		fs = jsfind.FindInfo(target, limiter, &wg)
	}()
	wg.Wait()
	u, _ := url.Parse(target)
	fs.JS = *jsfind.AppendSource(target, jsLinks)
	host := ""
	if customPrefix != "" {
		host = customPrefix
	} else {
		host = u.Scheme + "://" + u.Host
	}
	for _, jslink := range jsLinks {
		wg.Add(1)
		limiter <- true
		go func(js string) {
			newURL := host + "/" + js
			fs2 := jsfind.FindInfo(newURL, limiter, &wg)
			fs.IP_URL = append(fs.IP_URL, fs2.IP_URL...)
			fs.ChineseIDCard = append(fs.ChineseIDCard, fs2.ChineseIDCard...)
			fs.ChinesePhone = append(fs.ChinesePhone, fs2.ChinesePhone...)
			fs.SensitiveField = append(fs.SensitiveField, fs2.SensitiveField...)
			fs.APIRoute = append(fs.APIRoute, fs2.APIRoute...)
		}(jslink)
	}
	wg.Wait()
	fs.APIRoute = jsfind.RemoveDuplicatesInfoSource(fs.APIRoute)
	fs.ChineseIDCard = jsfind.RemoveDuplicatesInfoSource(fs.ChineseIDCard)
	fs.ChinesePhone = jsfind.RemoveDuplicatesInfoSource(fs.ChinesePhone)
	fs.SensitiveField = jsfind.RemoveDuplicatesInfoSource(fs.SensitiveField)
	fs.IP_URL = jsfind.FilterExt(jsfind.RemoveDuplicatesInfoSource(fs.IP_URL))
	return fs
}

func (a *App) WebIconMd5(target string) string {
	if _, err := url.Parse(target); err != nil {
		return ""
	}
	_, body, err := clients.NewSimpleGetRequest(target, clients.DefaultClient())
	if err != nil {
		return ""
	}
	hasher := md5.New()
	hasher.Write(body)
	sum := hasher.Sum(nil)
	return hex.EncodeToString(sum)
}

func (a *App) UseDruid(target string, attackType int, proxy clients.Proxy) (result []string) {
	druid := alibaba.NewDruid()
	if attackType == 1 {
		result, _ = druid.LoginBrute(target, clients.JudgeClient(proxy))
	} else {
		result, _ = druid.GetSession(target, clients.JudgeClient(proxy))
	}
	return
}
