{{/*
Common labels
*/}}
{{- define "greenscout-backend.labels" -}}
helm.sh/chart: {{ include "greenscout-backend.chart" . }}
{{ include "greenscout-backend.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "greenscout-backend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "greenscout-backend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "greenscout-backend.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "greenscout-backend.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Return the appropriate apiVersion for deployment
*/}}
{{- define "greenscout-backend.deployment.apiVersion" -}}
{{- if semverCompare ">=1.9-0" .Capabilities.KubeVersion.GitVersion -}}
apps/v1
{{- else -}}
apps/v1beta2
{{- end -}}
{{- end -}}
