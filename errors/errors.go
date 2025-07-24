package errors

import "fmt"

// ErrorType 定義錯誤類型
type ErrorType int

const (
	// NetworkError 網路相關錯誤
	NetworkError ErrorType = iota
	// ParseError 解析相關錯誤
	ParseError
	// FileError 檔案相關錯誤
	FileError
	// ConfigError 配置相關錯誤
	ConfigError
	// ValidationError 驗證相關錯誤
	ValidationError
)

// String 返回錯誤類型的字串表示
func (et ErrorType) String() string {
	switch et {
	case NetworkError:
		return "NetworkError"
	case ParseError:
		return "ParseError"
	case FileError:
		return "FileError"
	case ConfigError:
		return "ConfigError"
	case ValidationError:
		return "ValidationError"
	default:
		return "UnknownError"
	}
}

// CrawlerError 爬蟲自定義錯誤類型
type CrawlerError struct {
	Type    ErrorType
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error 實現 error 介面
func (e *CrawlerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type.String(), e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type.String(), e.Message)
}

// Unwrap 實現錯誤解包，支援 Go 1.13+ 的錯誤鏈
func (e *CrawlerError) Unwrap() error {
	return e.Cause
}

// WithContext 添加上下文資訊
func (e *CrawlerError) WithContext(key string, value interface{}) *CrawlerError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// GetContext 獲取上下文資訊
func (e *CrawlerError) GetContext(key string) (interface{}, bool) {
	if e.Context == nil {
		return nil, false
	}
	value, exists := e.Context[key]
	return value, exists
}

// Is 實現錯誤類型判斷，支援 Go 1.13+ 的 errors.Is
func (e *CrawlerError) Is(target error) bool {
	if target, ok := target.(*CrawlerError); ok {
		return e.Type == target.Type
	}
	return false
}

// NewCrawlerError 建立基本爬蟲錯誤
func NewCrawlerError(errorType ErrorType, message string) *CrawlerError {
	return &CrawlerError{
		Type:    errorType,
		Message: message,
	}
}

// NewCrawlerErrorWithCause 建立帶原因的爬蟲錯誤
func NewCrawlerErrorWithCause(errorType ErrorType, message string, cause error) *CrawlerError {
	return &CrawlerError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}

// 便利函數建立不同類型的錯誤

// NewNetworkError 建立網路錯誤
func NewNetworkError(message string, cause error) *CrawlerError {
	return NewCrawlerErrorWithCause(NetworkError, message, cause)
}

// NewParseError 建立解析錯誤
func NewParseError(message string, cause error) *CrawlerError {
	return NewCrawlerErrorWithCause(ParseError, message, cause)
}

// NewFileError 建立檔案錯誤
func NewFileError(message string, cause error) *CrawlerError {
	return NewCrawlerErrorWithCause(FileError, message, cause)
}

// NewConfigError 建立配置錯誤
func NewConfigError(message string, cause error) *CrawlerError {
	return NewCrawlerErrorWithCause(ConfigError, message, cause)
}

// NewValidationError 建立驗證錯誤
func NewValidationError(message string) *CrawlerError {
	return NewCrawlerError(ValidationError, message)
}

// IsNetworkError 檢查是否為網路錯誤
func IsNetworkError(err error) bool {
	if crawlerErr, ok := err.(*CrawlerError); ok {
		return crawlerErr.Type == NetworkError
	}
	return false
}

// IsParseError 檢查是否為解析錯誤
func IsParseError(err error) bool {
	if crawlerErr, ok := err.(*CrawlerError); ok {
		return crawlerErr.Type == ParseError
	}
	return false
}

// IsFileError 檢查是否為檔案錯誤
func IsFileError(err error) bool {
	if crawlerErr, ok := err.(*CrawlerError); ok {
		return crawlerErr.Type == FileError
	}
	return false
}

// IsConfigError 檢查是否為配置錯誤
func IsConfigError(err error) bool {
	if crawlerErr, ok := err.(*CrawlerError); ok {
		return crawlerErr.Type == ConfigError
	}
	return false
}

// IsValidationError 檢查是否為驗證錯誤
func IsValidationError(err error) bool {
	if crawlerErr, ok := err.(*CrawlerError); ok {
		return crawlerErr.Type == ValidationError
	}
	return false
}
