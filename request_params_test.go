package req

import (
	"fmt"
	"testing"
)

// 测试结构体
type QueryParams struct {
	Name    string   `query:"name"`
	Age     int      `query:"age,omitempty"`
	Active  bool     `query:"active"`
	Tags    []string `query:"tags"`
	Score   float64  `query:"score"`
	Ignored string   `query:"-"`
	NoTag   string   // 应该使用小写字段名
	Empty   string   `query:"empty,omitempty"` // 应该被忽略
}

type FormData struct {
	Username string `form:"username"`
	Password string `form:"password"`
	Remember bool   `form:"remember"`
	Age      int    `form:"age,omitempty"`
	Empty    string `form:"empty,omitempty"`
}

type JsonTags struct {
	Name  string `json:"name"`
	Value int    `json:"value,omitempty"`
}

func TestSetQueryParamsMarshal(t *testing.T) {
	client := C()

	// 测试基本功能
	params := QueryParams{
		Name:    "john",
		Age:     25,
		Active:  true,
		Tags:    []string{"tag1", "tag2"},
		Score:   98.5,
		Ignored: "should_be_ignored",
		NoTag:   "notag_value",
		Empty:   "", // 应该被omitempty忽略
	}

	req := client.R().SetQueryParamsStruct(params)

	// 检查query参数
	queryParams := req.QueryParams

	// 验证基本字段
	if queryParams.Get("name") != "john" {
		t.Errorf("Expected name=john, got %s", queryParams.Get("name"))
	}

	if queryParams.Get("age") != "25" {
		t.Errorf("Expected age=25, got %s", queryParams.Get("age"))
	}

	if queryParams.Get("active") != "true" {
		t.Errorf("Expected active=true, got %s", queryParams.Get("active"))
	}

	if queryParams.Get("score") != "98.5" {
		t.Errorf("Expected score=98.5, got %s", queryParams.Get("score"))
	}

	// 验证数组字段
	tags := queryParams["tags"]
	if len(tags) != 2 || tags[0] != "tag1" || tags[1] != "tag2" {
		t.Errorf("Expected tags=[tag1, tag2], got %v", tags)
	}

	// 验证忽略字段
	if queryParams.Get("ignored") != "" {
		t.Errorf("Expected ignored field to be empty, got %s", queryParams.Get("ignored"))
	}

	// 验证没有标签的字段
	if queryParams.Get("notag") != "notag_value" {
		t.Errorf("Expected notag=notag_value, got %s", queryParams.Get("notag"))
	}

	// 验证omitempty字段
	if queryParams.Get("empty") != "" {
		t.Errorf("Expected empty field to be omitted, got %s", queryParams.Get("empty"))
	}

	fmt.Printf("Query params: %v\n", queryParams)
}

func TestSetFormDataMarshal(t *testing.T) {
	client := C()

	formData := FormData{
		Username: "user123",
		Password: "secret",
		Remember: true,
		Age:      0,  // 应该被omitempty忽略
		Empty:    "", // 应该被omitempty忽略
	}

	req := client.R().SetFormDataStruct(formData)

	// 检查form数据
	form := req.FormData

	if form.Get("username") != "user123" {
		t.Errorf("Expected username=user123, got %s", form.Get("username"))
	}

	if form.Get("password") != "secret" {
		t.Errorf("Expected password=secret, got %s", form.Get("password"))
	}

	if form.Get("remember") != "true" {
		t.Errorf("Expected remember=true, got %s", form.Get("remember"))
	}

	// 验证omitempty字段（Age为0且有omitempty标签应该被忽略）
	if form.Get("age") != "" {
		t.Errorf("Expected age to be omitted, got %s", form.Get("age"))
	}

	if form.Get("empty") != "" {
		t.Errorf("Expected empty to be omitted, got %s", form.Get("empty"))
	}

	fmt.Printf("Form data: %v\n", form)
}

