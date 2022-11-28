---
aliases:
  - /docs/grafana/latest/alerting/contact-points/message-templating/
  - /docs/grafana/latest/alerting/contact-points/message-templating/create-message-template/
  - /docs/grafana/latest/alerting/message-templating/
  - /docs/grafana/latest/alerting/unified-alerting/message-templating/
  - /docs/grafana/latest/alerting/contact-points/message-templating/delete-message-template/
  - /docs/grafana/latest/alerting/contact-points/message-templating/edit-message-template/
  - /docs/grafana/latest/alerting/manage-notifications/create-message-template/
  - /docs/grafana/latest/alerting/contact-points/message-templating/
  - /docs/grafana/latest/alerting/contact-points/message-templating/example-template/
  - /docs/grafana/latest/alerting/message-templating/
  - /docs/grafana/latest/alerting/unified-alerting/message-templating/
  - /docs/grafana/latest/alerting/fundamentals/contact-points/example-template/
  - /docs/grafana/latest/alerting/contact-points/message-templating/template-data/
  - /docs/grafana/latest/alerting/message-templating/template-data/
  - /docs/grafana/latest/alerting/unified-alerting/message-templating/template-data/
  - /docs/grafana/latest/alerting/fundamentals/contact-points/template-data/
keywords:
  - grafana
  - alerting
  - guide
  - contact point
  - templating
title: Templating notifications
weight: 200
---

# Templating notifications

In Grafana it is possible to customize notifications with message templates.

## Supported features when templating notifications

Message templates can be used to change the title, message, and format of the message in notifications. For example, message templates can be used to add, remove, or re-order information in the notification including the summary, description, labels and annotations, values, and links; and format text in bold and italic, and add or remove line breaks.

Message templates cannot be used to change the following:

- How images are included in notifications
- The design of notifications in instant messaging services such as Slack and Microsoft Teams
- The data in webhook notifications, including the structure of the JSON request or sending data in other formats such as XML
- Adding or removing HTTP headers in webhook notifications other than those in the contact point configuration

If looking to template labels and annotations themselves see the documentation on [Templating labels and annotations]({{< relref "../fundamentals/annotation-label/variables-label-annotation/" >}}).

## How to create, edit and delete message templates

Message templates can be found in the Contact points tab on the Alerting page. Here you can see a list of all your message templates. A message template can contain more than one template. For example, you might want to create a message template called `email` that contains all email templates.

### Create message templates

To create a message template click the New template button:

{{< figure max-width="940px" src="/static/img/docs/alerting/unified/list-message-templates-9-3.png" caption="List message templates" >}}

You will be asked to choose a name for the message template. This name is not the name of the template that you will use in notification policies, but the name of the template as it appears in the Contact points page. The templates that you will use in notification policies should be written in the content field. For example, in the screenshot below there is a template called `email.subject`:

{{< figure max-width="940px" src="/static/img/docs/alerting/unified/new-message-template-9-3.png" caption="New message template" >}}

When defining a template it is important to make sure the name is unique. For example, there should be no more than one template called `email.subject` in the same message template or across different message templates. You should also avoid defining templates with the same name as default templates such as `__subject`, `__text_values_list`, `__text_alert_list`, `default.title` and `default.message`.

To save the message template click the Save button.

### Edit message templates

Find the message template that you want to edit in the list of message templates and click the Edit icon.

{{< figure max-width="940px" src="/static/img/docs/alerting/unified/list-message-templates-with-template-9-3.png" caption="List message templates" >}}

You can edit the message template just as if you were creating a new message template. To save the message template click the Save button.

### Delete message templates

Find the message template that you want to delete in the list of message templates and click the Delete (Trash) icon:

{{< figure max-width="940px" src="/static/img/docs/alerting/unified/list-message-templates-with-template-9-3.png" caption="List message templates" >}}

You will be ask to confirm that you want to delete the template. If you do want to delete the template click Yes, otherwise click Cancel:

