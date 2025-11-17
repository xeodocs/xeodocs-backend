# Tiltfile for XeoDocs backend dev environment

# External services
k8s_yaml('k8s/db-deployment.yaml')
k8s_yaml('k8s/db-service.yaml')
k8s_yaml('k8s/db-data-persistentvolumeclaim.yaml')

k8s_yaml('k8s/rabbitmq-deployment.yaml')
k8s_yaml('k8s/rabbitmq-service.yaml')
k8s_yaml('k8s/rabbitmq-data-persistentvolumeclaim.yaml')

k8s_yaml('k8s/kafka-deployment.yaml')
k8s_yaml('k8s/kafka-service.yaml')
k8s_yaml('k8s/kafka-data-persistentvolumeclaim.yaml')

k8s_yaml('k8s/kafka-ui-deployment.yaml')
k8s_yaml('k8s/kafka-ui-service.yaml')

k8s_yaml('k8s/influxdb-deployment.yaml')
k8s_yaml('k8s/influxdb-service.yaml')
k8s_yaml('k8s/influxdb-data-persistentvolumeclaim.yaml')

# Built services
docker_build('xeodocs/auth', '.', dockerfile='Dockerfile.auth',
  live_update=[
    sync('.', '/app'),
    run('cd /app && go build -o main ./cmd/auth', trigger=['./cmd/auth']),
  ]
)
k8s_yaml('k8s/auth-deployment.yaml')
k8s_yaml('k8s/auth-service.yaml')

docker_build('xeodocs/project', '.', dockerfile='Dockerfile.project',
  live_update=[
    sync('.', '/app'),
    run('cd /app && go build -o main ./cmd/project', trigger=['./cmd/project']),
  ]
)
k8s_yaml('k8s/project-deployment.yaml')
k8s_yaml('k8s/project-service.yaml')

docker_build('xeodocs/worker', '.', dockerfile='Dockerfile.worker',
  live_update=[
    sync('.', '/app'),
    run('cd /app && go build -o main ./cmd/worker', trigger=['./cmd/worker']),
  ]
)
k8s_yaml('k8s/worker-deployment.yaml')
k8s_yaml('k8s/worker-service.yaml')

docker_build('xeodocs/repository', '.', dockerfile='Dockerfile.repository',
  live_update=[
    sync('.', '/app'),
    run('cd /app && go build -o main ./cmd/repository', trigger=['./cmd/repository']),
  ]
)
k8s_yaml('k8s/repository-deployment.yaml')
k8s_yaml('k8s/repository-service.yaml')

docker_build('xeodocs/logging', '.', dockerfile='Dockerfile.logging',
  live_update=[
    sync('.', '/app'),
    run('cd /app && go build -o main ./cmd/logging', trigger=['./cmd/logging']),
  ]
)
k8s_yaml('k8s/logging-deployment.yaml')
k8s_yaml('k8s/logging-service.yaml')

docker_build('xeodocs/build', '.', dockerfile='Dockerfile.build',
  live_update=[
    sync('.', '/app'),
    run('cd /app && go build -o main ./cmd/build', trigger=['./cmd/build']),
  ]
)
k8s_yaml('k8s/build-deployment.yaml')
k8s_yaml('k8s/build-service.yaml')

docker_build('xeodocs/analytics', '.', dockerfile='Dockerfile.analytics',
  live_update=[
    sync('.', '/app'),
    run('cd /app && go build -o main ./cmd/analytics', trigger=['./cmd/analytics']),
  ]
)
k8s_yaml('k8s/analytics-deployment.yaml')
k8s_yaml('k8s/analytics-service.yaml')

docker_build('xeodocs/scheduler', '.', dockerfile='Dockerfile.scheduler',
  live_update=[
    sync('.', '/app'),
    run('cd /app && go build -o main ./cmd/scheduler', trigger=['./cmd/scheduler']),
  ]
)
k8s_yaml('k8s/scheduler-deployment.yaml')
k8s_yaml('k8s/scheduler-service.yaml')

docker_build('xeodocs/e2e', '.', dockerfile='Dockerfile.e2e',
  live_update=[
    sync('.', '/app'),
    run('cd /app && go test ./tests/e2e -v', trigger=['./tests/e2e']),
  ]
)
k8s_yaml('k8s/e2e-deployment.yaml')
k8s_yaml('k8s/e2e-service.yaml')

docker_build('xeodocs/gateway', '.', dockerfile='Dockerfile.gateway',
  live_update=[
    sync('.', '/app'),
    run('cd /app && go build -o main ./cmd/gateway', trigger=['./cmd/gateway']),
  ]
)
k8s_yaml('k8s/gateway-deployment.yaml')
k8s_yaml('k8s/gateway-service.yaml')

# PVC for repos
k8s_yaml('k8s/repos-persistentvolumeclaim.yaml')

# Port forwards for external access
k8s_resource('gateway', port_forwards=['12020'])
k8s_resource('kafka-ui', port_forwards=['8080'])
k8s_resource('db', port_forwards=['5432'])
k8s_resource('rabbitmq', port_forwards=['5672', '15672'])
k8s_resource('kafka', port_forwards=['9092'])
k8s_resource('influxdb', port_forwards=['8086'])
