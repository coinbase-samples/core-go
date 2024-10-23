/*
 * Copyright 2024-present Coinbase Global, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/http2"
)

const (
	EmptyQueryParams = ""

	QueryParamFirstSep = "?"

	QueryParamAdditionalSep = "&"

	appendQueryParamPattern = "%s%s%s=%s"
)

type apiRequest struct {
	Path                    string
	Query                   string
	HttpMethod              string
	Body                    []byte
	ExpectedHttpStatusCodes []int
	Client                  Client
}

type ApiResponse struct {
	Request        *apiRequest
	Body           []byte
	HttpStatusCode int
	HttpStatusMsg  string
	Error          *ApiError
}

type ApiError struct {
	Message      string `json:"message"`
	CodeExpected []int  `json:"-"`
	CodeReceived int    `json:"-"`
	ParsedUrl    string `json:"-"`
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("Unexpected response: %s, Expected Status Codes: %v, Received Status Code: %d, URL: %s", e.Message, e.CodeExpected, e.CodeReceived, e.ParsedUrl)
}

type HttpHeaderFunc func(req *http.Request, path string, body []byte, client Client, t time.Time)

func DefaultHttpClient() (http.Client, error) {

	tr := &http.Transport{
		ResponseHeaderTimeout: 5 * time.Second,
		Proxy:                 http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			DualStack: true,
			Timeout:   5 * time.Second,
		}).DialContext,
		MaxIdleConns:          10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		MaxIdleConnsPerHost:   5,
		ExpectContinueTimeout: 2 * time.Second,
	}

	if err := http2.ConfigureTransport(tr); err != nil {
		return http.Client{}, err
	}

	return http.Client{
		Transport: tr,
	}, nil
}

func HttpPost(
	ctx context.Context,
	client Client,
	path,
	query string,
	expectedHttpStatusCodes []int,
	request,
	response interface{},
	headersFunc HttpHeaderFunc,
) error {
	return call(ctx, client, path, query, http.MethodPost, expectedHttpStatusCodes, request, response, headersFunc)
}

func HttpGet(
	ctx context.Context,
	client Client,
	path,
	query string,
	expectedHttpStatusCodes []int,
	request,
	response interface{},
	headersFunc HttpHeaderFunc,
) error {
	return call(ctx, client, path, query, http.MethodGet, expectedHttpStatusCodes, request, response, headersFunc)
}

func HttpPut(
	ctx context.Context,
	client Client,
	path,
	query string,
	expectedHttpStatusCodes []int,
	request,
	response interface{},
	headersFunc HttpHeaderFunc,
) error {
	return call(ctx, client, path, query, http.MethodPut, expectedHttpStatusCodes, request, response, headersFunc)
}

func HttpDelete(
	ctx context.Context,
	client Client,
	path,
	query string,
	expectedHttpStatusCodes []int,
	request,
	response interface{},
	headersFunc HttpHeaderFunc,
) error {
	return call(ctx, client, path, query, http.MethodDelete, expectedHttpStatusCodes, request, response, headersFunc)
}

func HttpPatch(
	ctx context.Context,
	client Client,
	path,
	query string,
	expectedHttpStatusCodes []int,
	request,
	response interface{},
	headersFunc HttpHeaderFunc,
) error {
	return call(ctx, client, path, query, http.MethodPatch, expectedHttpStatusCodes, request, response, headersFunc)
}

func call(
	ctx context.Context,
	client Client,
	path,
	query,
	httpMethod string,
	expectedHttpStatusCodes []int,
	request,
	response interface{},
	headersFunc HttpHeaderFunc,
) error {

	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	resp := makeCall(
		ctx,
		&apiRequest{
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

func makeCall(ctx context.Context, request *apiRequest, headersFunc HttpHeaderFunc) *ApiResponse {

	response := &ApiResponse{
		Request: request,
	}

	callUrl := fmt.Sprintf("%s%s%s", request.Client.HttpBaseUrl(), request.Path, request.Query)

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
	if request.HttpMethod == http.MethodPost || request.HttpMethod == http.MethodPut || request.HttpMethod == http.MethodPatch {
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

	res, err := request.Client.HttpClient().Do(req)
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

func AppendHttpQueryParam(queryParams, key, value string) string {
	return fmt.Sprintf(appendQueryParamPattern, queryParams, HttpQueryParamSep(strings.Contains(queryParams, QueryParamFirstSep)), key, value)
}

func HttpQueryParamSep(appended bool) string {
	if appended {
		return QueryParamAdditionalSep
	}
	return QueryParamFirstSep
}
