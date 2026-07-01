package log

// Config 日志配置。
type Config struct {
	// Level 日志级别: debug / info / warn / error / fatal / panic
	Level string
	// Format 输出格式: json / console
	Format string
	// LogPath 日志文件路径
	LogPath string
	// Filename 日志文件名称
	Filename string
	// MaxSize 单个日志文件最大体积（MB），超出后自动切分
	MaxSize int
	// MaxBackups 保留的历史日志文件数量
	MaxBackups int
	// MaxAge 日志文件最大保留天数
	MaxAge int
	// Compress 是否压缩已切分的旧日志
	Compress bool
	// Console 是否同时输出到控制台
	Console bool
}

// DefaultConfig 返回默认配置。
func DefaultConfig() Config {
	return Config{
		Level:      "info",
		Format:     "json",
		LogPath:    "logs",
		Filename:   "app.log",
		MaxSize:    100,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   true,
		Console:    true,
	}
}
