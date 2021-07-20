package logger

import (
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.SugaredLogger
)

func init() {
	_ = Configure(DefaultOptions())
}

// Must be called once at process startup.
func Configure(options *Options) error {
	logLevel := options.GetOutputLevel()
	encoder := getEncoder(options)
	writer, err := getWriter(options)
	if err != nil {
		return err
	}

	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoder), writer, logLevel)
	raw := zap.New(core, zap.AddCallerSkip(1))
	logger = raw.Sugar()

	return nil
}

func getEncoder(options *Options) zapcore.EncoderConfig {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeTime:     formatDate,
	}

	if options.LocalTime {
		encoderConfig.EncodeTime = formatLocalDate
	}

	return encoderConfig
}

func buildTimeFormat(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	year, month, day := t.Date()
	hour, minute, second := t.Clock()
	micros := t.Nanosecond() / 1000

	buf := make([]byte, 27)

	buf[0] = byte((year/1000)%10) + '0'
	buf[1] = byte((year/100)%10) + '0'
	buf[2] = byte((year/10)%10) + '0'
	buf[3] = byte(year%10) + '0'
	buf[4] = '-'
	buf[5] = byte((month)/10) + '0'
	buf[6] = byte((month)%10) + '0'
	buf[7] = '-'
	buf[8] = byte((day)/10) + '0'
	buf[9] = byte((day)%10) + '0'
	buf[10] = 'T'
	buf[11] = byte((hour)/10) + '0'
	buf[12] = byte((hour)%10) + '0'
	buf[13] = ':'
	buf[14] = byte((minute)/10) + '0'
	buf[15] = byte((minute)%10) + '0'
	buf[16] = ':'
	buf[17] = byte((second)/10) + '0'
	buf[18] = byte((second)%10) + '0'
	buf[19] = '.'
	buf[20] = byte((micros/100000)%10) + '0'
	buf[21] = byte((micros/10000)%10) + '0'
	buf[22] = byte((micros/1000)%10) + '0'
	buf[23] = byte((micros/100)%10) + '0'
	buf[24] = byte((micros/10)%10) + '0'
	buf[25] = byte((micros)%10) + '0'
	buf[26] = 'Z'

	enc.AppendString(string(buf))
}

func formatDate(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	buildTimeFormat(t.UTC(), enc)
}

func formatLocalDate(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	buildTimeFormat(t.Local(), enc)
}

func getWriter(options *Options) (zapcore.WriteSyncer, error) {
	var rotateSink zapcore.WriteSyncer
	if options.RotateOutputPath != "" {
		rotateSink = zapcore.AddSync(&lumberjack.Logger{
			Filename:   options.RotateOutputPath,
			MaxSize:    options.RotationMaxSize,
			MaxBackups: options.RotationMaxBackups,
			MaxAge:     options.RotationMaxAge,
		})
	}

	var outputSink zapcore.WriteSyncer
	var err error
	if len(options.OutputPaths) > 0 {
		outputSink, _, err = zap.Open(options.OutputPaths...)
		if err != nil {
			return nil, err
		}
	}

	var writer zapcore.WriteSyncer
	if rotateSink != nil && outputSink != nil {
		writer = zapcore.NewMultiWriteSyncer(outputSink, rotateSink)
	} else if rotateSink != nil {
		writer = rotateSink
	} else {
		writer = outputSink
	}

	return writer, nil
}

func GetLogger() *zap.SugaredLogger {
	return logger
}
