{{/*
Expand the name of the chart.
*/}}
{{- define "pod-image-swap-webhook.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "pod-image-swap-webhook.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "pod-image-swap-webhook.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "pod-image-swap-webhook.labels" -}}
helm.sh/chart: {{ include "pod-image-swap-webhook.chart" . }}
{{ include "pod-image-swap-webhook.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "pod-image-swap-webhook.selectorLabels" -}}
app.kubernetes.io/name: {{ include "pod-image-swap-webhook.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "pod-image-swap-webhook.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "pod-image-swap-webhook.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
The pod template used by the Deployment and DaemonSet resources 
*/}}
{{- define "pod-image-swap-webhook.podTemplate" -}}
metadata:
  {{- with .Values.podAnnotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  labels:
    # Revision forces recreation of pods on upgrade.
    # Required to refresh the server certificate and key.
    revision: {{ .Release.Revision | quote }}
    {{- include "pod-image-swap-webhook.selectorLabels" . | nindent 4 }}
  {{- with .Values.podLabels }}
    {{- toYaml . | nindent 4 }}
  {{- end }}
spec:
  {{- with .Values.imagePullSecrets }}
  imagePullSecrets:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  serviceAccountName: {{ include "pod-image-swap-webhook.serviceAccountName" . }}
  securityContext:
    {{- toYaml .Values.podSecurityContext | nindent 4 }}
  containers:
    - name: {{ .Chart.Name }}
      securityContext:
        {{- toYaml .Values.securityContext | nindent 8 }}
      image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
      imagePullPolicy: {{ .Values.image.pullPolicy }}
      env:
        - name: PISW_CONFIG_PATH
          value: /config/config.yaml
        - name: PISW_CERT_DIR
          value: /certs
      ports:
        - name: metrics
          containerPort: 8080
          protocol: TCP
        - name: healthz
          containerPort: 8081
          protocol: TCP
        - name: webhook
          containerPort: 9443
          protocol: TCP
      livenessProbe:
        httpGet:
          path: /healthz
          port: healthz
      readinessProbe:
        httpGet:
          path: /readyz
          port: healthz
      resources:
        {{- toYaml .Values.resources | nindent 8 }}
      volumeMounts:
        - name: config
          mountPath: /config/config.yaml
          subPath: config.yaml
        - name: certs
          mountPath: /certs
  volumes:
    - configMap:
        name: {{ include "pod-image-swap-webhook.fullname" . }}
      name: config
    - secret:
        secretName: {{ include "pod-image-swap-webhook.fullname" . }}
      name: certs
  {{- with .Values.nodeSelector }}
  nodeSelector:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.affinity }}
  affinity:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.tolerations }}
  tolerations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end }}
