// Package output 提供输出格式化功能
package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// Format 输出格式类型
type Format string

const (
	// FormatText 文本格式
	FormatText Format = "text"
	
	// FormatJSON JSON格式
	FormatJSON Format = "json"
)

// Formatter 格式化器接口
type Formatter interface {
	// Format 格式化数据
	Format(data interface{}) (string, error)
	
	// Print 打印格式化后的数据
	Print(data interface{}) error
}

// TextFormatter 文本格式化器
type TextFormatter struct{}

// NewTextFormatter 创建文本格式化器
// 返回值：
//   - *TextFormatter: 文本格式化器实例
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

// Format 格式化数据为文本
// 参数：
//   - data: 要格式化的数据
// 返回值：
//   - string: 格式化后的文本
//   - error: 格式化过程中的错误
func (f *TextFormatter) Format(data interface{}) (string, error) {
	// 简单的文本格式化
	return fmt.Sprintf("%+v", data), nil
}

// Print 打印文本格式的数据
// 参数：
//   - data: 要打印的数据
// 返回值：
//   - error: 打印过程中的错误
func (f *TextFormatter) Print(data interface{}) error {
	text, err := f.Format(data)
	if err != nil {
		return err
	}
	
	fmt.Fprintln(os.Stdout, text)
	return nil
}

// JSONFormatter JSON格式化器
type JSONFormatter struct {
	indent bool // 是否缩进
}

// NewJSONFormatter 创建JSON格式化器
// 参数：
//   - indent: 是否使用缩进
// 返回值：
//   - *JSONFormatter: JSON格式化器实例
func NewJSONFormatter(indent bool) *JSONFormatter {
	return &JSONFormatter{
		indent: indent,
	}
}

// Format 格式化数据为JSON
// 参数：
//   - data: 要格式化的数据
// 返回值：
//   - string: 格式化后的JSON字符串
//   - error: 格式化过程中的错误
func (f *JSONFormatter) Format(data interface{}) (string, error) {
	var jsonData []byte
	var err error
	
	if f.indent {
		jsonData, err = json.MarshalIndent(data, "", "  ")
	} else {
		jsonData, err = json.Marshal(data)
	}
	
	if err != nil {
		return "", err
	}
	
	return string(jsonData), nil
}

// Print 打印JSON格式的数据
// 参数：
//   - data: 要打印的数据
// 返回值：
//   - error: 打印过程中的错误
func (f *JSONFormatter) Print(data interface{}) error {
	jsonStr, err := f.Format(data)
	if err != nil {
		return err
	}
	
	fmt.Fprintln(os.Stdout, jsonStr)
	return nil
}

// GetFormatter 根据格式类型获取格式化器
// 参数：
//   - format: 格式类型
// 返回值：
//   - Formatter: 格式化器实例
func GetFormatter(format Format) Formatter {
	switch format {
	case FormatJSON:
		return NewJSONFormatter(true)
	default:
		return NewTextFormatter()
	}
}


