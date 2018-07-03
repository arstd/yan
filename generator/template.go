package generator

const tpl = `
{{- /*************** header template *****************/}}
{{define "header" -}}
// !!! DO NOT EDIT THIS FILE. It is generated by 'light' tool.
// @light: https://github.com/arstd/light
// Generated from source: {{.Source}}
package {{.Package}}
import (
		"bytes"
		"fmt"
		"github.com/arstd/light/light"
		"github.com/arstd/light/null"
		{{- if .Log }}
			"github.com/arstd/log"
		{{- end}}

		{{- range $path, $short := .Imports}}
			{{/* $short */}} "{{$path}}"
		{{- end}}
)

{{if .VarName}}
func init() { {{.VarName}} = new(Store{{.Name}}) }
{{end}}

type Store{{.Name}} struct{}
{{end}}

{{- /*************** fragment template *****************/}}
{{define "fragment" -}}
{{- if .Fragment.Condition}}
	if {{.Fragment.Condition}} {
{{- end }}
{{- if .Fragment.Statement }}
	{{- if .Fragment.Range }}
		if len({{.Fragment.Range}}) > 0 {
			{{- if .Buf}}
				fmt.Fprintf(&{{.Buf}}, "{{.Fragment.Statement}} ", strings.Repeat(",?", len({{.Fragment.Range}}))[1:])
			{{- end}}
			{{- if .Args}}
				for _, v := range {{.Fragment.Range}} {
					{{.Args}} = append({{.Args}}, v)
				}
			{{- end}}
		}
	{{- else if .Fragment.Replacers }}
		{{- if .Buf}}
			fmt.Fprintf(&{{.Buf}}, "{{.Fragment.Statement}} "{{range $elem := .Fragment.Replacers}}, {{$elem}}{{end}})
		{{- end}}
	{{- else }}
		{{- if .Buf}}
			{{.Buf}}.WriteString("{{.Fragment.Statement}} ")
		{{- end}}
	{{- end }}
	{{- if .Fragment.Variables }}
		{{- if .Args}}
			{{.Args}} = append({{.Args}}{{range $elem := .Fragment.Variables}}, {{LookupValueOfParams $.Method $elem}}{{end}})
		{{- end}}
	{{- end }}
{{- else }}
	{{- range $fragment := .Fragment.Fragments }}
		{{- template "fragment" (aggregate $.Method $fragment $.Buf $.Args)}}
	{{- end }}
{{- end }}
{{- if .Fragment.Condition}}
	}
{{- end }}
{{- end}}


{{- /*************** ddl template *****************/}}
{{define "ddl" -}}
query := buf.String()
{{if .Interface.Log -}}
	log.Debug(query)
	{{if HasVariable $ -}}
		log.Debug(args...)
	{{end -}}
{{end -}}
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
_, err := exec.ExecContext(ctx, query{{if HasVariable $ }}, args...{{end}})
{{if .Interface.Log -}}
	if err != nil {
		log.Error(query)
		{{if HasVariable $ -}}
			log.Error(args...)
		{{end -}}
		log.Error(err)
	}
{{end -}}
return err
{{end}}

{{- /*************** update/delete template *****************/}}
{{define "update" -}}
query := buf.String()
{{if .Interface.Log -}}
	log.Debug(query)
	{{if HasVariable $ -}}
		log.Debug(args...)
	{{end -}}
{{end -}}
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
res, err := exec.ExecContext(ctx, query{{if HasVariable $ }}, args...{{end}})
if err != nil {
	{{if .Interface.Log -}}
		log.Error(query)
		{{if HasVariable $ -}}
			log.Error(args...)
		{{end -}}
		log.Error(err)
	{{end -}}
	return 0, err
}
return res.RowsAffected()
{{end -}}

{{- /*************** insert template *****************/}}
{{define "insert" -}}
query := buf.String()
{{if .Interface.Log -}}
	log.Debug(query)
	{{if HasVariable $ -}}
		log.Debug(args...)
	{{end -}}
{{end -}}
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
res, err := exec.ExecContext(ctx, query{{if HasVariable $ }}, args...{{end}})
if err != nil {
	{{if .Interface.Log -}}
		log.Error(query)
		{{if HasVariable $ -}}
			log.Error(args...)
		{{end -}}
		log.Error(err)
	{{end -}}
	return 0, err
}
return res.LastInsertId()
{{end}}


{{- /*************** bulky template *****************/}}
{{define "bulky" -}}
var buf bytes.Buffer
{{- range $i, $fragment := .Statement.Fragments }}
	{{template "fragment" (aggregate $ $fragment "buf" "")}}
{{- end }}

query := buf.String()
log.Debug(query)

{{- $tx := MethodTx $ -}}
{{- if $tx}}
	{{- if eq $tx "tx"}}{{else}}
		var tx = {{$tx}}
	{{- end}}
{{- else}}
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
{{- end}}

stmt, err := tx.Prepare(query)
if err != nil {
	return 0, err
}
var args []interface{}
for _, {{ParamsLastElem .Params}} := range {{ParamsLast .Params}} {
	args = args[:0]
	{{- range $i, $fragment := .Statement.Fragments }}
		{{- template "fragment" (aggregate $ $fragment "" "args")}}
	{{- end }}
	log.Debug(args...)
	if _, err := stmt.Exec(args...); err != nil {
		return 0, err
	}
}
{{- if not $tx}}
if err := tx.Commit(); err != nil {
	return 0, err
}
{{- end}}

return int64(len({{ParamsLast .Params}})), nil
{{end}}

{{- /*************** get template *****************/}}
{{define "get" -}}
query := buf.String()
{{if .Interface.Log -}}
	log.Debug(query)
	{{if HasVariable $ -}}
		log.Debug(args...)
	{{end -}}
{{end -}}
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
row := exec.QueryRowContext(ctx, query{{if HasVariable $ }}, args...{{end}})
xu := new({{ResultElemTypeName .Results.Result}})
xdst := []interface{}{
	{{- range $i, $field := .Statement.Fields -}}
		{{- if $i -}} , {{- end -}}
		{{- LookupScanOfResults $ $field -}}
	{{- end -}}
}
err := row.Scan(xdst...)
if err != nil {
	if err == sql.ErrNoRows {
		return nil, nil
	}
	{{if .Interface.Log -}}
		log.Error(query)
		{{if HasVariable $ -}}
			log.Error(args...)
		{{end -}}
		log.Error(err)
	{{end -}}
		return nil, err
	}
{{if .Interface.Log -}}
	log.Trace(xdst)
{{end -}}
return xu, err
{{end}}

{{- /*************** list template *****************/}}
{{define "list" -}}
query := buf.String()
{{if .Interface.Log -}}
	log.Debug(query)
	{{if HasVariable $ -}}
		log.Debug(args...)
	{{end -}}
{{end -}}
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
rows, err := exec.QueryContext(ctx, query{{if HasVariable $ }}, args...{{end}})
if err != nil {
	{{if .Interface.Log -}}
		log.Error(query)
		{{if HasVariable $ -}}
			log.Error(args...)
		{{end -}}
		log.Error(err)
	{{end -}}
	return nil, err
}
defer rows.Close()
var data {{ResultTypeName .Results.Result}}
for rows.Next() {
	xu := new({{ ResultElemTypeName .Results.Result }})
	data = append(data, xu)
	xdst := []interface{}{
		{{- range $i, $field := .Statement.Fields -}}
			{{- if $i -}} , {{- end -}}
			{{- LookupScanOfResults $ $field -}}
		{{- end -}}
	}
	err = rows.Scan(xdst...)
	if err != nil {
		{{if .Interface.Log -}}
			log.Error(query)
			{{if HasVariable $ -}}
				log.Error(args...)
			{{end -}}
			log.Error(err)
		{{end -}}
		return nil, err
	}
	{{if .Interface.Log -}}
		log.Trace(xdst)
	{{end -}}
}
if err = rows.Err(); err != nil {
	{{if .Interface.Log -}}
		log.Error(query)
		{{if HasVariable $ -}}
			log.Error(args...)
		{{end -}}
		log.Error(err)
	{{end -}}
	return nil, err
}
return data, nil
{{end}}

{{- /*************** page template *****************/}}
{{define "page" -}}
var total int64
totalQuery := "SELECT count(1) "+ buf.String()
{{if .Interface.Log -}}
	log.Debug(totalQuery)
	{{if HasVariable $ -}}
		log.Debug(args...)
	{{end -}}
{{end -}}
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
err := exec.QueryRowContext(ctx, totalQuery{{if HasVariable $ }}, args...{{end}}).Scan(&total)
if err != nil {
	{{if .Interface.Log -}}
		log.Error(totalQuery)
		{{if HasVariable $ -}}
			log.Error(args...)
		{{end -}}
		log.Error(err)
	{{end -}}
	return 0, nil, err
}
{{if .Interface.Log -}}
	log.Debug(total)
{{end -}}

query := xFirstBuf.String() + buf.String() + xLastBuf.String()
args = append(xFirstArgs, args...)
args = append(args, xLastArgs...)
{{if .Interface.Log -}}
	log.Debug(query)
	{{if HasVariable $ -}}
		log.Debug(args...)
	{{end -}}
{{end -}}
ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
rows, err := exec.QueryContext(ctx, query{{if HasVariable $ }}, args...{{end}})
if err != nil {
	{{if .Interface.Log -}}
		log.Error(query)
		{{if HasVariable $ -}}
			log.Error(args...)
		{{end -}}
		log.Error(err)
	{{end -}}
	return 0, nil, err
}
defer rows.Close()
var data {{ResultTypeName .Results.Result}}
for rows.Next() {
	xu := new({{ ResultElemTypeName .Results.Result }})
	data = append(data, xu)
	xdst := []interface{}{
		{{- range $i, $field := .Statement.Fields -}}
			{{- if $i -}} , {{- end -}}
			{{- LookupScanOfResults $ $field -}}
		{{- end -}}
	}
	err = rows.Scan(xdst...)
	if err != nil {
		{{if .Interface.Log -}}
			log.Error(query)
			{{if HasVariable $ -}}
				log.Error(args...)
			{{end -}}
			log.Error(err)
		{{end -}}
		return 0, nil, err
	}
	{{if .Interface.Log -}}
		log.Trace(xdst)
	{{end -}}
}
if err = rows.Err(); err != nil {
	{{if .Interface.Log -}}
		log.Error(query)
		{{if HasVariable $ -}}
			log.Error(args...)
		{{end -}}
		log.Error(err)
	{{end -}}
	return 0, nil, err
}
return total, data, nil
{{end}}


{{- /*************** agg template *****************/}}
{{define "agg" -}}
query := buf.String()
{{if .Interface.Log -}}
	log.Debug(query)
	{{if HasVariable $ -}}
		log.Debug(args...)
	{{end -}}
{{end -}}
var xu {{ResultTypeName .Results.Result}}
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
err := exec.QueryRowContext(ctx, query{{if HasVariable $ }}, args...{{end}}).Scan({{ResultWrap .Results.Result}})
if err != nil {
	if err == sql.ErrNoRows {
		{{- if .Interface.Log}}
			log.Debug(xu)
		{{- end}}
		return xu, nil
	}
	{{if .Interface.Log -}}
		log.Error(query)
		{{if HasVariable $ -}}
			log.Error(args...)
		{{end -}}
		log.Error(err)
	{{end -}}
	return xu, err
}
{{if .Interface.Log -}}
	log.Debug(xu)
{{end -}}
return xu, nil
{{end}}

{{- /*************** main *****************/ -}}
{{template "header" . -}}
{{range $method := .Methods -}}
	func (*Store{{$.Name}}) {{$method.Signature}} {
		{{- if eq $method.Type "bulky"}}
			{{template "bulky" $method -}}
		{{- else}}
			{{- $tx := MethodTx $method -}}
			var exec = {{if $tx }} light.GetExec({{$tx}}, db) {{else}} db {{end}}
			var buf bytes.Buffer
			{{if HasVariable $method -}}
				var args []interface{}
			{{end -}}

			{{- range $i, $fragment := .Statement.Fragments }}
				{{/* if type=page, return field statement and ordery by limit statement reserved */}}
				{{$last := sub (len $method.Statement.Fragments) 1 }}
				{{if and (eq $method.Type "page") (eq $i 0) }}
					var xFirstBuf bytes.Buffer
					var xFirstArgs []interface{}
					{{- template "fragment" (aggregate $method $fragment "xFirstBuf" "xFirstArgs")}}
				{{else if and (eq $method.Type "page") (eq $i $last) }}
					var xLastBuf bytes.Buffer
					var xLastArgs []interface{}
					{{- template "fragment" (aggregate $method $fragment "xLastBuf" "xLastArgs")}}
				{{else if not (and (eq $method.Type "page") (or (eq $i 0) (eq $i $last)))}}
					{{template "fragment" (aggregate $method $fragment "buf" "args")}}
				{{end}}
			{{- end }}

			{{- if eq $method.Type "ddl"}}
				{{- template "ddl" $method}}
			{{- else if or (eq $method.Type "update") (eq $method.Type "delete")}}
				{{- template "update" $method}}
			{{- else if eq $method.Type "insert"}}
				{{- template "insert" $method}}
			{{- else if eq $method.Type "get"}}
				{{- template "get" $method}}
			{{- else if eq $method.Type "list"}}
				{{- template "list" $method}}
			{{- else if eq $method.Type "page"}}
				{{- template "page" $method}}
			{{- else if eq $method.Type "agg"}}
				{{- template "agg" $method}}
			{{- else}}
				panic("unimplemented")
			{{- end -}}
		{{- end -}}
	}

{{end}}
`
