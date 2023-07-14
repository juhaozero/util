package logs

import (
	"go.uber.org/zap"
)

func Info(method string, oper string, otherFields ...zap.Field) {
	fields := []zap.Field{zap.String("uid", oper)}
	fields = append(fields, otherFields...)
	Logger.Info(method, fields...)
}

func Debug(method string, oper string, otherFields ...zap.Field) {
	fields := []zap.Field{zap.String("uid", oper)}
	fields = append(fields, otherFields...)
	Logger.Debug(method, fields...)
}

func Warn(method string, oper string, otherFields ...zap.Field) {
	fields := []zap.Field{zap.String("uid", oper)}
	fields = append(fields, otherFields...)
	Logger.Warn(method, fields...)
}
func Err(method string, oper string, otherFields ...zap.Field) {
	fields := []zap.Field{zap.String("uid", oper)}
	fields = append(fields, otherFields...)
	Logger.Error(method, fields...)
}
