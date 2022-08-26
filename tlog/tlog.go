package tlog

type TLog interface {
	// Debug debug 模式下 会输出到 终端 和 log 文件中
	Debug(category string, kv ...interface{})
	Info(category string, kv ...interface{})
	Warn(category string, kv ...interface{})
	Error(category string, err error, kv ...interface{})
	Printf(fmt string, args ...interface{})
	Println(args ...interface{})
}

type Hook interface {
	Run(leval int, msg string)
}

// TLCommarg 公共参数
type TLCommarg interface {
	// 设置公共参数
	SetCommonArg() (args []interface{})
}

// 默认实现 defLog
// debug 函数除了会写入日志文件之外，还会写入
func Debug(kv ...interface{}) {
	if _defLog == nil {
		_ = GetDefLog()
	}
	_defLog.WriteLog(LevalDebug, kv...)
}

func Info(kv ...interface{}) {
	if _defLog == nil {
		_ = GetDefLog()
	}
	_defLog.WriteLog(LevalInfo, kv...)
}

func Warn(kv ...interface{}) {
	if _defLog == nil {
		_ = GetDefLog()
	}
	_defLog.WriteLog(LevalWarn, kv...)
}

func Error(err error, kv ...interface{}) {
	if _defLog == nil {
		_ = GetDefLog()
	}
	// 获取错误产生的文件行号，及错误文件
	errMsg, file := GetErrFileInfo(err)
	_defLog.WriteLog(LevalError, "errMsg", errMsg, "errFile", file)
}

func Printf(fmtStr string, args ...interface{}) {
	if _defLog == nil {
		_ = GetDefLog()
	}
	_defLog.Printf(fmtStr, args...)
}

func Println(args ...interface{}) {
	if _defLog == nil {
		_ = GetDefLog()
	}
	_defLog.Println(args...)
}

//设置_defLog 的 Hook 函数
func SetDefalutHook(hk Hook) {
	if _defLog == nil {
		_ = GetDefLog()
	}
	_defLog.hook = hk
}

// 设置_defLog 的 公共参数 函数
func SetDefalutCommonArg(cm TLCommarg) {
	if _defLog == nil {
		_ = GetDefLog()
	}
	_defLog.common = cm
}

/*	设置 log 参数 */
func WithLogOption(t *TBaseLog, opts ...TBaseLogOption) *TBaseLog {
	if t == nil {
		return nil
	}
	for i := 0; i < len(opts); i++ {
		opt := opts[i]
		opt(t)
	}
	return t
}

// Closed main 函数退出的时候记得调用此函数，关闭文件ios
func Closed() {
	close(notifyChannel)
	// 等待log协程退出
	<-done
}
