package logger

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
)

const (
	debugLevel = "debug"
	infoLevel  = "info"
	warnLevel  = "warn"
	errorLevel = "error"

	defaultOutputLevel        = infoLevel
	defaultOutputPath         = "stdout"
	defaultRotateOutputPath   = "/tmp/nacos-k8s-sync.log"
	defaultRotationMaxAge     = 30
	defaultRotationMaxSize    = 100 * 1024 * 1024
	defaultRotationMaxBackups = 15
)

var levelMap = map[string]zapcore.Level{
	debugLevel: zapcore.DebugLevel,
	infoLevel:  zapcore.InfoLevel,
	warnLevel:  zapcore.WarnLevel,
	errorLevel: zapcore.ErrorLevel,
}

type Options struct {
	OutputLevel string

	OutputPaths []string

	RotateOutputPath string

	RotationMaxSize int

	RotationMaxAge int

	RotationMaxBackups int

	// localTime determines whether the time format of log is local time format.
	// Default is true
	LocalTime bool
}

func DefaultOptions() *Options {
	return &Options{
		OutputLevel:        defaultOutputLevel,
		OutputPaths:        []string{defaultOutputPath},
		RotateOutputPath:   defaultRotateOutputPath,
		RotationMaxSize:    defaultRotationMaxSize,
		RotationMaxAge:     defaultRotationMaxAge,
		RotationMaxBackups: defaultRotationMaxBackups,
		LocalTime:          true,
	}
}

func (o *Options) GetOutputLevel() zapcore.Level {
	if level, exist := levelMap[o.OutputLevel]; exist {
		return level
	}

	return levelMap[defaultOutputLevel]
}

func (o *Options) AttachCobraFlags(cmd *cobra.Command) {
	o.AttachFlags(
		cmd.Flags().StringArrayVar,
		cmd.Flags().StringVar,
		cmd.Flags().IntVar,
		cmd.Flags().BoolVar)
}

func (o *Options) AttachFlags(
	stringArrayVar func(p *[]string, name string, value []string, usage string),
	stringVar func(p *string, name string, value string, usage string),
	intVar func(p *int, name string, value int, usage string),
	boolVar func(p *bool, name string, value bool, usage string)) {

	stringArrayVar(&o.OutputPaths, "log_target", o.OutputPaths,
		"The set of paths where to output the log. This can be any path as well as the special values stdout and stderr")

	stringVar(&o.RotateOutputPath, "log_rotate", o.RotateOutputPath,
		"The path for the optional rotating log file")

	intVar(&o.RotationMaxAge, "log_rotate_max_age", o.RotationMaxAge,
		"The maximum age in days of a log file beyond which the file is rotated (0 indicates no limit)")

	intVar(&o.RotationMaxSize, "log_rotate_max_size", o.RotationMaxSize,
		"The maximum size in megabytes of a log file beyond which the file is rotated")

	intVar(&o.RotationMaxBackups, "log_rotate_max_backups", o.RotationMaxBackups,
		"The maximum number of log file backups to keep before older files are deleted (0 indicates no limit)")

	boolVar(&o.LocalTime, "local_time", o.LocalTime,
		"Whether to use local time as time format for log.")

	stringVar(&o.OutputLevel, "log_output_level", o.OutputLevel,
		"Can be denoted as debug, info, warn, or error.")
}
