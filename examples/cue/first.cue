import "text/template"

_instances: [
  "first",
  "second",
]

configs: [ for instance in _instances {
  app:       template.Execute("myapp-{{ . }}", instance)
  env:       "production"
  namespace: "production"
  chart: {
    repository: "https://chart.onechart.dev"
    name:       "cron-job"
    version:    0.32
  }
  values: {
    image: {
      repository: "<account>.dkr.ecr.eu-west-1.amazonaws.com/myapp"
      tag:        "1.1.1"
    }
  }
}]