func TestJsonTagsFallback(t *testing.T) {
	client := C()

	params := JsonTags{
		Name:  "test",
		Value: 0, // 应该被omitempty忽略
	}

	req := client.R().SetQueryParamsStruct(params)

	// 应该回退到json标签
	if req.QueryParams.Get("name") != "test" {
		t.Errorf("Expected name=test, got %s", req.QueryParams.Get("name"))
	}

	// 验证omitempty
	if req.QueryParams.Get("value") != "" {
		t.Errorf("Expected value to be omitted, got %s", req.QueryParams.Get("value"))
	}

	fmt.Printf("JSON tags query params: %v\n", req.QueryParams)
}

func TestNilParams(t *testing.T) {
	client := C()

	// 测试nil参数
	req := client.R().SetQueryParamsStruct(nil)
	if req.QueryParams != nil && len(req.QueryParams) > 0 {
		t.Errorf("Expected empty query params for nil input")
	}

	req = client.R().SetFormDataStruct(nil)
	if req.FormData != nil && len(req.FormData) > 0 {
		t.Errorf("Expected empty form data for nil input")
	}
}

func TestPointerParams(t *testing.T) {
	client := C()

	params := &QueryParams{
		Name:   "pointer_test",
		Active: true,
	}

	req := client.R().SetQueryParamsStruct(params)

	if req.QueryParams.Get("name") != "pointer_test" {
		t.Errorf("Expected name=pointer_test, got %s", req.QueryParams.Get("name"))
	}
}

func TestMergeWithExisting(t *testing.T) {
	client := C()

	// 先设置一些手动的参数
	req := client.R().SetQueryParam("manual", "value")

	params := QueryParams{
		Name: "merge_test",
	}

	// 再通过marshal添加参数
	req.SetQueryParamsStruct(params)

	// 验证两种参数都存在
	if req.QueryParams.Get("manual") != "value" {
		t.Errorf("Expected manual=value, got %s", req.QueryParams.Get("manual"))
	}

	if req.QueryParams.Get("name") != "merge_test" {
		t.Errorf("Expected name=merge_test, got %s", req.QueryParams.Get("name"))
	}

	fmt.Printf("Merged query params: %v\n", req.QueryParams)
}

func TestZeroValueHandling(t *testing.T) {
	client := C()

	// 测试零值处理
	type ZeroValueTest struct {
		Name         string  `query:"name"`
		Age          int     `query:"age"`             // 零值应该被包含
		Score        float64 `query:"score"`           // 零值应该被包含
		Active       bool    `query:"active"`          // false应该被包含
		EmptyWithTag string  `query:"empty,omitempty"` // 零值应该被忽略
		ZeroWithTag  int     `query:"zero,omitempty"`  // 零值应该被忽略
	}

	params := ZeroValueTest{
		Name:         "",    // 空字符串应该被包含
		Age:          0,     // 零值应该被包含
		Score:        0.0,   // 零值应该被包含
		Active:       false, // false应该被包含
		EmptyWithTag: "",    // 应该被忽略（omitempty）
		ZeroWithTag:  0,     // 应该被忽略（omitempty）
	}

	req := client.R().SetQueryParamsStruct(params)
	queryParams := req.QueryParams

	// 验证零值字段被包含
	if queryParams.Get("name") != "" {
		t.Errorf("Expected empty name to be included, got %s", queryParams.Get("name"))
	}

	if queryParams.Get("age") != "0" {
		t.Errorf("Expected age=0, got %s", queryParams.Get("age"))
	}

	if queryParams.Get("score") != "0" {
		t.Errorf("Expected score=0, got %s", queryParams.Get("score"))
	}

	if queryParams.Get("active") != "false" {
		t.Errorf("Expected active=false, got %s", queryParams.Get("active"))
	}

	// 验证omitempty字段被忽略
	if queryParams.Has("empty") {
		t.Errorf("Expected empty field with omitempty to be omitted")
	}

	if queryParams.Has("zero") {
		t.Errorf("Expected zero field with omitempty to be omitted")
	}

	fmt.Printf("Zero value test query params: %v\n", queryParams)
}
