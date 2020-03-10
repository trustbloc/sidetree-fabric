/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package filehandler

// File contains the file upload request
type File struct {
	ContentType string `json:"contentType"`
	Content     []byte `json:"content"`
}
