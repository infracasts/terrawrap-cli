// All modifications to original contents, hereto known as the Derivative Works,
// are copyright 2022 InfraCasts, LLC, and licensed under the Apache License, Version 2.0.
//
// Copyright © 2021 Steve Francia <spf@spf13.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package terraform

import (
	"github.com/spf13/cobra-cli/cmd"
	"os"
)

// Module contains name and paths to generated modules.
type Module struct {
	TerrawrapLine string

	Copyright    string
	Legal        cmd.License
	Name         string

	AbsolutePath string
}

func (m *Module) Initialize() error {
	// check if AbsolutePath exists
	if _, err := os.Stat(m.AbsolutePath); os.IsNotExist(err) {
		// create directory
		if err := os.MkdirAll(m.AbsolutePath, 0754); err != nil {
			return err
		}
	}

	return nil
}
