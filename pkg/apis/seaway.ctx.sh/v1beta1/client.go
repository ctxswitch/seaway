package v1beta1

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type UploadResponse struct {
	// ETag is the ETag for the uploaded object.
	ETag string `json:"etag"`
	// Key is the key for the uploaded object.
	Key string `json:"key"`
	// Size is the size of the uploaded object.
	Size int64 `json:"size"`
	// Code is the HTTP status code for the upload.
	Code int `json:"code"`
	// Error is the error message if the upload failed.
	Error string `json:"error"`
}

type ClientAPI interface {
	Upload(ctx context.Context, path string, params map[string]string) error
}

type Client struct {
	url string
}

func NewClient(url string) *Client {
	return &Client{url: url}
}

func (u *Client) Upload(ctx context.Context, path string, params map[string]string) (*UploadResponse, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("archive", info.Name())
	if err != nil {
		return nil, err
	}
	part.Write(data)

	// For the params I want to capture the name, namespace, and etag so we can verify.
	for k, v := range params {
		_ = writer.WriteField(k, v)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response UploadResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, err
}
