package tlog

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

const (
	logLevalDebug = "DEBUG"
	logLevalWarn  = "WARN"
	logLevalInfo  = "INFO"
	logLevalError = "ERROR"
	// LogDateFmt 每天
	LogDateFmt = "2006-01-02"
	//每个小时
	LogHourFmt = "2006-01-02 15"
	// 每分钟
	LogMinFmt = "2006-01-02 15:04"
	nsFmt     = "2006-01-02 15:04:05.000000微妙"
)

const (
	LevalDebug = iota
	LevalInfo
	LevalWarn
	LevalError
)

const (
	// bufio 默认大小 10M
	writerBufferSize = 1024 * 1024 * 15
	// 默认日志文件大小 10G
	defLogFileSize = 1024 * 1024 * 1024 * 10
)

var (
	notifyChannel = make(chan bool)
	done          = make(chan struct{})
	levalMap      = make(map[int]string)
	defMux        = new(sync.Mutex)
	builderPool   sync.Pool
)

var _defLog *TBaseLog

type TBaseLog struct {
	file        *os.File //文件io
	writer      *bufio.Writer
	filePath    string    // 全路径
	initPath    string    // 初始化时候的路径
	timeLayout  string    // 格式化时间，创建文件的时候用到
	debugWriter io.Writer // 如果是debug ，则往终端写入
	ProjectName string    // 项目名称
	preFix      string    // log文件名前缀，用于创建多个log
	LogFileSize int64     // log 文件大小
	logIndex    int       // 文件名索引
	hook        Hook      // hook 函数
	common      TLCommarg // 公共参数
	wg          sync.WaitGroup
	mux         sync.Mutex
}

func init() {
	levalMap[LevalDebug] = logLevalDebug
	levalMap[LevalInfo] = logLevalInfo
	levalMap[LevalWarn] = logLevalWarn
	levalMap[LevalError] = logLevalError
	fn := func() interface{} {
		return strings.Builder{}
	}
	builderPool = sync.Pool{
		New: fn,
	}
}

/*
获取默认 log , 全局单利,只会创建一次
*/
func GetDefLog() *TBaseLog {
	defMux.Lock()
	defer defMux.Unlock()
	if _defLog == nil {
		pwd, _ := os.Getwd()
		_defLog = NewTBaseLog(WithLogDir(pwd+"/"), WithPrefixName("defalut_log"), WithProjectName("TLog"), WithTimeLayout(LogDateFmt))
	}
	return _defLog
}