{{< figure max-width="940px" src="/static/img/docs/alerting/unified/delete-message-template-9-3.png" caption="Delete message template" >}}

## How to use message templates in contact points

To use message templates when sending notifications you must [execute a template]({{< relref "#executing-a-template" >}}) from within a contact point. To execute a template you use the `template` directive, followed by the name of the template, and the cursor to be passed to the template. If you are unfamiliar with how to write templates, the `template` directive, or cursors check out [How to write templates]({{< relref "#how-to-write-templates" >}}).

{{< figure max-width="940px" src="/static/img/docs/alerting/unified/use-message-template-9-3.png" caption="Use message template" >}}

## How to write templates

Message templates are written in Go's templating language, [text/template](https://pkg.go.dev/text/template), which is a little different from a number of other popular templating languages such as Jinja. In text/template, template code starts with `{{` and ends with `}}` irrespective of whether the template code prints a variable or executes control structures such as if statements. This is different from other templating languages such as Jinja where printing a variable uses `{{` and `}}` and control structures use `{%` and `%}`.

In the following topics we will look at how to write message templates in Go's templating language, and then look at a number of more complex examples that you might want to use as a reference for customizing your notifications.

### Printing variables

Unlike other templating languages which use variables, in text/template there is a special cursor called dot (written as `.`). You can think of this cursor as a variable whose value changes depending where in the template it is used. For example, at the start of all message templates the cursor, called dot, contains a number of fields including `Alerts`, `Status`, `GroupLabels`, `CommonLabels`, `CommonAnnotations` and `ExternalURL`. For example, to print all alerts you could write the following template code:

```
{{ .Alerts }}
```

### Iterating over alerts

To print just the labels for each alert; rather than the labels, annotations, and metadata for all alerts; we cannot just write a template with `{{ .Labels }}` because `Labels` does not exist at the start of our message template. Instead, we must use a `range` to iterate the alerts in `.Alerts`:

```
{{ range .Alerts }}
{{ .Labels }}
{{ end }}
```

You might have noticed that inside the `range` we can write `{{ .Labels }}` to print the labels of each alert. This works because `range .Alerts` changes the cursor to refer to the current alert in the list of alerts. When the range is finished the cursor is reset to the value it had before the start of the range:

```
{{ range .Alerts }}
{{ .Labels }}
{{ end }}
{{/* does not work, .Labels does not exist here */}}
{{ .Labels }}
{{/* works, cursor was reset */}}
{{ .Status }}
```

### Iterating over annotations and labels

Now that have a good understanding of `range` and how the cursor works let's write a template to print the labels of each alert in the format `The name of the label is $name, and the value is $value`, where `$name` and `$value` contain the name and value of each label.

Like in the previous example, we need to use a range to iterate over the alerts in `.Alerts` and change the cursor to refer to the current alert in the list. However, we then need to use a second range on the sorted labels so the cursor is updated once more to refer to each individual label pair in `.Labels.SortedPairs`. Here we can use `.Name` and `.Value` to print the name and value of each label:

```
{{ range .Alerts }}
{{ range .Labels.SortedPairs }}
The name of the label is {{ .Name }}, and the value is {{ .Value }}
{{ end }}
{{ end }}
```

It is important to understand that in the second range it is not possible to use `.Labels` as the cursor is now referring to the current label pair in the current alert:

```
{{ range .Alerts }}
{{ range .Labels.SortedPairs }}
The name of the label is {{ .Name }}, and the value is {{ .Value }}
{{/* does not work because in the second range . is a label not an alert */}}
{{ .Labels.SortedPairs }}
{{ end }}
{{ end }}
```

It is possible to get around this using variables, which we will look at next.

### Variables

While text/template has a cursor, it is still possible to define variables. However, unlike other templating languages such as Jinja where variables refer to data passed into the template, variables in text/template must be created within the template.

The following example creates a variable called `variable` with the current value of the cursor:

```
{{ $variable := . }}
```

