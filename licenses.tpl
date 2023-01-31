[
{{- range $i, $e := . }}
    {{- if $i }},{{ end }}
    {
        "package": "{{ $e.Name }}",
        "license": "{{ $e.LicenseName }}"
    }
{{- end }}
]