/*
logDir 必须是文件夹
projectName 项目名称
logPreFix 日志前缀
一个 log 对象对应 一个日志文件 io 对应一把 mutex
*/
func NewTBaseLog(opts ...TBaseLogOption) *TBaseLog {
	fileLog := new(TBaseLog)
	fileLog = WithLogOption(fileLog, opts...)
	// 判断目录是否是文件夹,说明已经存在
	s, err := os.Stat(fileLog.initPath)
	if s == nil {
		fmt.Println("logDir 日志目录错误 ", fileLog.initPath)
		panic(err)
	}
	// 文件不存在 && 文件不是文件夹类型,创建文件夹
	if err != nil && !os.IsExist(err) && !s.IsDir() {
		// 创建文件夹
		err := os.MkdirAll(fileLog.initPath, 0644)
		if err != nil {
			fmt.Println("创建文件夹失败 logDir = ", fileLog.initPath)
			panic(err)
		}
	}
	if fileLog.ProjectName == "" {
		fileLog.ProjectName = "defalut_"
	}
	if fileLog.preFix == "" {
		fileLog.preFix = "_go"
	}
	// 设置日志文件大小
	if fileLog.LogFileSize <= 0 {
		// 默认1G切换文件
		fileLog.LogFileSize = defLogFileSize
	}
	infos, err := ioutil.ReadDir(fileLog.initPath)
	if err != nil {
		panic(err)
	}
	filenames := make([]string, 0, 20)
	tmFmt := time.Now().Format(fileLog.timeLayout)
	// 过滤文件名
	for i := 0; i < len(infos); i++ {
		info := infos[i]
		if !info.IsDir() && strings.Contains(info.Name(), tmFmt) {
			filenames = append(filenames, info.Name())
		}
	}
	if len(filenames) > 0 {
		tmpArr := make([]int, 0, 20)
		for i := 0; i < len(filenames); i++ {
			name := filenames[i]
			sub := strings.Split(name, "_")
			idx, _ := strconv.Atoi(strings.Replace(sub[len(sub)-1], ".json", "", -1))
			tmpArr = append(tmpArr, idx)
		}
		sort.Ints(tmpArr)
		lastFileName := filenames[len(filenames)-1]
		idx := tmpArr[len(tmpArr)-1]
		tmp := fileLog.initPath + lastFileName
		info, _ := os.Stat(tmp)
		if info.Size() < fileLog.LogFileSize {
			fileLog.logIndex = idx
		} else {
			fileLog.logIndex = idx + 1
		}
	} else {
		fileLog.logIndex = 1
	}
	// 创建文件
	filePath := fileLog.getFullFilePath(time.Now())
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR|os.O_SYNC, 0644)
	if err != nil {
		fmt.Println("日志文件创建失败 ", err, " filepath ", filePath)
		panic(err)
	}
	// 偏移到文件末尾，提高写入速度
	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		err = errors.Wrap(err, "")
		errMsg, fileline := GetErrFileInfo(err)
		fmt.Println("seek error ", errMsg, fileline)
		panic(err)
	}
	fileLog.file = file
	fileLog.filePath = filePath
	// 缓存io
	fileLog.writer = bufio.NewWriterSize(file, writerBufferSize)
	// debug 模式下写入 stderr 和 log 文件中
	fileLog.debugWriter = io.MultiWriter(os.Stderr, fileLog.writer)
	fileLog.startFlushWork()

	return fileLog
}

type TBaseLogOption func(t *TBaseLog)

// 设置日志目录
func WithLogDir(dir string) TBaseLogOption {
	return func(t *TBaseLog) {
		t.initPath = dir
	}
}

// 设置日志项目名称
func WithProjectName(name string) TBaseLogOption {
	return func(t *TBaseLog) {
		t.ProjectName = name
	}
}

// 设置日志文件名称前缀
func WithPrefixName(prefixName string) TBaseLogOption {
	return func(t *TBaseLog) {
		t.preFix = prefixName
	}
}

// 设置日志文件名称日期 layout
func WithTimeLayout(layout string) TBaseLogOption {
	return func(t *TBaseLog) {
		t.timeLayout = layout
	}
}

// WithCommon 设置日志公共参数
// 切勿再这个函数做大量耗时操作,因为每次 log 都会调用这个方法
func WithCommon(com TLCommarg) TBaseLogOption {
	return func(t *TBaseLog) {
		t.common = com
	}
}

// 设置日志hook 函数
// 切勿再这个函数做大量耗时操作,因为每次 log 都会调用这个方法,这个方法是 go出去的
func WithHook(hook Hook) TBaseLogOption {
	return func(t *TBaseLog) {
		t.hook = hook
	}
}

// WithLogFileSize 设置日志文件大小 函数
func WithLogFileSize(size int64) TBaseLogOption {
	return func(t *TBaseLog) {
		t.LogFileSize = size
	}
}

