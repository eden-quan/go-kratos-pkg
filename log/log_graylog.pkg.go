package logpkg

import (
	"fmt"
	"io"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
)

// ConfigGraylog ...
type ConfigGraylog struct {
	// Level 日志级别
	Level log.Level
	// CallerSkip 日志 runtime caller skips
	CallerSkip int

	GraylogConfig GraylogConfig
}

// GraylogConfig ...
type GraylogConfig struct {
	Facility      string `json:"facility"` // app.Name
	Proto         string `json:"proto"`    // tcp、udp、...
	Addr          string `json:"addr"`
	AsyncPoolSize int    `json:"async_pool_size"`
}

// Graylog ...
type Graylog struct {
	loggerHandler *zap.Logger
}

// NewGraylogLogger ...
func NewGraylogLogger(conf *ConfigGraylog, opts ...Option) (*Graylog, error) {
	handler := &Graylog{}
	if err := handler.initLogger(conf, opts...); err != nil {
		return handler, err
	}
	return handler, nil
}

// initLogger .
func (s *Graylog) initLogger(conf *ConfigGraylog, opts ...Option) (err error) {
	// 可选项
	option := options{
		writer:     nil,
		loggerKeys: DefaultLoggerKey(),
		timeFormat: DefaultTimeFormat,
	}
	for _, o := range opts {
		o(&option)
	}

	// 参考 zap.NewProductionEncoderConfig()
	encoderConf := zapcore.EncoderConfig{
		MessageKey:    LoggerKeyMessage.Value(),
		LevelKey:      LoggerKeyLevel.Value(),
		TimeKey:       LoggerKeyTime.Value(),
		NameKey:       LoggerKeyName.Value(),
		CallerKey:     LoggerKeyCaller.Value(),
		FunctionKey:   LoggerKeyFunction.Value(),
		StacktraceKey: LoggerKeyStacktrace.Value(),

		LineEnding:  zapcore.DefaultLineEnding,
		EncodeLevel: zapcore.CapitalLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(option.timeFormat))
		},
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		//EncodeCaller: zapcore.FullCallerEncoder,
	}
	SetZapLoggerKeys(&encoderConf, option.loggerKeys)

	// writer
	if option.writer == nil {
		option.writer, err = NewGraylogWriter(&conf.GraylogConfig)
		if err != nil {
			return err
		}
	}

	encoder := zapcore.NewJSONEncoder(encoderConf)
	zapCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(option.writer),
		zap.NewAtomicLevelAt(ToZapLevel(conf.Level)),
	)

	// logger
	callerSkip := DefaultCallerSkip
	if conf.CallerSkip > 0 {
		callerSkip = conf.CallerSkip
	}
	stacktraceLevel := zapcore.DPanicLevel
	s.loggerHandler = zap.New(zapCore,
		zap.WithCaller(true),
		zap.AddCallerSkip(callerSkip),
		zap.AddStacktrace(stacktraceLevel),
	)
	return err
}

// NewGraylogWriter log writer
func NewGraylogWriter(conf *GraylogConfig) (io.Writer, error) {
	var (
		writer io.Writer
	)

	switch conf.Proto {
	case "tcp", "TCP":
		graylog, err := gelf.NewTCPWriter(conf.Addr)
		if err != nil {
			return nil, err
		}
		if conf.Facility != "" {
			graylog.GelfWriter.Facility = conf.Facility
		}
		writer = graylog
	default:
		graylog, err := gelf.NewUDPWriter(conf.Addr)
		if err != nil {
			return nil, err
		}
		if conf.Facility != "" {
			graylog.GelfWriter.Facility = conf.Facility
		}
		writer = graylog
	}

	return NewAsyncWriter(writer, conf.AsyncPoolSize), nil
}

// sync zap.Logger.Sync
func (s *Graylog) sync() error {
	return s.loggerHandler.Sync()
}

// Close zap.Logger.Sync
func (s *Graylog) Close() error {
	return s.loggerHandler.Sync()
}

// Log .
func (s *Graylog) Log(level log.Level, keyvals ...interface{}) (err error) {
	if len(keyvals) == 0 {
		return err
	}
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "KEYVALS UNPAIRED")
	}

	// field
	var (
		msg  = "\n"
		data []zap.Field
	)
	for i := 0; i < len(keyvals); i += 2 {
		data = append(data, zap.Any(fmt.Sprint(keyvals[i]), keyvals[i+1]))
	}

	switch level {
	case log.LevelDebug:
		s.loggerHandler.Debug(msg, data...)
	case log.LevelInfo:
		s.loggerHandler.Info(msg, data...)
	case log.LevelWarn:
		s.loggerHandler.Warn(msg, data...)
	case log.LevelError:
		s.loggerHandler.Error(msg, data...)
	case log.LevelFatal:
		s.loggerHandler.Fatal(msg, data...)
	}
	return err
}
