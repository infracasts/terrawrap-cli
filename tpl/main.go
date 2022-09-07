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

package tpl

import (
	"fmt"
	"strings"
)

func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func TFStringFormatter(str string) string {
	const lineLimit = 80
	if len(str) > lineLimit {
		pos := 0
		newStr := ""

		for i := 0; i < (len(str) / lineLimit); i++ {
			startPos := pos
			if pos >= len(str) {
				break
			}

			if i != 0 {
				newStr = newStr + "\n"
			}

			searchSpace := Min(len(str)-1, pos+lineLimit)

			nextSpace := strings.Index(str[searchSpace:], " ")
			pos = searchSpace + nextSpace
			if pos >= len(str) {
				newStr = newStr + str[startPos:]
				break
			}

			newStr = newStr + str[startPos:pos]
		}

		if pos < len(str) {
			newStr = newStr + "\n" + str[pos:]
		}

		return fmt.Sprintf(`<<EOF
%s
EOF`, newStr)
	} else if strings.Contains(str, `"`) {
		return fmt.Sprintf("`%s`", str)
	} else {
		return fmt.Sprintf(`"%s"`, str)
	}
}

func ResourceTemplate() []byte {
	return []byte(`/*
{{ .TerrawrapLine }}

{{ .Copyright }}

{{ if .Legal.Header }}{{ .Legal.Header }}{{ end -}}
*/

resource "{{ .Type }}" "{{ .Name }}" {
{{- $max_argument_len := .MaxArgumentLength }}
{{- range $index, $elem := .Arguments }}
  {{- with $elem }}
    {{- if .Deprecated }}
    // {{ .Name }} = var.{{ .PrefixedName }} // DEPRECATED
	{{- else }}
    {{ .Name }} = var.{{ .PrefixedName }}
    {{- end -}}
  {{- end -}}
{{- end }}
}
`)
}

func VariablesTemplate() []byte {
	return []byte(`/*
{{ .TerrawrapLine }}

{{ .Copyright }}

{{ if .Legal.Header }}{{ .Legal.Header }}{{ end -}}
*/

{{ range $index, $elem := .Arguments }}
variable "{{ $elem.PrefixedName }}" {
  type = {{ $elem.ValueType }}
  {{ if $elem.DefaultValue }}default = {{ $elem.DefaultValue }}{{ end }}
  description = {{ $elem.Description | tfStringFormat }}
}
{{ end }}
`)
}

// TODO add prefix
func OutputTemplate() []byte {
	return []byte(`/*
{{ .TerrawrapLine }}

{{ .Copyright }}

{{- if .Legal.Header -}}{{ .Legal.Header }}{{- end }}
*/

{{- $resource_type := .Type }}
{{- $resource_name := .Name }}
{{ range $index, $elem := .Attributes }}
output "{{ $elem.PrefixedName }}" {
  value = {{ $resource_type }}.{{ $resource_name }}.{{ $elem.Name }}
  description = {{ $elem.Description | tfStringFormat }}
}
{{ end -}}
`)
}
