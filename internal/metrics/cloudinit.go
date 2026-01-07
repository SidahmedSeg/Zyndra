package metrics

import "fmt"

// GenerateCloudInitScript generates a cloud-init script that installs and runs Node Exporter
// Returns the user_data string that can be passed to instance creation
func GenerateCloudInitScript() string {
	return `#cloud-config
packages:
  - docker.io

runcmd:
  # Install Node Exporter
  - |
    docker run -d \
      --name=node-exporter \
      --restart=always \
      --net="host" \
      --pid="host" \
      -v "/:/host:ro,rslave" \
      quay.io/prometheus/node-exporter:latest \
      --path.rootfs=/host \
      --collector.filesystem.mount-points-exclude="^/(sys|proc|dev|host|etc)($$|/)"
  
  # Optional: Install cAdvisor for container metrics
  - |
    docker run -d \
      --name=cadvisor \
      --restart=always \
      --volume=/:/rootfs:ro \
      --volume=/var/run:/var/run:ro \
      --volume=/sys:/sys:ro \
      --volume=/var/lib/docker/:/var/lib/docker:ro \
      --volume=/dev/disk/:/dev/disk:ro \
      --publish=8080:8080 \
      --privileged \
      --device=/dev/kmsg \
      gcr.io/cadvisor/cadvisor:latest

write_files:
  - path: /etc/systemd/system/node-exporter.service
    content: |
      [Unit]
      Description=Node Exporter
      After=docker.service
      Requires=docker.service
      
      [Service]
      Type=oneshot
      RemainAfterExit=yes
      ExecStart=/bin/true
      
      [Install]
      WantedBy=multi-user.target
    permissions: '0644'
`
}

// GenerateCloudInitScriptWithContainer generates cloud-init script that includes
// both Node Exporter and the application container
func GenerateCloudInitScriptWithContainer(imageURL string, envVars map[string]string, port int) string {
	envVarsStr := ""
	for key, value := range envVars {
		envVarsStr += fmt.Sprintf("      - %s=%s\n", key, value)
	}
	
	return fmt.Sprintf(`#cloud-config
packages:
  - docker.io

runcmd:
  # Install Node Exporter
  - |
    docker run -d \
      --name=node-exporter \
      --restart=always \
      --net="host" \
      --pid="host" \
      -v "/:/host:ro,rslave" \
      quay.io/prometheus/node-exporter:latest \
      --path.rootfs=/host \
      --collector.filesystem.mount-points-exclude="^/(sys|proc|dev|host|etc)($$|/)"
  
  # Install cAdvisor for container metrics
  - |
    docker run -d \
      --name=cadvisor \
      --restart=always \
      --volume=/:/rootfs:ro \
      --volume=/var/run:/var/run:ro \
      --volume=/sys:/sys:ro \
      --volume=/var/lib/docker/:/var/lib/docker:ro \
      --volume=/dev/disk/:/dev/disk:ro \
      --publish=8080:8080 \
      --privileged \
      --device=/dev/kmsg \
      gcr.io/cadvisor/cadvisor:latest
  
  # Run application container
  - |
    docker run -d \
      --name=app \
      --restart=always \
      --publish=%d:%d \
%s      %s
`, port, port, envVarsStr, imageURL)
}

