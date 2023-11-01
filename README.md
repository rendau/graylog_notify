# Graylog notify service

In graylog, create event-definition with notification to this service.

Add these fields (type: template) to event-definition:
```
tag: ${source.tag}
message: ${source.message}
```

Example of filter query: `m_level: /(fatal|error|warn)/`
