package gohessian

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type hessianRequest struct {
	body []byte
}

//向hessian服务发请求,并将解析结果返回
//url string hessian 服务地址
//method string hessian 公开的方法
//params ...Any 请求参数
func Request(url string, method string, params ...Any) (interface{}, error) {
	r := &hessianRequest{}
	r.packHead(method)
	for _, v := range params {
		r.packParam(v)
	}
	r.packEnd()

	resp, err := httpPost(url, bytes.NewReader(r.body))
	if err != nil {
		return nil, err
	}

	h := NewHessian(bytes.NewReader(resp))
	v, err := h.Parse()

	if err != nil {
		return nil, err
	}

	return v, nil
}

//http post 请求,返回body字节数组
func httpPost(url string, body io.Reader) (rb []byte, err error) {
	var resp *http.Response
	if resp, err = http.Post(url, "application/binary", body); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf(resp.Status)
		return
	}
	defer resp.Body.Close()
	rb, err = ioutil.ReadAll(resp.Body)
	return
}

//封装  hessian 请求头
func (h *hessianRequest) packHead(method string) {
	tmp_b, _ := util.PackUint16(uint16(len(method)))
	h.body = append(h.body, []byte{99, 0, 1, 109}...)
	h.body = append(h.body, tmp_b...)
	h.body = append(h.body, []byte(method)...)
}

//封装参数
func (h *hessianRequest) packParam(p Any) {
	tmp_b, err := Encode(p)
	if err != nil {
		panic(err)
	}
	h.body = append(h.body, tmp_b...)
}

//封装包尾
func (h *hessianRequest) packEnd() {
	h.body = append(h.body, 'z')
}