We can use this to create a variable called `$alert` to fix the issue we had in the previous example:

```
{{ range .Alerts }}
{{ $alert := . }}
{{ range .Labels.SortedPairs }}
The name of the label is {{ .Name }}, and the value is {{ .Value }}
{{/* works because we created a variable called $alert */}}
{{ $alert.SortedPairs }}
{{ end }}
{{ end }}
```

### If statements

If statements are supported in text/template too. For example to print `There are no alerts` if there are no alerts in `.Alerts` we could write the following template code:

```
{{ if .Alerts }}
{{ range .Alerts }}
{{ .Labels }}
{{ end }}
{{ else }}
There are no alerts
{{ end }}
```

### Indentation

It is possible to use indentation to make templates more readable:

```
{{ if .Alerts }}
  {{ range .Alerts }}
    {{ .Labels }}
  {{ end }}
{{ else }}
  There are no alerts
{{ end }}
```

However, such intention is then included in the printed text. Next we will see how to remove it.

### Removing spaces and line breaks

Suppose we have the following template:

```
{{ range .Alerts }}
{{ range .Labels.SortedPairs }}
{{ .Name }} = {{ .Value }}
{{ end }}
{{ end }}
```

The output of this template might be:

```
alertname = High CPU usage
instance = server1
```

But what if we want:

```
alertname = High CPU usage instance = server1
```

In text/template we can use `{{-` and `-}}` to remove leading and trailing spaces and line breaks:

```
{{ range .Alerts -}}
{{ range .Labels.SortedPairs -}}
{{ .Name }} = {{ .Value }}
{{- end }}
{{- end }}
```

We can use `{{-` and `-}}` to remove indentation too:

```
{{ if .Alerts -}}
  {{ range .Alerts -}}
    {{ .Labels }}
  {{ end -}}
{{ else -}}
  There are no alerts
{{- end }}
```

### Comments

It is possible to add comments with `{{/*` and `*/}}`:

```
{{/* This is a comment */}}
```

### Defining a template

When writing complex template code, or template code intended to be shared across a number of different contact points, it is recommended to define it within a template. Here we are defining a template called `print_labels`:

```
{{ define "print_labels" }}
{{ range .Alerts }}
{{ range .Labels.SortedPairs }}
The name of the label is {{ .Name }}, and the value is {{ .Value }}
{{ end }}
{{ end }}
{{ end }}
```

When defining a template it is important to make sure the name is unique. It is not recommend definining templates with the same name as default templates such as `__subject`, `__text_values_list`, `__text_alert_list`, `default.title` and `default.message`.

### Executing a template

To execute a template use `template`, passing as arguments the name of the template in double quotes, and the cursor that should be passed into the template:

```
{{ template "print_labels" . }}
```

### Templates and cursors

It is possible to pass cursors into templates other than the dot cursor. For example, we can change the `print_labels` template to iterate over `.` instead of `.Alerts`:

```
{{ define "print_labels" }}
{{ range . }}
{{ range .Labels.SortedPairs }}
The name of the label is {{ .Name }}, and the value is {{ .Value }}
{{ end }}
{{ end }}
{{ end }}
```

When executing the template, instead of passing it the original dot cursor we would pass the list of alerts which as we know from previous examples is `.Alerts`:

```
{{ template "print_labels" .Alerts }}
```

The advantage of this is that we can either pass all alerts to the template, or just the firing alerts, without having the change the template:

```
{{ template "print_labels" .Alerts.Firing }}
```

To avoid comments from adding line breaks use:

```
{{- /* This is a comment with no leading or trailing line breaks */ -}}
```

## Example templates

Here are number of more complex examples that you might want to use as a reference for customizing your notifications.

### Templating the subject of an email

Here is an example of templating the subject of an email with the text **1 firing alert(s), 0 resolved alerts(s)**:

```
{{ define "email.subject" }}
{{ len .Alerts.Firing }} firing alert(s), {{ len .Alerts.Resolved }} resolved alert(s)
{{ end }}
```