func (l *TBaseLog) WriteLog(level int, kv ...interface{}) {
	levelStr := levalMap[level]
	tmp := builderPool.Get()
	builder := tmp.(strings.Builder)
	builder.Reset()
	now := time.Now()
	builder.WriteString("{")
	sep := "\""
	builder.WriteString(sep)
	builder.WriteString("ProjectName")
	builder.WriteString(sep)
	builder.WriteString(":")
	builder.WriteString(sep)
	builder.WriteString(l.ProjectName)
	builder.WriteString(sep)
	builder.WriteString(",")
	builder.WriteString(sep)
	builder.WriteString("level")
	builder.WriteString(sep)
	builder.WriteString(":")
	builder.WriteString(sep)
	builder.WriteString(levelStr)
	builder.WriteString(sep)
	builder.WriteString(",")
	builder.WriteString(sep)
	builder.WriteString("cur_time")
	builder.WriteString(sep)
	builder.WriteString(":")
	builder.WriteString(sep)
	builder.WriteString(now.Format(nsFmt))
	builder.WriteString(sep)
	builder.WriteString(",")
	builder.WriteString(sep)
	builder.WriteString("cur_unix_time")
	builder.WriteString(sep)
	builder.WriteString(":")
	builder.WriteString(sep)
	builder.WriteString(strconv.FormatInt(now.Unix(), 10))
	builder.WriteString(sep)
	builder.WriteString(",")

	// 拼接公共参数
	if l.common != nil {
		c := l.common
		argsList := c.SetCommonArg()
		for i := 0; i < len(argsList); i++ {
			kv = append(kv, argsList[i])
		}
	}
	if len(kv) <= 0 {
		// 写文件
		l.writeToFile(builder, level)
		return
	}
	if len(kv)%2 != 0 {
		kv = append(kv, "unknown")
	}
	// 拼接kv
	for i := 0; i < len(kv); i++ {
		key := ""
		v := kv[i]
		key, ok := v.(string)
		if !ok && v != nil {
			key = getObjectStr(v)
		}
		if i%2 == 0 {
			builder.WriteString(sep)
			builder.WriteString(key)
			builder.WriteString(sep)
			builder.WriteString(":")
		} else {
			builder.WriteString(sep)
			builder.WriteString(key)
			builder.WriteString(sep)
			builder.WriteString(",")
		}
	}
	// 写文件
	l.writeToFile(builder, level)
}

// 写入到文件
func (l *TBaseLog) writeToFile(builder strings.Builder, leval int) {
	jsonStr := strings.TrimSuffix(builder.String(), ",")
	builderPool.Put(builder)
	jsonStr = jsonStr + "}\n"
	// hook 函数 , 判断是否实现
	if l.hook != nil {
		l.wg.Add(1)
		go func() {
			defer l.wg.Done()
			l.hook.Run(leval, jsonStr)
		}()
	}
	if leval == LevalDebug {
		l.mux.Lock()

		d := String2Bytes(jsonStr)
		// 写入终端
		_, _ = l.debugWriter.Write(d)
		defer l.mux.Unlock()
		return
	}
	// 写文件
	l.mux.Lock()
	defer l.mux.Unlock()
	_, err := l.writer.WriteString(jsonStr)
	if err != nil {
		fmt.Println("写入失败 writeToFile_err", err)
		return
	}
}

func getObjectStr(obj interface{}) (str string) {
	// 判断是否实现String()接口
	sst, ok := obj.(fmt.Stringer)
	if ok {
		str = sst.String()
		if str == "" {
			str = fmt.Sprintf("%+v", obj)
		}
		return
	}
	switch data := obj.(type) {
	case bool:
		str = strconv.FormatBool(data)
	case string:
		str = data
	case int:
		str = strconv.FormatInt(int64(data), 10)
	case int8:
		str = strconv.FormatInt(int64(data), 10)
	case int16:
		str = strconv.FormatInt(int64(data), 10)
	case int32:
		str = strconv.FormatInt(int64(data), 10)
	case int64:
		str = strconv.FormatInt(data, 10)
	case uint:
		str = strconv.FormatUint(uint64(data), 10)
	case uint8:
		str = strconv.FormatUint(uint64(data), 10)
	case uint16:
		str = strconv.FormatUint(uint64(data), 10)
	case uint32:
		str = strconv.FormatUint(uint64(data), 10)
	case uint64:
		str = strconv.FormatUint(data, 10)
	case []string:
		subBuild := new(strings.Builder)
		subBuild.WriteString("[ ")
		strList := data
		for i := 0; i < len(strList); i++ {
			subBuild.WriteString("\"")
			subBuild.WriteString(strList[i])
			subBuild.WriteString("\"")
			subBuild.WriteString(",")
		}
		subBuild.WriteString(" ]")
		str = subBuild.String()
	case float64:
		str = strconv.FormatFloat(data, 'f', 10, 64)
	case float32:
		str = strconv.FormatFloat(float64(data), 'f', 10, 64)
	case map[string]string:
		mBuild := new(strings.Builder)
		mBuild.WriteString("{ ")
		for k, v := range data {
			mBuild.WriteString(" \"")
			mBuild.WriteString(k)
			mBuild.WriteString("\":")
			mBuild.WriteString("\"")
			mBuild.WriteString(v)
			mBuild.WriteString("\",")
		}
		mBuild.WriteString("}")
		str = mBuild.String()
	case []byte:
		// 这种性能高不少
		str = *(*string)(unsafe.Pointer(&data))
	default:
		str = fmt.Sprintf("%+v", data)
	}

	return
}

