package gologger

import (
	"context"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	Level_INFO    = "[INF]"
	Level_WARN    = "[WAR]"
	Level_ERROR   = "[ERR]"
	Level_DEBUG   = "[DEB]"
	Level_Success = "[SUC]"
)

type MsgInfo struct {
	Level string
	Msg   string
}

func Info(ctx context.Context, i interface{}) {
	runtime.EventsEmit(ctx, "gologger", &MsgInfo{
		Level: Level_INFO,
		Msg:   Msg(i),
	})
}

func Warning(ctx context.Context, i interface{}) {
	runtime.EventsEmit(ctx, "gologger", &MsgInfo{
		Level: Level_WARN,
		Msg:   Msg(i),
	})
}

func Error(ctx context.Context, i interface{}) {
	runtime.EventsEmit(ctx, "gologger", &MsgInfo{
		Level: Level_ERROR,
		Msg:   Msg(i),
	})
}

func Debug(ctx context.Context, i interface{}) {
	runtime.EventsEmit(ctx, "gologger", &MsgInfo{
		Level: Level_DEBUG,
		Msg:   Msg(i),
	})
}

func Success(ctx context.Context, i interface{}) {
	runtime.EventsEmit(ctx, "gologger", &MsgInfo{
		Level: Level_Success,
		Msg:   Msg(i),
	})
}

func Msg(i interface{}) string {
	return fmt.Sprintf("%s %v", currentTime(), i)
}

func currentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
