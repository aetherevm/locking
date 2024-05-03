// Copyright Aether Labs
// SPDX-License-Identifier:BSLv1.1(https://github.com/aetherevm/aether/blob/master/LICENSE)
package tests

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Mock the codec Codec
type MockAppCoded struct {
	codec.Codec
}

// Mock the Codec Marshal function
func (m MockAppCoded) Marshal(_ codec.ProtoMarshaler) ([]byte, error) {
	return nil, fmt.Errorf("mocked error")
}

// Mock the Codec Unmashal function
func (m MockAppCoded) Unmarshal(_ []byte, _ codec.ProtoMarshaler) error {
	return fmt.Errorf("mocked error")
}