// 将log 刷新到磁盘,每10秒同步一次
func (l *TBaseLog) startFlushWork() {
	go func() {
		timer := time.NewTimer(time.Second * 1)
		defer func() {
			fmt.Println("HBaseLog closed")
			l.wg.Wait()
			done <- struct{}{}
		}()
		for {
			select {
			case <-timer.C:
				l.Sync()
				l.checkFile()
				timer.Reset(time.Second * 1)
			case <-notifyChannel:
				l.closed()
				timer.Stop()
				return
			default:
				runtime.Gosched()
			}
			runtime.Gosched()
		}
	}()
}

// 关闭文件IO ，回收资源
func (l *TBaseLog) closed() {
	// 关闭文件通知，查看缓存里面是否有数据,有数据继续写入，写入完成在退出
	// 刷新到磁盘
	l.Sync()
	fmt.Println("关闭channel ", l.writer.Size())
	err := l.file.Close()
	if err != nil {
		fmt.Println("closed file err ", err)
		return
	}
}

func (l *TBaseLog) Sync() {
	if l.writer.Buffered() <= 0 {
		return
	}
	l.mux.Lock()
	defer l.mux.Unlock()
	// 再刷新
	err := l.writer.Flush()
	if err != nil {
		fmt.Println("l.writer.Flush()_err ", err)
		return
	}
}

// 日志文件格式 projectname_date_log.json
// 写入文件前先检验一下,文件名 是否包含日期,默认是一天一个文件
// 不包含日期格式,写入文件,关闭文件io,重新赋值文件io
func (l *TBaseLog) checkFile() {
	l.mux.Lock()
	defer l.mux.Unlock()
	path := l.getFullFilePath(time.Now())
	// 获取文件大小
	fileinfo, err := os.Stat(l.filePath)
	if err != nil {
		err = errors.Wrap(err, "")
		l.Error(err)
		return
	}
	// 大小超过限制 ,切割文件
	// 判断当前文件名是否需要按特定时间切割文件
	if !(fileinfo.Size() > l.LogFileSize || l.filePath != path) {
		return
	}
	l.logIndex = l.logIndex + 1
	// 重新设置文件名
	path = l.getFullFilePath(time.Now())
	// 将剩余的写入磁盘
	err = l.writer.Flush()
	if err != nil {
		err = errors.Wrap(err, "")
		msg, fileline := GetErrFileInfo(err)
		fmt.Println("文件同步失败 ", msg, fileline)
		return
	}
	// 关闭文件IO
	err = l.file.Close()
	if err != nil {
		err = errors.Wrap(err, "")
		msg, fileline := GetErrFileInfo(err)
		fmt.Println("closed file io error ", msg, fileline)
		return
	}
	// 创建文件
	fd, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR|os.O_SYNC, 0644)
	if err != nil {
		err = errors.Wrap(err, "")
		msg, fileline := GetErrFileInfo(err)
		fmt.Println("OpenFile  io error ", msg, fileline)
		return
	}
	l.file = fd
	l.writer = bufio.NewWriterSize(fd, writerBufferSize)
	l.debugWriter = io.MultiWriter(os.Stderr, l.writer)
	l.filePath = path
}

