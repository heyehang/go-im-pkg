package util

type PanicErr func(err any)

// 捕获err go 协程,callBack 只有panic的时候才会有
func CaptureErrGo(function func(), callBack PanicErr) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				if callBack != nil {
					callBack(err)
				}
			}
		}()
		function()
	}()
}
