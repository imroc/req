package req

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// 解决request无法使用结构体添加query和form的问题
// 为什么要这样做?因为map是无序的,在某些情况下,参数顺序不一致,会导致请求失败和调试困难

// SetQueryParamsStruct 从结构体序列化为query参数，保持字段定义顺序
// 支持的标签: `query:"name"` 或 `json:"name"` 或 `form:"name"`
// 支持 `query:"-"` 忽略字段
// 支持 `query:"name,omitempty"` 忽略零值
func (r *Request) SetQueryParamsStruct(params any) *Request {
	if params == nil {
		return r
	}

	queryParams := r.marshalToUrlValues(params, "query")
	if r.QueryParams == nil {
		r.QueryParams = queryParams
	} else {
		// 合并到现有的查询参数中
		for key, values := range queryParams {
			for _, value := range values {
				r.QueryParams.Add(key, value)
			}
		}
	}
	return r
}

// SetFormDataStruct 从结构体序列化为form数据，保持字段定义顺序
// 支持的标签: `form:"name"` 或 `json:"name"` 或 `query:"name"`
// 支持 `form:"-"` 忽略字段
// 支持 `form:"name,omitempty"` 忽略零值
func (r *Request) SetFormDataStruct(params any) *Request {
	if params == nil {
		return r
	}

	formData := r.marshalToUrlValues(params, "form")
	if r.FormData == nil {
		r.FormData = formData
	} else {
		// 合并到现有的表单数据中
		for key, values := range formData {
			for _, value := range values {
				r.FormData.Add(key, value)
			}
		}
	}
	return r
}

// marshalToUrlValues 将结构体转换为url.Values，保持字段顺序
func (r *Request) marshalToUrlValues(params any, primaryTag string) url.Values {
	result := url.Values{}

	rv := reflect.ValueOf(params)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return result
		}
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		r.appendError(fmt.Errorf("params must be a struct or pointer to struct, got %T", params))
		return result
	}

	rt := rv.Type()

	// 按照字段定义顺序遍历
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		// 跳过非导出字段
		if !field.IsExported() {
			continue
		}

		// 获取标签名和选项
		tagName, omitempty := r.getFieldTag(field, primaryTag)
		if tagName == "-" {
			continue
		}

		// 只有在设置了omitempty且值为零值时才跳过
		if omitempty && r.isZeroValue(fieldValue) {
			continue
		}

		// 转换值为字符串
		values := r.convertToStringValues(fieldValue)
		for _, value := range values {
			result.Add(tagName, value)
		}
	}

	return result
}

// getFieldTag 获取字段的标签名和选项
func (r *Request) getFieldTag(field reflect.StructField, primaryTag string) (string, bool) {
	// 优先使用指定的标签
	if tag := field.Tag.Get(primaryTag); tag != "" {
		return r.parseTag(tag, field.Name)
	}

	// 回退标签顺序
	fallbackTags := []string{"json", "form", "query"}
	for _, tagName := range fallbackTags {
		if tagName == primaryTag {
			continue
		}
		if tag := field.Tag.Get(tagName); tag != "" {
			return r.parseTag(tag, field.Name)
		}
	}

	// 使用字段名的小写形式
	return strings.ToLower(field.Name), false
}

// parseTag 解析标签，返回名称和是否有omitempty选项
func (r *Request) parseTag(tag, fieldName string) (string, bool) {
	parts := strings.Split(tag, ",")
	name := strings.TrimSpace(parts[0])

	if name == "" {
		name = strings.ToLower(fieldName)
	}

	omitempty := false
	for i := 1; i < len(parts); i++ {
		if strings.TrimSpace(parts[i]) == "omitempty" {
			omitempty = true
			break
		}
	}

	return name, omitempty
}

// isZeroValue 检查值是否为零值
func (r *Request) isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !r.isZeroValue(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() && !r.isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// convertToStringValues 将值转换为字符串数组
func (r *Request) convertToStringValues(v reflect.Value) []string {
	switch v.Kind() {
	case reflect.String:
		return []string{v.String()}
	case reflect.Bool:
		return []string{strconv.FormatBool(v.Bool())}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []string{strconv.FormatInt(v.Int(), 10)}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return []string{strconv.FormatUint(v.Uint(), 10)}
	case reflect.Float32:
		return []string{strconv.FormatFloat(v.Float(), 'f', -1, 32)}
	case reflect.Float64:
		return []string{strconv.FormatFloat(v.Float(), 'f', -1, 64)}
	case reflect.Slice, reflect.Array:
		var result []string
		for i := 0; i < v.Len(); i++ {
			values := r.convertToStringValues(v.Index(i))
			result = append(result, values...)
		}
		return result
	case reflect.Ptr:
		if v.IsNil() {
			return []string{}
		}
		return r.convertToStringValues(v.Elem())
	case reflect.Interface:
		if v.IsNil() {
			return []string{}
		}
		return r.convertToStringValues(v.Elem())
	default:
		// 对于其他类型，使用fmt.Sprintf
		return []string{fmt.Sprintf("%v", v.Interface())}
	}
}