Here is an example of a template where the subject shows the number of firing and resolved alerts if the number is more than 0. For example, when the notification contains just firing alerts the subject will be **1 firing alert(s)**, when the notification contains just resolved alerts the subject will be **1 resolved alert(s)**, and when the notification contains both firing and resolved alerts the subject will be **1 firing alert(s) 1 resolved alerts**:

```
{{ define "email.subject" }}
{{ if .Alerts.Firing -}}
{{ len .Alerts.Firing }} firing alert(s)
{{ end }}
{{ if .Alerts.Firing -}}
{{ len .Alerts.Resolved }} resolved alert(s)
{{ end }}
{{ end }}
```

### Templating the message of an email

Here is an example of an email with the following message:

```
There are 2 firing alert(s), and 1 resolved alert(s)

Firing alerts:
- alertname=Test 1 grafana_folder=GrafanaCloud has value(s) B=1
- alertname=Test 2 grafana_folder=GrafanaCloud has value(s) B=2

Resolved alerts:
- alertname=Test 3 grafana_folder=GrafanaCloud has value(s) B=0
```

The message is created from two templates: **email.message_alert** and **email.message**. The **email.message_alert** template is used to print the labels and values for each firing and resolved alert while the **email.message** template contains the structure of the email.

```
{{- define "email.message_alert" -}}
{{- range .Labels.SortedPairs }}{{ .Name }}={{ .Value }} {{ end }} has value(s)
{{- range $k, $v := .Values }} {{ $k }}={{ $v }}{{ end }}
{{- end -}}

{{ define "email.message" }}
There are {{ len .Alerts.Firing }} firing alert(s), and {{ len .Alerts.Resolved }} resolved alert(s)

{{ if .Alerts.Firing -}}
Firing alerts:
{{- range .Alerts.Firing }}
- {{ template "email.message_alert" . }}
{{- end }}
{{- end }}

{{ if .Alerts.Resolved -}}
Resolved alerts:
{{- range .Alerts.Resolved }}
- {{ template "email.message_alert" . }}
{{- end }}
{{- end }}

{{ end }}
```

### Templating the labels, annotations, SilenceURL and DashboardURL of all alerts

Here is an example of templating the labels, annotations, SilenceURL and DashboardURL of all alerts.

```
{{- define "custom.print_alert" -}}
[{{.Status}}] {{ .Labels.alertname }}
Labels:
{{- range .Labels.SortedPairs }}
  {{ .Name }}: {{ .Value }}
{{- end }}
{{- if .Annotations -}}
Annotations:
{{- range .Annotations.SortedPairs }}
  {{ .Name }}: {{ .Value }}
{{- end }}
{{- end }}
{{- if .SilenceURL -}}
  Silence alert: {{ .SilenceURL }}
{{- end }}
{{- if .DashboardURL -}}
  Go to dashboard: {{ .DashboardURL }}
{{- end -}}
{{- end -}}

{{ define "custom.message" }}
{{ if .Alerts.Firing -}}
{{ len .Alerts.Firing }} firing alerts:
{{ range .Alerts.Firing }}
{{ template "custom.print_alert" .}}
{{ end }}
{{ end }}
{{ if .Alerts.Resolved -}}
{{ len .Alerts.Resolved }} resolved alerts:
{{ range .Alerts.Resolved }}
{{ template "custom.print_alert" .}}
{{ end }}
{{ end }}
{{ end }}
```

## Reference

The following reference contains all the data and template functions that are available when templating notifications.

### Data

#### ExtendedData

When templating notifications, the cursor at the start of the template refers to a structure called **ExtendedData**. This structure contains a number of fields including `Alerts`, `Status`, `GroupLabels`, `CommonLabels`, `CommonAnnotations` and `ExternalURL`. The complete set of fields can be found in the following table:

| Name              | Dot notation         | Kind        | Notes                                                                                                                |
| ----------------- | -------------------- | ----------- | -------------------------------------------------------------------------------------------------------------------- |
| Receiver          | `.Receiver`          | string      | Name of the contact point that the notification is being sent to.                                                    |
| Status            | `.Status`            | string      | `firing` if at least one alert is firing, otherwise `resolved`.                                                      |
| Alerts            | `.Alerts`            | Alert       | List of alert objects that are included in this notification (see below).                                            |
| Firing alerts     | `.Alerts.Firing`     | Alert       | List of alert objects that are included in this notification (see below).                                            |
| Resolved alerts   | `.Alerts.Resolved`   | Alert       | List of alert objects that are included in this notification (see below).                                            |
| GroupLabels       | `.GroupLabels`       | Named Pairs | Labels these alerts were grouped by.                                                                                 |
| CommonLabels      | `.CommonLabels`      | Named Pairs | Labels common to all the alerts included in this notification.                                                       |
| CommonAnnotations | `.CommonAnnotations` | Named Pairs | Annotations common to all the alerts included in this notification.                                                  |
| ExternalURL       | `.ExternalURL`       | string      | Back link to the Grafana that sent the notification. If using external Alertmanager, back link to this Alertmanager. |

#### Alert

**ExtendedData** contains a list of alerts. Each alert in the list refers to a structure called **Alert**. This structure contains a number of fields including `Status`, `Labels`, `Annotations` and `Values`. The complete set of fields can be found in the following table:

| Name         | Dot notation    | Kind                                 | Notes                                                                                                                                          |
| ------------ | --------------- | ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| Status       | `.Status`       | string                               | `firing` or `resolved`.                                                                                                                        |
| Labels       | `.Labels`       | Named Pairs                          | A set of labels attached to the alert.                                                                                                         |
| Annotations  | `.Annotations`  | Named Pairs                          | A set of annotations attached to the alert.                                                                                                    |
| Values       | `.Values`       | Named Pairs                          | A set of annotations attached to the alert.                                                                                                    |
| StartsAt     | `.StartsAt`     | [Time](https://pkg.go.dev/time#Time) | Time the alert started firing.                                                                                                                 |
| EndsAt       | `.EndsAt`       | [Time](https://pkg.go.dev/time#Time) | Only set if the end time of an alert is known. Otherwise set to a configurable timeout period from the time since the last alert was received. |
| GeneratorURL | `.GeneratorURL` | string                               | A back link to Grafana or external Alertmanager.                                                                                               |
| SilenceURL   | `.SilenceURL`   | string                               | Link to grafana silence for with labels for this alert pre-filled. Only for Grafana managed alerts.                                            |
| DashboardURL | `.DashboardURL` | string                               | Link to grafana dashboard, if alert rule belongs to one. Only for Grafana managed alerts.                                                      |
| PanelURL     | `.PanelURL`     | string                               | Link to grafana dashboard panel, if alert rule belongs to one. Only for Grafana managed alerts.                                                |
| Fingerprint  | `.Fingerprint`  | string                               | Fingerprint that can be used to identify the alert.                                                                                            |
| ValueString  | `.ValueString`  | string                               | A string that contains the labels and value of each reduced expression in the alert.                                                           |

#### GroupLabels, CommonLabels and CommonAnnotations

`KeyValue` is a set of key/value string pairs that represent labels and annotations.

Here is an example containing two annotations:

```json
{
  "summary": "alert summary",
  "description": "alert description"
}
```

In addition to direct access of data (labels and annotations) stored as KeyValue, there are also methods for sorting, removing and transforming.

| Name        | Arguments | Returns                                 | Notes                                                       |
| ----------- | --------- | --------------------------------------- | ----------------------------------------------------------- |
| SortedPairs |           | Sorted list of key & value string pairs |
| Remove      | []string  | KeyValue                                | Returns a copy of the Key/Value map without the given keys. |
| Names       |           | []string                                | List of label names                                         |
| Values      |           | []string                                | List of label values                                        |

### Functions
