package pyroscopesdk

import (
	"context"
	"fmt"
	"github.com/pyroscope-io/client/pyroscope"
	"runtime"
	"sync"
)

var (
	profileTool *pyroscope.Profiler
)

func Start(ctx context.Context, wg *sync.WaitGroup, appName, serverAddr string, logger pyroscope.Logger, openBlockProfile bool) {
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			if err := recover(); err != nil {
				fmt.Printf("pyroscope_start_err:%+v\n", err)
				return
			}
		}()
		if logger == nil {
			logger = pyroscope.StandardLogger
		}
		if openBlockProfile {
			runtime.SetMutexProfileFraction(100)
			runtime.SetBlockProfileRate(100)
		}
		var err error
		profileTool, err = pyroscope.Start(pyroscope.Config{
			ApplicationName: appName,
			// replace this with the address of pyroscope server
			ServerAddress: serverAddr,
			// you can disable logging by setting this to nil
			Logger: logger,
			// optionally, if authentication is enabled, specify the API key:
			// AuthToken: os.Getenv("PYROSCOPE_AUTH_TOKEN"),
			ProfileTypes: []pyroscope.ProfileType{
				// these profile types are enabled by default:
				pyroscope.ProfileCPU,
				pyroscope.ProfileAllocObjects,
				pyroscope.ProfileAllocSpace,
				pyroscope.ProfileInuseObjects,
				pyroscope.ProfileInuseSpace,
				// these profile types are optional:
				pyroscope.ProfileGoroutines,
				pyroscope.ProfileMutexCount,
				pyroscope.ProfileMutexDuration,
				pyroscope.ProfileBlockCount,
				pyroscope.ProfileBlockDuration,
			},
			SampleRate: 100,
		})
		if err != nil {
			panic(err)
			return
		}
		select {
		case <-ctx.Done():
			_ = profileTool.Stop()
		}
	}()
}

func Closed() {
	_ = profileTool.Stop()
}
