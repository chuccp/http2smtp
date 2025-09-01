package core

import (
	"github.com/chuccp/http2smtp/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

func initLogger(path string) (*zap.Logger, error) {
	writeFileCore, err := getFileLogWriter(path)
	if err != nil {
		return nil, err
	}
	core := zapcore.NewTee(writeFileCore, getStdoutLogWriter())
	return zap.New(core, zap.AddCaller()), nil
}

func getEncoder() zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout(util.TimestampFormat)
	return zapcore.NewJSONEncoder(config)
}

func getFileLogWriter(path string) (zapcore.Core, error) {
	logger := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     30,   //days
		Compress:   true, // disabled by default
	}
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, zapcore.AddSync(logger), zapcore.InfoLevel)
	return core, nil
}
func getStdoutLogWriter() zapcore.Core {
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, os.Stdout, zapcore.DebugLevel)
	return core
}
