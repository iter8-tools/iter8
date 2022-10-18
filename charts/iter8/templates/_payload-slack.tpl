{
  "text": "Your Iter8 report is ready: {{ regexReplaceAll "\"" (regexReplaceAll "\n" (.Report | toPrettyJson) "\\n") "\\\""}}"
}