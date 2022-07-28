package crawlergo

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Tw1ps/crawlergo/pkg"
	"github.com/Tw1ps/crawlergo/pkg/config"
	"github.com/Tw1ps/crawlergo/pkg/logger"
	model2 "github.com/Tw1ps/crawlergo/pkg/model"
	"github.com/Tw1ps/crawlergo/pkg/tools"
	"github.com/sirupsen/logrus"
)

type Result struct {
	ReqList       []Request         `json:"req_list"`
	AllReqList    []Request         `json:"all_req_list"`
	AllDomainList []string          `json:"all_domain_list"`
	SubDomainList []string          `json:"sub_domain_list"`
	FoundMap      map[string]string `json:"found_map"`
}

type Request struct {
	Url     string                 `json:"url"`
	Method  string                 `json:"method"`
	Headers map[string]interface{} `json:"headers"`
	Data    string                 `json:"data"`
	Source  string                 `json:"source"`
}

const (
	DefaultLogLevel = "Info"
)

var (
	signalChan chan os.Signal
)

type Options struct {
	URLs       []string
	TaskConfig pkg.TaskConfig
	PostData   string
	LogLevel   string
}

func Run(opt *Options) *pkg.Result {
	signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	// 设置日志输出级别
	level, err := logrus.ParseLevel(opt.LogLevel)
	if err != nil {
		logger.Logger.Fatal(err)
	}
	logger.Logger.SetLevel(level)

	var targets []*model2.Request
	for _, _url := range opt.URLs {
		var req model2.Request
		url, err := model2.GetUrl(_url)
		if err != nil {
			logger.Logger.Error("parse url failed, ", err)
			continue
		}
		if opt.PostData != "" {
			req = model2.GetRequest(config.POST, url, getOption(opt.TaskConfig, opt.PostData))
		} else {
			req = model2.GetRequest(config.GET, url, getOption(opt.TaskConfig, opt.PostData))
		}
		req.Proxy = opt.TaskConfig.Proxy
		targets = append(targets, &req)
	}
	if opt.TaskConfig.Proxy != "" {
		logger.Logger.Info("request with proxy: ", opt.TaskConfig.Proxy)
	}

	if len(targets) == 0 {
		logger.Logger.Fatal("no validate target.")
	}

	if len(opt.TaskConfig.SearchKeywords) > 0 {
		logger.Logger.Infof("Search keyword: %v", opt.TaskConfig.SearchKeywords)
	}

	// 开始爬虫任务
	task, err := pkg.NewCrawlerTask(targets, opt.TaskConfig)
	if err != nil {
		logger.Logger.Error("create crawler task failed.")
		os.Exit(-1)
	}
	if len(targets) != 0 {
		logger.Logger.Info(fmt.Sprintf("Init crawler task, host: %s, max tab count: %d, max crawl count: %d.",
			targets[0].URL.Host, opt.TaskConfig.MaxTabsCount, opt.TaskConfig.MaxCrawlCount))
		logger.Logger.Info("filter mode: ", opt.TaskConfig.FilterMode)
	}

	// 提示自定义表单填充参数
	if len(opt.TaskConfig.CustomFormValues) > 0 {
		logger.Logger.Info("Custom form values, " + tools.MapStringFormat(opt.TaskConfig.CustomFormValues))
	}
	// 提示自定义表单填充参数
	if len(opt.TaskConfig.CustomFormKeywordValues) > 0 {
		logger.Logger.Info("Custom form keyword values, " + tools.MapStringFormat(opt.TaskConfig.CustomFormKeywordValues))
	}
	if _, ok := opt.TaskConfig.CustomFormValues["default"]; !ok {
		logger.Logger.Info("If no matches, default form input text: " + config.DefaultInputText)
		opt.TaskConfig.CustomFormValues["default"] = config.DefaultInputText
	}

	go handleExit(task)
	logger.Logger.Info("Start crawling.")
	task.Run()
	result := task.Result

	logger.Logger.Info(fmt.Sprintf("Task finished, %d results, %d requests, %d subdomains, %d domains found, %d keyword found.",
		len(result.ReqList), len(result.AllReqList), len(result.SubDomainList), len(result.AllDomainList), len(result.FoundMap)))

	return task.Result
}

func handleExit(t *pkg.CrawlerTask) {
	<-signalChan
	fmt.Println("exit ...")
	t.Pool.Tune(1)
	t.Pool.Release()
	t.Browser.Close()
	os.Exit(-1)
}

func getOption(taskConfig pkg.TaskConfig, postData string) model2.Options {
	var option model2.Options
	if postData != "" {
		option.PostData = postData
	}
	if taskConfig.ExtraHeadersString != "" {
		err := json.Unmarshal([]byte(taskConfig.ExtraHeadersString), &taskConfig.ExtraHeaders)
		if err != nil {
			logger.Logger.Fatal("custom headers can't be Unmarshal.")
			panic(err)
		}
		option.Headers = taskConfig.ExtraHeaders
	}
	return option
}
