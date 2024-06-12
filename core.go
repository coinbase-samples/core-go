package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const EmptyQueryParams = ""

type Client struct {
	HttpBaseUrl string
	Credentials *Credentials
	HttpClient  *http.Client
}

type Credentials struct {
	AccessKey   string
	Passphrase  string
	SigningKey  string
	PortfolioId string
}

type ApiRequest struct {
	Path                    string
	Query                   string
	HttpMethod              string
	Body                    []byte
	ExpectedHttpStatusCodes []int
	Client                  Client
}

type ApiResponse struct {
	Request        *ApiRequest
	Body           []byte
	HttpStatusCode int
	HttpStatusMsg  string
	Error          *ApiError
}

type ApiError struct {
	Message      string `json:"message"`
	CodeExpected []int
	CodeReceived int
	ParsedUrl    string
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("Error: %s, Expected Status Codes: %v, Received Status Code: %d, URL: %s", e.Message, e.CodeExpected, e.CodeReceived, e.ParsedUrl)
}

type HeaderFunc func(req *http.Request, path string, body []byte, client Client, t time.Time)

func Post(
	ctx context.Context,
	client Client,
	path,
	query string,
	request,
	response interface{},
	headersFunc HeaderFunc,
) error {
	return Call(ctx, client, path, query, http.MethodPost, []int{http.StatusOK}, request, response, headersFunc)
}

func Get(
	ctx context.Context,
	client Client,
	path,
	query string,
	request,
	response interface{},
	headersFunc HeaderFunc,
) error {
	return Call(ctx, client, path, query, http.MethodGet, []int{http.StatusOK}, request, response, headersFunc)
}

func Put(
	ctx context.Context,
	client Client,
	path,
	query string,
	request,
	response interface{},
	headersFunc HeaderFunc,
) error {
	return Call(ctx, client, path, query, http.MethodPut, []int{http.StatusOK}, request, response, headersFunc)
}

func Delete(
	ctx context.Context,
	client Client,
	path,
	query string,
	request,
	response interface{},
	headersFunc HeaderFunc,
) error {
	return Call(ctx, client, path, query, http.MethodDelete, []int{http.StatusOK}, request, response, headersFunc)
}

func Patch(
	ctx context.Context,
	client Client,
	path,
	query string,
	request,
	response interface{},
	headersFunc HeaderFunc,
) error {
	return Call(ctx, client, path, query, http.MethodPatch, []int{http.StatusOK}, request, response, headersFunc)
}

func Call(
	ctx context.Context,
	client Client,
	path,
	query,
	httpMethod string,
	expectedHttpStatusCodes []int,
	request,
	response interface{},
	headersFunc HeaderFunc,
) error {

	if client.Credentials == nil {
		return errors.New("credentials not set")
	}

	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	resp := MakeCall(
		ctx,
		&ApiRequest{
			Path:                    path,
			Query:                   query,
			HttpMethod:              httpMethod,
			Body:                    body,
			ExpectedHttpStatusCodes: expectedHttpStatusCodes,
			Client:                  client,
		},
		headersFunc,
	)

	if resp.Error != nil {
		return resp.Error
	}

	if err := json.Unmarshal(resp.Body, response); err != nil {
		return err
	}

	return nil
}

func MakeCall(ctx context.Context, request *ApiRequest, headersFunc HeaderFunc) *ApiResponse {

	response := &ApiResponse{
		Request: request,
	}

	callUrl := fmt.Sprintf("%s%s%s", request.Client.HttpBaseUrl, request.Path, request.Query)

	parsedUrl, err := url.Parse(callUrl)
	if err != nil {
		response.Error = &ApiError{
			Message:      fmt.Sprintf("invalid URL: %s - %v", callUrl, err),
			ParsedUrl:    callUrl,
			CodeReceived: 0,
		}
		return response
	}

	var requestBody []byte
	if request.HttpMethod == http.MethodPost || request.HttpMethod == http.MethodPut {
		requestBody = request.Body
	}

	req, err := http.NewRequestWithContext(ctx, request.HttpMethod, callUrl, bytes.NewReader(requestBody))
	if err != nil {
		response.Error = &ApiError{
			Message:      err.Error(),
			CodeReceived: 0,
		}
		return response
	}

	headersFunc(req, parsedUrl.Path, requestBody, request.Client, time.Now())

	res, err := request.Client.HttpClient.Do(req)
	if err != nil {
		response.Error = &ApiError{
			Message:      err.Error(),
			CodeReceived: 0,
		}
		return response
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		response.Error = &ApiError{
			Message:      err.Error(),
			CodeReceived: 0,
		}
		return response
	}

	response.Body = body
	response.HttpStatusCode = res.StatusCode
	response.HttpStatusMsg = res.Status

	isExpectedStatusCode := false
	for _, code := range request.ExpectedHttpStatusCodes {
		if res.StatusCode == code {
			isExpectedStatusCode = true
			break
		}
	}

	if !isExpectedStatusCode {
		var apiErr ApiError
		if jsonErr := json.Unmarshal(response.Body, &apiErr); jsonErr != nil {
			apiErr.Message = string(body)
		}

		apiErr.CodeExpected = request.ExpectedHttpStatusCodes
		apiErr.CodeReceived = res.StatusCode
		apiErr.ParsedUrl = callUrl

		response.Error = &apiErr
	}

	return response
}