// 获取日志全路径
func (l *TBaseLog) getFullFilePath(t time.Time) string {
	filePath := l.initPath + l.preFix + "_" + l.ProjectName + "_" + t.Format(l.timeLayout) + "_log" + "_" + strconv.Itoa(l.logIndex) + ".json"
	return filePath
}

// GetErrFileInfo 获取错误行号
func GetErrFileInfo(err error) (errMsg, errFile string) {
	if err == nil {
		return
	}
	errMsg = err.Error()
	frame := fmt.Sprintf("%+v", err)
	// 获取函数栈帧
	lines := strings.Split(frame, "\n")
	if len(lines) >= 3 {
		// 获取错误行号
		errLine := getErrorLine(lines)
		// 不包含行号,说明错误行号获取错了
		if !strings.Contains(errLine, ":") {
			return
		}
		pwd, _ := os.Getwd()
		// 脱敏,去除具体路径
		errLine = strings.Replace(errLine, "\t"+pwd+"/", "", -1)
		// 错误产生的文件
		errFile = errLine
	}
	return
}

// 找到错误的行
func getErrorLine(lines []string) string {
	str := ""
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], ".go:") {
			str = lines[i]
			break
		}
	}
	return str
}

// GetFuncFrame 因为 runtime.Caller 很耗性能，所以在需要使用的时候才调用，比如有error的时候才使用
func GetFuncFrame(skip int) (funcName string, lineNum int, fileName string) {
	// 获取文件行号
	pc, file, line, _ := runtime.Caller(skip)
	fun := runtime.FuncForPC(pc)
	pwd, _ := os.Getwd()
	// 脱敏,去除具体路径
	file = strings.Replace(file, pwd+"/", "", -1)
	funcName = fun.Name()
	lineNum = line
	fileName = file
	return
}

func (l *TBaseLog) Printf(fmtStr string, args ...interface{}) {
	funcName, line, file := GetFuncFrame(2)
	t := time.Now().Format("2006-01-02 15:04:05.000000000")
	str := fmt.Sprintf("%s 函数名:%s 行号:%d 文件名:%s ", t, funcName, line, file)
	str = str + fmtStr
	fmt.Printf(str, args...)
}

func (l *TBaseLog) Println(args ...interface{}) {
	funcName, line, file := GetFuncFrame(2)
	t := time.Now().Format("2006-01-02 15:04:05.000000000")
	str := fmt.Sprintf("%s 函数名:%s 行号:%d 文件名:%s ", t, funcName, line, file)
	tmpArr := make([]interface{}, 0, len(args)*2)
	tmpArr = append(tmpArr, str)
	for i := 0; i < len(args); i++ {
		if args != nil && args[i] != nil {
			tmpArr = append(tmpArr, args[i])
		}
	}
	tmpArr = append(tmpArr, "\n")
	fmt.Print(tmpArr...)
}

func (l *TBaseLog) Debug(kv ...interface{}) {
	l.WriteLog(LevalDebug, kv...)
}

func (l *TBaseLog) Info(kv ...interface{}) {
	l.WriteLog(LevalInfo, kv...)
}

func (l *TBaseLog) Warn(kv ...interface{}) {
	l.WriteLog(LevalWarn, kv...)
}

func (l *TBaseLog) Error(err error, kv ...interface{}) {
	// 获取错误产生的文件行号，及错误文件
	errMsg, file := GetErrFileInfo(err)
	kv = append(kv, "errMsg")
	kv = append(kv, errMsg)
	kv = append(kv, "errFile")
	kv = append(kv, file)
	l.WriteLog(LevalError, kv...)
}

func (l *TBaseLog) SetHook(hk Hook) {
	l.hook = hk
}

func (l *TBaseLog) SetCommonArg(cm TLCommarg) {
	l.common = cm
}

// String2Bytes 这种转换性能要高出不少,少了内存拷贝
func String2Bytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

// Bytes2String 这种转换性能要高出不少,少了内存拷贝
func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
