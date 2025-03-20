{{ color .PodColor .PodName }}:{{with $d := .Message | tryParseJSON}} {{ toTimestamp $d.time "15:04:05" "Europe/Moscow" }} {{levelColor $d.level}} {{$d.message}}{{if $d.app_method}} {{colorYellow $d.app_method}}{{end}}{{if $d.user_id}} by {{$d.user_id}}{{end}}{{if $d.request_id}} {{colorWhite $d.request_id}}{{end}}{{ range $key, $value := $d}}{{if eq $key "app_method" "request_id" "level" "time" "message" "caller" "user_id"}}{{continue}}{{end}} {{colorCyan $key}}={{json $value}}{{end}} {{ else }} {{ .Message }} {{ end }}
